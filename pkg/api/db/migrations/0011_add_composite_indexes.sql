-- +migrate Up

create index if not exists instance_application_instance_id_application_id_idx on instance_application (instance_id,application_id);

create index if not exists activity_class_created_ts_severity_version_group_id_applica_idx on activity (class,created_ts,severity,"version",group_id,application_id);

create index if not exists instance_application_application_id_group_id_last_check_for_idx on instance_application(application_id,group_id,last_check_for_updates,instance_id);

-- +migrate Down

drop index if exists instance_application_instance_id_application_id_idx;

drop index if exists activity_class_created_ts_severity_version_group_id_applica_idx;

drop index if exists instance_application_application_id_group_id_last_check_for_idx;
