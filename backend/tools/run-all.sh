#!/bin/bash

# This script sets up a database and runs the local nebraska
# instance. The database is torn down after nebraska quits.
#
# To use a different port for the database (7890 in the example):
#
# ./run-all.sh -P 7890
#
#
# Long names of the parameters are available too:
#
# -P is --port

set -euo pipefail

function fail() {
    echo "${@}" >&2
    exit 1
}

# Parsing the arguments manually instead of relying on getopt so that the
# script also works on macOS, which ships the BSD getopt that does not
# support GNU long options (--longoptions).

port=''

while [[ "${#}" -gt 0 ]]; do
    case "${1}" in
        '-P'|'--port')
            [[ "${#}" -ge 2 ]] || fail "Missing value for ${1}"
            port="${2}"
            shift 2
            ;;
        '--port='*)
            port="${1#--port=}"
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

tools_dir="$(dirname "${0}")"
cid_file=$(mktemp)

# This is used as a bare variable (without putting into double quotes)
# to enable the use of DOCKER_CMD="sudo docker".
DOCKER_CMD="${DOCKER_CMD:-docker}"

cleanup_stage=0
function cleanup() {
    if [[ ${cleanup_stage} -ge 1 ]]; then
        ${DOCKER_CMD} kill "$(cat "${cid_file}")"
        ${DOCKER_CMD} rm "$(cat "${cid_file}")"
    fi
    rm -f "${cid_file}"
}
trap cleanup EXIT

sld_opts=(
    --id-file="${cid_file}"
    --db-name=nebraska
)

if [[ -n "${port}" ]]; then
    sld_opts+=(
        --port="${port}"
    )
fi

"${tools_dir}/setup_local_db.sh" "${sld_opts[@]}"

cleanup_stage=1

if [[ -n "${port}" ]]; then
    export NEBRASKA_DB_URL="postgres://postgres@127.0.0.1:${port}/nebraska?sslmode=disable&connect_timeout=10"
fi

"${tools_dir}/run-local-nebraska.sh"
