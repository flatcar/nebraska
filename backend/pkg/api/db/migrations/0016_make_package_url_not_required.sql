-- +migrate Up

alter table package alter column url drop not null;

-- +migrate Down

alter table package alter column url set not null;
