-- +migrate Up

alter table team alter column name type varchar(100);

-- +migrate Down

alter table team alter column name type varchar(25);
