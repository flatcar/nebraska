-- +migrate Up

alter table package alter column "url" type varchar(60000);

-- +migrate Down

alter table package alter column "url" type varchar(256);