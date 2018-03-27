#!/bin/sh

set -ex

mkdir -p "/run/postgresql"
chown -R postgres:postgres "/run/postgresql"

if [ ! -f "$PGDATA/postgresql.conf" ]; then
    mkdir -p "$PGDATA"
    chown -R postgres:postgres "$PGDATA"

    gosu postgres initdb
    # listen on all addresses
    sed -ri "s|^#(listen_addresses\s*=\s*)\S+|\1'*'|" "$PGDATA"/postgresql.conf

    # setup ssl: https://www.postgresql.org/docs/10/static/ssl-tcp.html
    if [ -f "$PGTLS/server.crt" ]; then
      sed -ri "s|^#(ssl\s*=\s*)\S+|\1on|" "$PGDATA"/postgresql.conf
      sed -ri "s|^#(ssl_cert_file\s*=\s*)\S+|\1'$PGTLS/server.crt'|" "$PGDATA"/postgresql.conf
      sed -ri "s|^#(ssl_key_file\s*=\s*)\S+|\1'$PGTLS/server.key'|" "$PGDATA"/postgresql.conf

      # only allow clients when clientcert is enabled and provided:
      # https://www.postgresql.org/docs/10/static/auth-pg-hba-conf.html
      if [ -f "$PGTLS/ca.crt" ]; then
        sed -ri "s|^#(ssl_ca_file\s*=\s*)\S+|\1'$PGTLS/ca.crt'|" "$PGDATA"/postgresql.conf

        echo "hostssl all all 0.0.0.0/0 cert clientcert=1" >> "$PGDATA/pg_hba.conf"
      else
        echo "no trusted certificate authority given, cannot enable client certs" >&2
      fi
    else
      echo "unable to enable ssl, no server.crt file found" >&2
    fi

    gosu postgres pg_ctl -w start

    gosu postgres psql -c "CREATE DATABASE coreroller;"
    gosu postgres psql -c "ALTER DATABASE coreroller SET TIMEZONE = 'UTC';"

    gosu postgres psql -c "CREATE DATABASE coreroller_tests;"
    gosu postgres psql -c "ALTER DATABASE coreroller_tests SET TIMEZONE = 'UTC';"

    gosu postgres pg_ctl -m fast -w stop
fi

exec gosu postgres postgres 
