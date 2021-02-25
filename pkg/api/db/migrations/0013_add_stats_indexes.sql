-- +migrate Up

create index instance_application_group_id_last_check_for_updates_instan_idx on instance_application(group_id,last_check_for_updates,instance_id);

create index instance_status_history_status_created_ts_idx on instance_status_history(status,created_ts);

-- +migrate Down

drop index instance_application_group_id_last_check_for_updates_instan_idx;

drop index instance_status_history_status_created_ts_idx;


