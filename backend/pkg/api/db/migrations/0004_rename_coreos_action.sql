-- +migrate Up

alter table coreos_action rename to flatcar_action;

-- +migrate Down

alter table flatcar_action rename to coreos_action;
