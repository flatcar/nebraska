#!/bin/bash

# This script runs a local instance of nebraska. It expects that both
# frontend and backend are already built, and the database in expected
# is already running. To override the database location nebraska is
# using, specify the NEBRASKA_DB_URL environment variable.
#
# Whatever parameters passed to the script are forwarded to nebraska.

set -euo pipefail

tools_dir="$(dirname "${0}")"
binary="${tools_dir}/../bin/nebraska"
static_dir="${tools_dir}/../../frontend/dist"

"${binary}" \
    -auth-mode noop \
    -http-log \
    -http-static-dir "${static_dir}" \
    "${@}"
