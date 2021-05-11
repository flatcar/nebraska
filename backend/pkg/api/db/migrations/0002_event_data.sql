-- +migrate Up

-- Initial data

-- Event types
-- Add initial event type values if the table is empty
-- +migrate StatementBegin
DO LANGUAGE plpgsql $$
begin
    if not exists (select id from event_type limit 1) then
        insert into event_type (type, result, description) values
        (3, 0, 'Instance reported an error during an update step.'),
        (3, 1, 'Updater has processed and applied package.'),
        (3, 2, 'Instances upgraded to current channel version.'),
        (13, 1, 'Downloading latest version.'),
        (14, 1, 'Update package arrived successfully.'),
        (800, 1, 'Install success. Update completion prevented by instance.');
    end if;
end;
$$;
-- +migrate StatementEnd

-- +migrate Down
