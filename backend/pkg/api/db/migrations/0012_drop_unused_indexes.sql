-- +migrate Up

drop index if exists users_team_id_idx;

drop index if exists instance_status_history_application_id_idx;

drop index if exists event_event_type_id_idx;

drop index if exists event_application_id_idx;

drop index if exists channel_package_id_idx;

-- +migrate Down

create index if not exists users_team_id_idx on users (team_id);

create index if not exists instance_status_history_application_id_idx on instance_status_history (application_id);

create index if not exists event_event_type_id_idx on event (event_type_id);

create index if not exists event_application_id_idx on event (application_id);

create index if not exists channel_package_id_idx on channel (package_id);