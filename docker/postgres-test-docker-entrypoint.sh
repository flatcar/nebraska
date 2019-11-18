#!/bin/sh

set -ex

mkdir -p "/run/postgresql"
chown -R postgres:postgres "/run/postgresql"

if [ ! -f "$PGDATA/postgresql.conf" ]; then
    mkdir -p "$PGDATA"
    chown -R postgres:postgres "$PGDATA"

    su-exec postgres initdb
    # listen on all addresses
    sed -ri "s|^#(listen_addresses\s*=\s*)\S+|\1'*'|" "$PGDATA"/postgresql.conf
    echo "host all all 172.17.0.1/32 trust" >> "$PGDATA/pg_hba.conf"

    su-exec postgres pg_ctl -w start

    su-exec postgres psql -c "CREATE DATABASE nebraska_tests;"
    su-exec postgres psql -c "ALTER DATABASE nebraska_tests SET TIMEZONE = 'UTC';"

    su-exec postgres pg_ctl -m fast -w stop
fi

touch "${NEBTMP}/postgres_init_done"

exec su-exec postgres postgres
