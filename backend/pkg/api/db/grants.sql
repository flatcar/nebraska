-- Privileges for the Nebraska serving roles described in RFC #1375.
--
-- The role names carry the database, because role identity is cluster-global
-- while these privileges are per-database; sharing one role across two Nebraska
-- databases in a cluster would bridge them. dbroles.go derives the names and
-- passes them in, so they are bounded to 63 bytes in exactly one place.
--
-- Applied after every migration run and idempotent: re-running it only
-- re-asserts the same grants, so a role created later is picked up on the next
-- start. When neither logical role exists, the whole block is a no-op and no
-- table ACL is touched.
--
-- Every relation is named rather than granted through "all tables in schema
-- public": that form fails outright when the migrations role holds nothing on
-- some unrelated table sharing the schema, which would leave the server unable
-- to start.
--
-- Adding a table? Put it in one of the three arrays here, and classify it in
-- dbroles_test.go. TestServingRoleGrants covers this file; TestTableClassification
-- covers the test's own lists.

do $$
declare
    -- Written by pkg/api/admin. Replicated control -> edge in a distributed
    -- deployment.
    admin_tables text[] := array[
        'admin_activity',
        'application',
        'channel',
        'channel_package_floors',
        'flatcar_action',
        'groups',
        'package',
        'package_channel_blacklist',
        'package_file',
        'team',
        'users'
    ];

    -- Written by pkg/api/runtime. Local to each node, never replicated.
    runtime_tables text[] := array[
        'activity',
        'event',
        'group_local',
        'instance',
        'instance_application',
        'instance_stats',
        'instance_status_history'
    ];

    -- Both roles read everything: an edge node resolves updates from the
    -- replicated admin tables. The tail is written by migrations alone, or by
    -- nothing at all.
    readable_tables text[] := admin_tables || runtime_tables || array[
        'all_activity',
        'database_migrations',
        'event_type',
        'instance_status'
    ];

    target record;
    tbl        text;
    seq        text;
    admin_role   text := current_setting('nebraska.admin_role');
    runtime_role text := current_setting('nebraska.runtime_role');
begin
    for target in
        select * from (values
            (runtime_role, runtime_tables),
            -- The admin role writes the runtime tables too: a control node also
            -- serves Omaha for its own region.
            (admin_role, admin_tables || runtime_tables)
        ) as v(role_name, tables)
    loop
        continue when not exists (
            select 1 from pg_roles where rolname = target.role_name);

        execute format('grant usage on schema public to %I', target.role_name);

        -- grant only warns when the grantor holds a privilege without grant
        -- option, so verify rather than trust it.
        foreach tbl in array readable_tables loop
            execute format('grant select on table public.%I to %I', tbl, target.role_name);

            if not has_table_privilege(target.role_name, format('public.%I', tbl)::regclass, 'select') then
                raise exception 'granting select on public.% to % had no effect; the role running migrations must own the nebraska tables',
                    tbl, target.role_name;
            end if;
        end loop;

        foreach tbl in array target.tables loop
            execute format('grant insert, update, delete on table public.%I to %I',
                           tbl, target.role_name);

            if not has_table_privilege(target.role_name, format('public.%I', tbl)::regclass, 'insert') then
                raise exception 'granting write on public.% to % had no effect; the role running migrations must own the nebraska tables',
                    tbl, target.role_name;
            end if;

            for seq in
                select pg_get_serial_sequence(format('public.%I', tbl), attname)
                from pg_attribute
                where attrelid = format('public.%I', tbl)::regclass
                  and attnum > 0
                  and not attisdropped
            loop
                continue when seq is null;
                execute format('grant usage, select on sequence %s to %I',
                               seq, target.role_name);
            end loop;
        end loop;
    end loop;
end $$;
