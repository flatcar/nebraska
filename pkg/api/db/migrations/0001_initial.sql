-- +migrate Up

-- CoreRoller schema

alter database coreroller set timezone = 'utc';

create extension if not exists "uuid-ossp";

create table team (
	id uuid primary key default uuid_generate_v4(),
	name varchar(25) not null check (name <> '') unique,
	created_ts timestamptz default current_timestamp not null
);

create table users (
	id uuid primary key default uuid_generate_v4(),
	username varchar(25) not null check (username <> '') unique,
	secret varchar(50) not null check (secret <> ''),
	created_ts timestamptz default current_timestamp not null,
	team_id uuid not null references team (id) on delete cascade
);

create index on users (team_id);

create table application (
	id uuid primary key default uuid_generate_v4(),
	name varchar(50) not null check (name <> ''),
	description text,
	created_ts timestamptz default current_timestamp not null,
	team_id uuid not null references team (id) on delete cascade,
	unique (team_id, name)
);

create table package (
	id uuid primary key default uuid_generate_v4(),
	type int not null check (type > 0),
	version varchar(255) not null check (version <> ''),
	url varchar(256) not null check (url <> ''),
	filename varchar(100),
	description text,
	size varchar(20),
	hash varchar(64),
	created_ts timestamptz default current_timestamp not null,
	application_id uuid not null references application (id) on delete cascade,
	unique(application_id, version)
);

create table coreos_action (
	id uuid primary key default uuid_generate_v4(),
	event varchar(20) default 'postinstall', 
	chromeos_version varchar(255) default '', 
	sha256 varchar(64),
	needs_admin boolean default false,
	is_delta boolean default false,
	disable_payload_backoff boolean default true,
	metadata_signature_rsa varchar(256) default '',
	metadata_size varchar(100) default '',
	deadline varchar(100) default '',
	created_ts timestamptz default current_timestamp not null,
	package_id uuid not null references package (id) on delete cascade
);

create index on coreos_action (package_id);

create table channel (
	id uuid primary key default uuid_generate_v4(),
	name varchar(25) not null check (name <> ''),
	color varchar(25) not null,
	created_ts timestamptz default current_timestamp not null,
	application_id uuid not null references application (id) on delete cascade,
	package_id uuid references package (id) on delete set null,
	unique (application_id, name)
);

create index on channel (package_id);

create table groups (
	id uuid primary key default uuid_generate_v4(),
	name varchar(50) not null check (name <> ''),
	description text not null,
	rollout_in_progress boolean default false not null,
	policy_updates_enabled boolean default true not null,
	policy_safe_mode boolean default true not null,
	policy_office_hours boolean default false not null,
	policy_timezone varchar(40),
	policy_period_interval varchar(20) not null check (policy_period_interval <> ''),
	policy_max_updates_per_period integer not null check (policy_max_updates_per_period > 0),
	policy_update_timeout varchar(20) not null check (policy_update_timeout <> ''),
	created_ts timestamptz default current_timestamp not null,
	application_id uuid not null references application (id) on delete cascade,
	channel_id uuid references channel (id) on delete set null,
	unique (application_id, name)
);

create index on groups (channel_id);

create table instance (
	id varchar(50) primary key check (id <> ''),
	ip inet not null,
	created_ts timestamptz default current_timestamp not null
);

create table instance_status (
	id serial primary key,
	name varchar(20) not null,
	color varchar(25) not null,
	icon varchar(20) not null
);

create table instance_application (
	version varchar(255) not null check (version <> ''),
	created_ts timestamptz default current_timestamp not null,
	status integer,
	last_check_for_updates timestamptz default current_timestamp not null,
	last_update_granted_ts timestamptz,
	last_update_version varchar(255) check (last_update_version <> ''),
	update_in_progress boolean default false not null,
	instance_id varchar(50) not null references instance (id) on delete cascade,
	application_id uuid not null references application (id) on delete cascade,
	group_id uuid references groups (id) on delete set null,
	primary key (instance_id, application_id)
);

create index on instance_application (instance_id);
create index on instance_application (application_id);
create index on instance_application (group_id);

create table instance_status_history (
	id serial primary key,
	status integer,
	version varchar(255) check (version <> ''),
	created_ts timestamptz default current_timestamp not null,
	instance_id varchar(50) not null references instance (id) on delete cascade,
	application_id uuid not null references application (id) on delete cascade,
	group_id uuid references groups (id) on delete cascade
);

create index on instance_status_history (instance_id);
create index on instance_status_history (application_id);
create index on instance_status_history (group_id);

create table event_type (
	id serial primary key,
	type integer not null,
	result integer not null,
	description varchar(100) not null
);

create table event (
	id serial primary key,
	created_ts timestamptz default current_timestamp not null,
	previous_version varchar(255),
	error_code varchar(100),
	instance_id varchar(50) not null references instance (id) on delete cascade,
	application_id uuid not null references application (id) on delete cascade,
	event_type_id integer not null references event_type (id) 
);

create index on event (instance_id);
create index on event (application_id);
create index on event (event_type_id);

create table activity (
	id serial primary key,
	created_ts timestamptz default current_timestamp not null,
	class integer not null,
	severity integer not null,
	version varchar(255) not null check (version <> ''),
	application_id uuid not null references application (id) on delete cascade,
	group_id uuid references groups (id) on delete cascade,
	channel_id uuid references channel (id) on delete cascade,
	instance_id varchar(50) references instance (id) on delete cascade
);

create index on activity (application_id);
create index on activity (group_id);
create index on activity (channel_id);
create index on activity (instance_id);

create table package_channel_blacklist (
	package_id uuid not null references package (id) on delete cascade,
	channel_id uuid not null references channel (id) on delete cascade,
	primary key (package_id, channel_id)
);

-- +migrate Down

drop table if exists team cascade;
drop table if exists users cascade;
drop table if exists application cascade;
drop table if exists package cascade;
drop table if exists coreos_action cascade;
drop table if exists channel cascade;
drop table if exists groups cascade;
drop table if exists instance cascade;
drop table if exists instance_status cascade;
drop table if exists instance_application cascade;
drop table if exists instance_status_history cascade;
drop table if exists event_type cascade;
drop table if exists event cascade;
drop table if exists activity cascade;
drop table if exists package_channel_blacklist cascade;
