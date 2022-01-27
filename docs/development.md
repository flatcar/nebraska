---
title: Development
weight: 20
---

This section shows how to set up a development environment for Nebraska.
Be also sure to read the [contributing](./contributing) section for the
general guidelines when contributing to Nebraska.

Nebraska needs a running `PostgresSQL` database. Apart from that, the
recommended way to work on it, is to run the backend with the `noop`
authentication backend, and the frontend in a "watch" mode (where it
auto-refreshes when there are changes to the frontend).

# Preparing the Database

Nebraska uses the `PostgreSQL` database, and expects the used
database to be set to the UTC timezone.

By default Nebraska uses a database with the name `nebraska`, and
`nebraska_tests` for the tests, respectively. For the main database, the full
URL (with a different database name if desired) can be overridden by the
environment variable `NEBRASKA_DB_URL`.

For a quick setup of `PostgreSQL` for Nebraska's development, you can use
the `postgres` container as follows:

- Start `Postgres`:
    - `docker run --rm -d --name nebraska-postgres-dev -p 5432:5432 -e POSTGRES_PASSWORD=nebraska postgres`

- Create the database for Nebraska (by default it is `nebraska`):
    - `psql postgres://postgres:nebraska@localhost:5432/postgres -c 'create database nebraska;'`

- Set the timezone to Nebraska's database:
    - `psql postgres://postgres:nebraska@localhost:5432/nebraska -c 'set timezone = "utc";'`

- Set up the nebraska_tests database for running unit tests

```bash
psql postgres://postgres:nebraska@localhost:5432/postgres -c 'create database nebraska_tests;'
psql postgres://postgres:nebraska@localhost:5432/nebraska_tests -c 'set timezone = "utc";'
```

# Development Quickstart

- Go to the Nebraska project directory and run `make`

- Run the backend (with the `noop` authentication): `make run-backend`

- Then, in a different terminal tab/window, run the frontend: `make run-frontend`

Any changes to the backend means that the `make run-backend` command should be
run again. Changes to the frontend should be automatically re-built and the
opened browser page should automatically refresh.

# Development Concepts

## Frontend

The [frontend](https://github.com/kinvolk/nebraska/tree/main/frontend) side of Nebraska is a web application built using `React` and `Material-UI`.

## Backend

The Nebraska backend is written in Go. The backend source code is structured as follows:

- **`pkg/api`**: provides functionality to do CRUD operations on all elements found in Nebraska (applications, groups, channels, packages, etc), abstracting the rest of the components from the underlying datastore (PostgreSQL). It also controls the groups' roll-out policy logic and the instances/events registration.

- **`pkg/omaha`**: provides functionality to validate, handle, process and reply to Omaha updates and events requests received from the Omaha clients. It relies on the `api` package to get update packages, store events, or register instances when needed.

- **`pkg/syncer`**: provides some functionality to synchronize packages available in the official Flatcar Container Linux channels, storing the references to them in your Nebraska datastore and even downloading packages payloads when configured to do so. It's basically in charge of keeping up to date your the Flatcar Container Linux application in your Nebraska installation.

- **`cmd/nebraska`**: is the main backend process, exposing the functionality described above in the different packages through its http server. It provides several http endpoints used to drive most of the functionality of the dashboard as well as handling the Omaha updates and events requests received from your servers and applications.

- **`cmd/initdb`**: is just a helper to reset your database, and causing the migrations to be re-run. `nebraska` will apply all database migrations automatically, so this process should only be used to wipe out all your data and start from a clean state (you should probably never need it).


### Backend Testing

Most unit tests like beside the code for example inside `backend/pkg/`.
Tests which depend on a database mostly live in `backend/pkg/api`.
Some server integration tests are separate, and live in `backend/test/api/`.

#### Environment variables

Here are some test related environment variables.

- NEBRASKA_SKIP_TESTS, if set do not run slow tests like DB using tests.
- NEBRASKA_RUN_SERVER_TESTS, if set run the integration tests.
- NEBRASKA_DB_URL, database connection URL.
- NEBRASKA_TEST_SERVER_URL, where the test server is running "http://localhost:8000"

#### Test make targets.

There are a number of make targets setup to run different tests.

##### make ci

Run all the tests that are run on CI with github actions.

##### make code-checks

Just build it, run quick tests, and lint it.

Does not run tests that require a testing db, or a test server.

##### make check

Run tests except for the integration tests by default. Requires a test database to be running.

You can use `NEBRASKA_RUN_SERVER_TESTS=1 make check` to also test the server integration tests, 
but you need to be running a server (with `make run-backend`).

##### make check-backend-with-container

Run tests inside a container, including the integration tests.

It starts it's own test server and test database.

##### make check-code-coverage

Like make check, but it outputs test coverage information.

