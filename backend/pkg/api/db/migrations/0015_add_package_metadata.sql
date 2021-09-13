-- +migrate Up

alter table package add column metadata_type varchar(250) default null;
alter table package add column metadata_content text default null;

-- +migrate Down

alter table package drop column metadata_type;
alter table package drop column metadata_content;
