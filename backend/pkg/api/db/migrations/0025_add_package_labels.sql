-- +migrate Up

alter table package add column labels jsonb not null default '{}'::jsonb
    check (jsonb_typeof(labels) = 'object'
           and not jsonb_path_exists(labels, '$.* ? (@.type() != "string")'));

-- +migrate Down

alter table package drop column labels;
