-- +migrate Up

create index instance_application_instance_id_application_id_idx on instance_application (instance_id,application_id);

create index activity_class_created_ts_severity_version_group_id_applica_idx on activity (class,created_ts,severity,"version",group_id,application_id);

-- +migrate Down

drop index instance_application_instance_id_application_id_idx;

drop index activity_class_created_ts_severity_version_group_id_applica_idx;