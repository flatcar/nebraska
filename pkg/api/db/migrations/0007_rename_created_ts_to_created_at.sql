-- +migrate Up

alter table team rename column created_ts to created_at;
alter table users rename column created_ts to created_at;
alter table application rename column created_ts to created_at;
alter table package rename column created_ts to created_at;
alter table flatcar_action rename column created_ts to created_at;
alter table channel rename column created_ts to created_at;
alter table groups rename column created_ts to created_at;
alter table instance rename column created_ts to created_at;
alter table instance_application rename column created_ts to created_at;
alter table instance_status_history rename column created_ts to created_at;
alter table event rename column created_ts to created_at;
alter table activity rename column created_ts to created_at;

-- +migrate Down

alter table team rename column created_at to created_ts;
alter table users rename column created_at to created_ts;
alter table application rename column created_at to created_ts;
alter table package rename column created_at to created_ts;
alter table flatcar_action rename column created_at to created_ts;
alter table channel rename column created_at to created_ts;
alter table groups rename column created_at to created_ts;
alter table instance rename column created_at to created_ts;
alter table instance_application rename column created_at to created_ts;
alter table instance_status_history rename column created_at to created_ts;
alter table event rename column created_at to created_ts;
alter table activity rename column created_at to created_ts;
