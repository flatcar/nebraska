# Development on macOS

Nebraska's development tooling targets Linux, but the project can be developed
on macOS as well. This guide documents the platform-specific steps.

## Prerequisites

- [Go](https://go.dev/dl/) (the version required by `backend/go.mod`)
- [Node.js](https://nodejs.org/) (the version required by `frontend/package.json`)
- [Docker](https://docs.docker.com/get-docker/) with a running daemon (only
  needed for running the PostgreSQL test database)
- A Bourne-again shell (bash). macOS ships bash 3.2 at `/bin/bash`.

## Building the backend

```sh
make -C backend build
```

This produces `backend/bin/nebraska`. On macOS the Makefile drops the
`-extldflags "-static"` linker flag automatically, because the system linker
(`clang`) does not support static linking (the Linux container image keeps the
flag).

## Running the whole stack

```sh
make run
```

This builds and starts the backend (on port 8000) and the frontend dev server
(on port 3000, proxying `/api` to the backend) in parallel.

## Local PostgreSQL database

The backend and its tests expect a PostgreSQL database. The helper scripts in
`backend/tools/` spin one up in Docker:

```sh
./backend/tools/setup_local_db.sh -f cid -n nebraska
```

The scripts previously relied on GNU `getopt` long options, which the BSD
`getopt` shipped with macOS does not support. The argument parsing is now
self-contained, so both short (`-f cid`) and long (`--id-file=cid`) forms work
on macOS without installing GNU tools.

To tear the database down:

```sh
docker kill "$(cat cid)"
docker rm "$(cat cid)"
rm -f cid
```

## Running the tests

Unit tests without a database (skipping the DB-backed tests):

```sh
NEBRASKA_SKIP_TESTS=1 make -C backend check-code-coverage
```

`make check` runs the full test suite, which starts the test PostgreSQL
container through Docker Compose, so it needs a working Docker daemon.

## macOS specifics

- **bash 3.2 vs. bash 4+:** `backend/tools/check_pkg_test.sh` no longer relies
  on `mapfile` (bash 4+), so it runs under the bash 3.2 shipped with macOS.
- **GNU vs. BSD utilities:** the helper scripts in `backend/tools/` parse their
  arguments portably and no longer require GNU `getopt` or GNU coreutils.
- If you prefer, you can install a newer bash and GNU tools with Homebrew
  (`brew install bash coreutils gnu-getopt`), but it is not required.
