-- +migrate Up

create index if not exists instance_status_history_group_id_status_created_ts_idx on instance_status_history(group_id,status,created_ts);

create index if not exists instance_status_history_instance_id_status_created_ts_idx on instance_status_history(instance_id,status,created_ts);

create index if not exists instance_application_instance_id_version_group_id_last_chec_idx on instance_application(instance_id,version,group_id,last_check_for_updates);

-- +migrate Down

drop index if exists instance_status_history_group_id_status_created_ts_idx;

drop index if exists instance_status_history_instance_id_status_created_ts_idx;

drop index if exists instance_application_instance_id_version_group_id_last_chec_idx;

