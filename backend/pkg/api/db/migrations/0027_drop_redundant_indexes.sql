-- +migrate Up

-- Each is the primary key again, or a prefix of a wider index on the same table.

drop index if exists instance_application_instance_id_application_id_idx;

drop index if exists instance_application_instance_id_idx;

drop index if exists instance_application_group_id_idx;

drop index if exists instance_status_history_instance_id_idx;

drop index if exists instance_status_history_group_id_idx;

drop index if exists idx_channel_package_floors_channel;

-- +migrate Down

create index if not exists idx_channel_package_floors_channel on channel_package_floors (channel_id);

create index if not exists instance_status_history_group_id_idx on instance_status_history (group_id);

create index if not exists instance_status_history_instance_id_idx on instance_status_history (instance_id);

create index if not exists instance_application_group_id_idx on instance_application (group_id);

create index if not exists instance_application_instance_id_idx on instance_application (instance_id);

create index if not exists instance_application_instance_id_application_id_idx on instance_application (instance_id, application_id);
