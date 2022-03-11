-- +migrate Up

create table if not exists package_file (
	id serial primary key,
	package_id uuid not null references package (id) on delete cascade,
	name varchar(250) not null check (name <> ''),
	hash varchar(250),
	size varchar(100),
	created_ts timestamptz default current_timestamp not null
);

-- +migrate Down

drop table if exists package_file;
