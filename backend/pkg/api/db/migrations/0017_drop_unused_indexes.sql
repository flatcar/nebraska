-- +migrate Up

drop index if exists instance_application_application_id_group_id_last_check_for_idx;

-- +migrate Down

create index if not exists instance_application_application_id_group_id_last_check_for_idx on instance_application(application_id,group_id,last_check_for_updates,instance_id);
