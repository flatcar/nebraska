-- +migrate Up

-- Only add a team if it does not exist yet.

-- +migrate StatementBegin

DO LANGUAGE plpgsql $$
begin
    if not exists (select id from team limit 1) then
        insert into team (id, name) values ('d89342dc-9214-441d-a4af-bdd837a3b239', 'default');
    end if;
end;
$$;

-- +migrate StatementEnd

-- +migrate Down
