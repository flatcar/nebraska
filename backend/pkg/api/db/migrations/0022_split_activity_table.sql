-- +migrate Up

-- Convert activity.id from serial to uuid so admin_activity and activity can
-- share a UNION ALL view without id-space collisions, and so each node in a
-- future distributed deployment can generate ids without coordination.
-- This rewrites the table; existing numeric ids are replaced with fresh uuids.
alter table activity drop constraint activity_pkey;
alter table activity alter column id drop default;
alter table activity alter column id type uuid using uuid_generate_v4();
alter table activity alter column id set default uuid_generate_v4();
alter table activity add primary key (id);
drop sequence if exists activity_id_seq;

create table admin_activity (
	id uuid primary key default uuid_generate_v4(),
	created_ts timestamptz default current_timestamp not null,
	class integer not null,
	severity integer not null,
	version varchar(255) not null check (version <> ''),
	application_id uuid not null references application (id) on delete cascade,
	group_id uuid references groups (id) on delete cascade,
	channel_id uuid references channel (id) on delete cascade
);

create index on admin_activity (application_id);
create index on admin_activity (group_id);
create index on admin_activity (channel_id);

-- Move existing admin-originated (channel package updated) rows to the new
-- table. Ids carry over from the just-converted uuid column above.
insert into admin_activity (id, created_ts, class, severity, version,
                            application_id, group_id, channel_id)
select id, created_ts, class, severity, version,
       application_id, group_id, channel_id
from activity where class = 6;

delete from activity where class = 6;

drop index if exists activity_channel_id_idx;
alter table activity drop column channel_id;

create view all_activity as
	select id, created_ts, class, severity, version,
	       application_id, group_id,
	       null::uuid as channel_id,
	       instance_id
	from activity
	union all
	select id, created_ts, class, severity, version,
	       application_id, group_id, channel_id,
	       null::varchar(50) as instance_id
	from admin_activity;

-- +migrate Down
-- The new id column ends up last (PG cannot reorder columns in place).

drop view if exists all_activity;

alter table activity add column channel_id uuid references channel (id) on delete cascade;
create index on activity (channel_id);

insert into activity (created_ts, class, severity, version,
                      application_id, group_id, channel_id)
select created_ts, class, severity, version,
       application_id, group_id, channel_id
from admin_activity;

drop table if exists admin_activity;

alter table activity drop constraint activity_pkey;
alter table activity drop column id;
alter table activity add column id serial primary key;
