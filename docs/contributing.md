---
title: Contributing
weight: 10
---

Nebraska is an Open Source project and contributions are welcome. It is usually a good idea to discuss new features or major changes before submitting any code. For doing that, please file a new [issue](https://github.com/Nebraska/Nebraska/issues).

To build the whole project (backend + frontend), you can run:

    make

Below you will find some introductory notes that should point you in the right direction to start playing with the Nebraska source code.

## Backend

The Nebraska backend is written in Go. The backend source code is structured as follows:

- **`pkg/api`**: provides functionality to do CRUD operations on all elements found in Nebraska (applications, groups, channels, packages, etc), abstracting the rest of the components from the underlying datastore (PostgreSQL). It also controls the groups' roll-out policy logic and the instances/events registration.

- **`pkg/omaha`**: provides functionality to validate, handle, process and reply to Omaha updates and events requests received from the Omaha clients. It relies on the `api` package to get update packages, store events, or register instances when needed.

- **`pkg/syncer`**: provides some functionality to synchronize packages available in the official Flatcar Container Linux channels, storing the references to them in your Nebraska datastore and even downloading packages payloads when configured to do so. It's basically in charge of keeping up to date your the Flatcar Container Linux application in your Nebraska installation.

- **`cmd/nebraska`**: is the main backend process, exposing the functionality described above in the different packages through its http server. It provides several http endpoints used to drive most of the functionality of the dashboard as well as handling the Omaha updates and events requests received from your servers and applications.

- **`cmd/initdb`**: is just a helper to reset your database, and causing the migrations to be re-run. `nebraska` will apply all database migrations automatically, so this process should only be used to wipe out all your data and start from a clean state (you should probably never need it).

To build the backend you can run:

    make backend

### Backend database

Nebraska uses `PostgreSQL`. You can install it locally or use the docker image available in the quay.io registry (`quay.io/flatcar/nebraska-postgres`). If you don't use the docker image provided, you'll have to set up the database yourself. By default Nebraska uses a database with the name `nebraska`, and `nebraska_tests`
for the tests, respectively. For the main database, the full URL (with a different database name if desired) can be overridden by the environment variable
`NEBRASKA_DB_URL`.

Note: the timezone for the Nebraska database is expected to be UTC (`set timezone = 'utc';`).

## Frontend

The frontend side of Nebraska (dashboard) is a web application built using `react` and `material-ui`.

To build the webapp you have to install `node.js`. After that, you can build it using:

    make frontend

It is very useful to be able to quickly build the frontend while changing it during development, and for that you can use the following make target:

    make frontend-watch
