-- +migrate Up

-- Foreign key actions are origin-only triggers, so a delete replicated to an
-- edge node leaves the node-local rows that pointed at the deleted row (see
-- #1632). These triggers repeat those actions for the node-local tables.

-- +migrate StatementBegin
create function delete_local_rows_for_application() returns trigger as $$
begin
    delete from public.activity                where application_id = OLD.id;
    delete from public.event                   where application_id = OLD.id;
    delete from public.instance_application    where application_id = OLD.id;
    delete from public.instance_status_history where application_id = OLD.id;
    return OLD;
end;
$$ language plpgsql;
-- +migrate StatementEnd

-- +migrate StatementBegin
create function delete_local_rows_for_group() returns trigger as $$
begin
    delete from public.group_local             where group_id = OLD.id;
    delete from public.activity                where group_id = OLD.id;
    delete from public.instance_status_history where group_id = OLD.id;
    update public.instance_application set group_id = null where group_id = OLD.id;
    return OLD;
end;
$$ language plpgsql;
-- +migrate StatementEnd

-- Replication applies changes row by row and never fires statement triggers.
create trigger application_delete_local_rows
    after delete on application
    for each row execute function delete_local_rows_for_application();

create trigger groups_delete_local_rows
    after delete on groups
    for each row execute function delete_local_rows_for_group();

-- REPLICA fires only in the apply worker; everywhere else the foreign keys do
-- this already.
alter table application enable replica trigger application_delete_local_rows;
alter table groups enable replica trigger groups_delete_local_rows;

-- +migrate Down

drop trigger if exists groups_delete_local_rows on groups;
drop trigger if exists application_delete_local_rows on application;
drop function if exists delete_local_rows_for_group();
drop function if exists delete_local_rows_for_application();
