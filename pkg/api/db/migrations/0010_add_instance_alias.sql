-- +migrate Up

alter table instance add column alias varchar(256) default '';

-- +migrate Down

alter table instance drop column alias;
