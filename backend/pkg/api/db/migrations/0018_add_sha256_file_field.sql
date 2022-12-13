-- +migrate Up

alter table package_file add column if not exists hash256 varchar(64) default '';

-- +migrate Down

alter table package_file drop column if exists hash256;
