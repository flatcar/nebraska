-- +migrate Up

-- group_local holds the node-local fields for each group, kept writable on
-- every node so the runtime can update them even when groups is replicated
-- read-only on edge nodes (see RFC #1375). The policy_*_override columns
-- mirror the matching groups columns and are nullable; effective value is
-- COALESCE(override, default).
create table group_local (
    group_id                               uuid    primary key references groups(id) on delete cascade,
    rollout_in_progress                    boolean not null default false,
    policy_updates_enabled_override        boolean,
    policy_safe_mode_override              boolean,
    policy_office_hours_override           boolean,
    policy_timezone_override               varchar(40),
    policy_period_interval_override        varchar(20),
    policy_max_updates_per_period_override integer,
    policy_update_timeout_override         varchar(20),
    created_ts                             timestamptz not null default current_timestamp
);

insert into group_local (group_id, rollout_in_progress, created_ts)
select id, rollout_in_progress, created_ts from groups;

alter table groups drop column rollout_in_progress;

-- ENABLE ALWAYS fires the trigger on edge nodes during the initial
-- replication COPY too.
-- +migrate StatementBegin
create function create_group_local_for_group() returns trigger as $$
begin
    insert into public.group_local (group_id) values (NEW.id)
        on conflict (group_id) do nothing;
    return NEW;
end;
$$ language plpgsql;
-- +migrate StatementEnd

create trigger groups_create_group_local
    after insert on groups
    for each row execute function create_group_local_for_group();

alter table groups enable always trigger groups_create_group_local;

-- +migrate Down

alter table groups add column rollout_in_progress boolean not null default false;

-- Merge any local overrides back into the admin columns so the effective
-- value at the moment of rollback is preserved.
update groups set
    rollout_in_progress           = gl.rollout_in_progress,
    policy_updates_enabled        = coalesce(gl.policy_updates_enabled_override,        groups.policy_updates_enabled),
    policy_safe_mode              = coalesce(gl.policy_safe_mode_override,              groups.policy_safe_mode),
    policy_office_hours           = coalesce(gl.policy_office_hours_override,           groups.policy_office_hours),
    policy_timezone               = coalesce(gl.policy_timezone_override,               groups.policy_timezone),
    policy_period_interval        = coalesce(gl.policy_period_interval_override,        groups.policy_period_interval),
    policy_max_updates_per_period = coalesce(gl.policy_max_updates_per_period_override, groups.policy_max_updates_per_period),
    policy_update_timeout         = coalesce(gl.policy_update_timeout_override,         groups.policy_update_timeout)
    from group_local gl where groups.id = gl.group_id;

drop trigger if exists groups_create_group_local on groups;
drop function if exists create_group_local_for_group();
drop table if exists group_local;
