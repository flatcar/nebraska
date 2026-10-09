#!/usr/bin/env bash

# This script checks if all the packages in the running directory or
# below support skipping the tests if the NEBRASKA_SKIP_TESTS is
# defined in the environment. The check itself is quite lame (just a
# grep for the name), so it can be easily circumvented.

set -euo pipefail

# Get all Go packages and their test files to inspect below.
pkg_test_entries=()
while IFS= read -r line; do
    pkg_test_entries+=("${line}")
done < <(go list -f '{{.Dir}} - {{.TestGoFiles}} {{.XTestGoFiles}}' ./...)

if [[ ${#pkg_test_entries[@]} -eq 0 ]]; then
    exit 0
fi

first_dir=$(echo "${pkg_test_entries[0]}" | cut -d' ' -f1)
root=$(go list -f '{{.Root}}' "${first_dir}")
status=0

for line in "${pkg_test_entries[@]}"; do
    dir=$(echo "${line}" | cut -d' ' -f1)
    files=$(echo "${line}" | cut -d' ' -f3-)
    files="${files#'['}"
    files="${files%']'}"
    files="${files/] [/ }"
    if [[ $files == " " ]]; then
        continue
    fi
    pushd "${dir}" >/dev/null
    if ! grep -q 'NEBRASKA_SKIP_TESTS' ${files}; then
        status=1
        echo "tests in package ${dir#${root}/} do not check the NEBRASKA_SKIP_TESTS env var"
    fi
    popd >/dev/null
done

exit ${status}
