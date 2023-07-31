-- +migrate Up

create table if not exists instance_stats (
    timestamp timestamptz not null,
    channel_name varchar(25) not null,
    arch varchar(7) not null,
    version varchar(255) not null,
    instances int not null check (instances >= 0),
    unique(timestamp, channel_name, arch, version)
);

-- +migrate Down

drop table if exists instance_stats;
