-- +migrate Up

create index if not exists instance_application_group_id_last_update_granted_ts_idx on instance_application (group_id, last_update_granted_ts);

create index if not exists instance_application_group_id_version_last_update_version_idx on instance_application (group_id, version, last_update_version);

create index if not exists instance_application_group_id_update_in_progress_idx on instance_application (group_id, last_update_granted_ts) where update_in_progress;

-- +migrate Down

drop index if exists instance_application_group_id_update_in_progress_idx;

drop index if exists instance_application_group_id_version_last_update_version_idx;

drop index if exists instance_application_group_id_last_update_granted_ts_idx;
