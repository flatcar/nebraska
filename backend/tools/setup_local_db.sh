#!/bin/bash

# This script sets a up local postgres docker container with a
# database. The first parameter to the script should be a file name
# where an ID of the postgres docker container will be saved. The
# second parameter should be a name of the database in the container.
#
#
# To run a local db for checking the frontend, run:
#
# ./tools/setup_local_db.sh -f cid -n nebraska
#
#
# To run a local db for API tests, run:
#
# ./tools/setup_local_db.sh -f cid -n nebraska_tests -p nebraska
#
#
# It's possible to specify a port under which the database is
# available (7890 in the example):
#
# ./tools/setup_local_db.sh -f cid -n nebraska_tests -p nebraska -P 7890
#
#
# Long names of the parameters are available too:
#
# -p is --password
# -f is --id-file
# -n is --db-name
# -P is --port
# -d is --pg-version
#
# After being finished with the database, do:
#
# docker kill "$(cat cid)"
# docker rm "$(cat cid)"
# rm -f cid


set -euo pipefail

function fail() {
    echo "${@}" >&2
    exit 1
}

# Parsing the arguments manually instead of relying on getopt so that the
# script also works on macOS, which ships the BSD getopt that does not
# support GNU long options (--longoptions).

id_file=''
db_name=''
password=''
port=''
pg_version='latest'

while [[ "${#}" -gt 0 ]]; do
    case "${1}" in
        '-f'|'--id-file')
            [[ "${#}" -ge 2 ]] || fail "Missing value for ${1}"
            id_file="${2}"
            shift 2
            ;;
        '--id-file='*)
            id_file="${1#--id-file=}"
            shift
            ;;
        '-n'|'--db-name')
            [[ "${#}" -ge 2 ]] || fail "Missing value for ${1}"
            db_name="${2}"
            shift 2
            ;;
        '--db-name='*)
            db_name="${1#--db-name=}"
            shift
            ;;
        '-p'|'--password')
            [[ "${#}" -ge 2 ]] || fail "Missing value for ${1}"
            password="${2}"
            shift 2
            ;;
        '--password='*)
            password="${1#--password=}"
            shift
            ;;
        '-P'|'--port')
            [[ "${#}" -ge 2 ]] || fail "Missing value for ${1}"
            port="${2}"
            shift 2
            ;;
        '--port='*)
            port="${1#--port=}"
            shift
            ;;
        '-d'|'--pg-version')
            [[ "${#}" -ge 2 ]] || fail "Missing value for ${1}"
            pg_version="${2}"
            shift 2
            ;;
        '--pg-version='*)
            pg_version="${1#--pg-version=}"
            shift
            ;;
        '--')
            shift
            break
            ;;
        *)
            fail "Unrecognized argument: ${1}"
            ;;
    esac
done

if [[ "${#}" -ne 0 ]]; then
    fail "Leftover unrecognized arguments: ${*}"
fi

if [[ -z "${id_file}" ]]; then
    fail 'No container ID file name passed, use -f or --id-file'
fi
if [[ -z "${db_name}" ]]; then
    fail 'No database name passed, use -n or --db-name'
fi
if [[ -z "${port}" ]]; then
    port=5432
fi
if [[ -z "${pg_version}" ]]; then
    pg_version="latest"
fi

# This is used as a bare variable (without putting into double quotes)
# to enable the use of DOCKER_CMD="sudo docker".
DOCKER_CMD="${DOCKER_CMD:-docker}"

cleanup_stage=0
function cleanup() {
    if [[ ${cleanup_stage} -ge 2 ]]; then
        ${DOCKER_CMD} kill "$(cat "${id_file}")"
        ${DOCKER_CMD} rm "$(cat "${id_file}")"
    fi
    if [[ ${cleanup_stage} -ge 1 ]]; then
        rm -f "${id_file}"
    fi
}
trap cleanup ERR

run_opts=(
    --detach
    --publish "127.0.0.1:${port}:5432"
)

if [[ -n "${password}" ]]; then
    run_opts+=(
        -e POSTGRES_PASSWORD="${password}"
    )
else
    run_opts+=(
        -e POSTGRES_HOST_AUTH_METHOD=trust
    )
fi

run_opts+=(
    postgres:"${pg_version}"
)

cleanup_stage=1
${DOCKER_CMD} run "${run_opts[@]}" >"${id_file}"

cleanup_stage=2
cid="$(cat "${id_file}")"
until ${DOCKER_CMD} exec "${cid}" pg_isready -h localhost
do
    sleep 3
done
${DOCKER_CMD} exec "${cid}" \
       psql -h localhost -U postgres -c "create database ${db_name};"
${DOCKER_CMD} exec "${cid}" \
       psql -h localhost -U postgres -d "${db_name}" -c 'set timezone = "utc";'
