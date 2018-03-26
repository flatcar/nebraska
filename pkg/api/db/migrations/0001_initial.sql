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

-- Initial data

-- Default team
insert into team (id, name) values ('d89342dc-9214-441d-a4af-bdd837a3b239', 'default');

-- Event types
insert into event_type (type, result, description) values (3, 0, 'Instance reported an error during an update step.');
insert into event_type (type, result, description) values (3, 1, 'Updater has processed and applied package.');
insert into event_type (type, result, description) values (3, 2, 'Instances upgraded to current channel version.');
insert into event_type (type, result, description) values (13, 1, 'Downloading latest version.');
insert into event_type (type, result, description) values (14, 1, 'Update package arrived successfully.');
insert into event_type (type, result, description) values (800, 1, 'Install success. Update completion prevented by instance.');

-- CoreOS application
insert into application (id, name, description, team_id) values ('e96281a6-d1af-4bde-9a0a-97b76e56dc57', 'CoreOS', 'Linux for massive server deployments', 'd89342dc-9214-441d-a4af-bdd837a3b239');
insert into package values ('2ba4c984-5e9b-411e-b7c3-b3eb14f7a261', 1, '766.3.0', 'https://commondatastorage.googleapis.com/update-storage.core-os.net/amd64-usr/766.3.0/', 'update.gz', NULL, '154967458', 'l4Kw7AeBLrVID9JbfyMoJeB5yKg=', '2015-09-20 00:12:37.523938', 'e96281a6-d1af-4bde-9a0a-97b76e56dc57');
insert into package values ('337b3f7e-ff29-47e8-a052-f0834d25bdb5', 1, '766.4.0', 'https://commondatastorage.googleapis.com/update-storage.core-os.net/amd64-usr/766.4.0/', 'update.gz', NULL, '155018912', 'frkka+B/zTv7OPWgidY+k4SnDSg=', '2015-09-20 06:15:29.108266', 'e96281a6-d1af-4bde-9a0a-97b76e56dc57');
insert into package values ('c2a36312-b989-403e-ab57-06c055a7eac2', 1, '808.0.0', 'https://commondatastorage.googleapis.com/update-storage.core-os.net/amd64-usr/808.0.0/', 'update.gz', NULL, '177717414', 'bq3fQRHP8xB3RFUjCdAf3wQYC2E=', '2015-09-20 00:09:06.839989', 'e96281a6-d1af-4bde-9a0a-97b76e56dc57');
insert into package values ('43580892-cad8-468a-a0bb-eb9d0e09eca4', 1, '815.0.0', 'https://commondatastorage.googleapis.com/update-storage.core-os.net/amd64-usr/815.0.0/', 'update.gz', NULL, '178643579', 'kN4amoKYVZUG2WoSdQH1PHPzr5A=', '2015-09-25 13:55:20.741419', 'e96281a6-d1af-4bde-9a0a-97b76e56dc57');
insert into package values ('284d295b-518f-4d67-999e-94968d0eed90', 1, '829.0.0', 'https://commondatastorage.googleapis.com/update-storage.core-os.net/amd64-usr/829.0.0/', 'update.gz', NULL, '186245514', '2lhoUvvnoY359pi2FnaS/xsgtig=', '2015-10-10 23:11:10.825985', 'e96281a6-d1af-4bde-9a0a-97b76e56dc57');
insert into channel values ('e06064ad-4414-4904-9a6e-fd465593d1b2', 'stable', '#14b9d6', '2015-09-19 05:09:34.261241', 'e96281a6-d1af-4bde-9a0a-97b76e56dc57', '337b3f7e-ff29-47e8-a052-f0834d25bdb5');
insert into channel values ('128b8c29-5058-4643-8e67-a1a0e3c641c9', 'beta', '#fc7f33', '2015-09-19 05:09:34.264334', 'e96281a6-d1af-4bde-9a0a-97b76e56dc57', '337b3f7e-ff29-47e8-a052-f0834d25bdb5');
insert into channel values ('a87a03ad-4984-47a1-8dc4-3507bae91ee1', 'alpha', '#1fbb86', '2015-09-19 05:09:34.265754', 'e96281a6-d1af-4bde-9a0a-97b76e56dc57', '284d295b-518f-4d67-999e-94968d0eed90');
insert into groups values ('9a2deb70-37be-4026-853f-bfdd6b347bbe', 'Stable', 'For production clusters', false, true, true, false, 'Australia/Sydney', '15 minutes', 2, '60 minutes', '2015-09-19 05:09:34.269062', 'e96281a6-d1af-4bde-9a0a-97b76e56dc57', 'e06064ad-4414-4904-9a6e-fd465593d1b2');
insert into groups values ('3fe10490-dd73-4b49-b72a-28ac19acfcdc', 'Beta', 'Promoted alpha releases, to catch bugs specific to your configuration', true, true, true, false, 'Australia/Sydney', '15 minutes', 2, '60 minutes', '2015-09-19 05:09:34.273244', 'e96281a6-d1af-4bde-9a0a-97b76e56dc57', '128b8c29-5058-4643-8e67-a1a0e3c641c9');
insert into groups values ('5b810680-e36a-4879-b98a-4f989e80b899', 'Alpha', 'Tracks current development work and is released frequently', false, true, true, false, 'Australia/Sydney', '15 minutes', 1, '30 minutes', '2015-09-19 05:09:34.274911', 'e96281a6-d1af-4bde-9a0a-97b76e56dc57', 'a87a03ad-4984-47a1-8dc4-3507bae91ee1');
insert into coreos_action values ('b2b16e2e-57f8-4775-827f-8f0b11ae9bd2', 'postinstall', '', 'k8CB8tMe0M8DyZ5RZwzDLyTdkHjO/YgfKVn2RgUMokc=', false, false, true, '', '', '', '2015-09-20 00:12:37.532281', '2ba4c984-5e9b-411e-b7c3-b3eb14f7a261');
insert into coreos_action values ('d5a2cbf3-b810-4e8c-88e8-6df91fc264c6', 'postinstall', '', 'QUGnmP51hp7zy+++o5fBIwElInTAms7/njnkxutn/QI=', false, false, true, '', '', '', '2015-09-20 06:15:29.11685', '337b3f7e-ff29-47e8-a052-f0834d25bdb5');
insert into coreos_action values ('299c54d1-3344-4ae9-8ad2-5c63d56d6c14', 'postinstall', '', 'SCv89GYzx7Ix+TljqbNsd7on65ooWqBzcCrLFL4wChQ=', false, false, true, '', '', '', '2015-09-20 00:09:06.927461', 'c2a36312-b989-403e-ab57-06c055a7eac2');
insert into coreos_action values ('748df5fc-12a5-4dad-a71e-465cc1668048', 'postinstall', '', '9HUs4whizfyvb4mgl+WaNaW3VLQYwsW1GHNHJNpcFg4=', false, false, true, '', '', '', '2015-09-25 13:55:20.825242', '43580892-cad8-468a-a0bb-eb9d0e09eca4');
insert into coreos_action values ('9cd474c5-efa3-4989-9992-58ddb852ed84', 'postinstall', '', '1S9zQCLGjmefYnE/aFcpCjL1NsguHhQGj0UCm5f0M98=', false, false, true, '', '', '', '2015-10-10 23:11:10.913778', '284d295b-518f-4d67-999e-94968d0eed90');

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
