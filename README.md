# Nebraska

Nebraska is an update manager for [Flatcar Linux](https://www.flatcar-linux.org/).

## Overview

Nebraska offers an easy way to monitor and manage the rollout of updates to applications that use
the [Omaha](https://code.google.com/p/omaha/) protocol, with special functionality for Flatcar Linux updates.

## Features

- Monitor and control application updates;
- Optimized for serving updates to Flatcar Linux;
- Automatically fetch new Flatcar Linux updates;
- Store and serve Flatcar Linux payloads (optional);
- Compatible with any applications that use the Omaha protocol;
- Define groups, channels, and packages;
- Control what updates are rolled out for which instance groups, as well as when and how they are updates;
- Pause/resume updates at any time;
- Statistics about the versions installed for instances, status history, and updates progress, etc.;
- Activity timeline to quickly see important events or errors;

## Screenshots

<table>
    <tr>
        <td width="33%"><img src="https://github.com/kinvolk/nebraska/raw/screenshots/screenshots/main.png"></td>
        <td width="33%"><img src="https://github.com/kinvolk/nebraska/raw/screenshots/screenshots/flatcar_app.png"></td>
    </tr>
    <tr>
        <td width="33%"><img src="https://github.com/kinvolk/nebraska/raw/screenshots/screenshots/group_details.png"></td>
        <td width="33%"><img src="https://github.com/kinvolk/nebraska/raw/screenshots/screenshots/instance_details.png"></td>
    </tr>
</table>

## Issues

Please report any issues in [here](https://github.com/kinvolk/nebraska/issues).

## Managing Flatcar updates

Once you have Nebraska up and running, a common use-case is to manage Flatcar Linux updates.

By default, your Flatcar Linux instances use the public servers to get updates, so you have to point them to your Nebraska deployment for it to
manage their updates. The process for doing so is slightly different depending on whether you have existing machines or new ones.

### New machines

For new machines, you can set up the updates server in their cloud config. Here is a small example of how to do it:

	coreos:
		update:
			group: stable
			server: http://your.nebraska.host:port/v1/update/

In addition to the default `stable`, `beta` and `alpha` groups, you can also create and use **custom groups** for greater control over the updates. In that case, you **must** use the group id (not the name) you will find next to the group name in the dashboard.

	coreos:
		update:
			group: ab51a000-02dc-4fc7-a6b0-c42881c89856
			server: http://your.nebraska.host:port/v1/update/

**Note**: The sample Nebraska containers provided use the `port 8000` by default (**plain HTTP, no SSL**). Please adjust the update URL setup in your servers to match your Nebraska deployment.

### Existing machines

To update the update server in existing instances please edit `/etc/flatcar/update.conf` and update the `SERVER` value (and optionally `GROUP` if needed):

	SERVER=https://your.nebraska.host/v1/update/

Again, when using custom groups instead of the official ones (stable, beta, alpha) the group id **must** be used, not the group name:

    GROUP=ab51a000-02dc-4fc7-a6b0-c42881c89856

To apply these changes run:

	sudo systemctl restart update-engine

In may take a few minutes to see an update request coming through. If you want to see it sooner, you can force it running this command:

	update_engine_client -update

### Flatcar Linux packages in Nebraska

Nebraska is able to periodically poll the public Flatcar Linux update servers and create new packages to update the corresponding channels. So if Nebraska is connected to the internet, new packages will show up automatically for the official Flatcar Linux. This functionality is optional, and turned off by default. If you
prefer to use it, you should pass the option `-enable-syncer=true` when running Nebraska.

Notice that by default Nebraska only stores metadata about the Flatcar Linux updates, not the updates payload. This means that the updates served to your instances contain instructions to download the packages payload from the public Flatcar Linux update servers directly, so your servers need access to the Internet to download them.

It is also possible to host the Flatcar Linux packages payload in Nebraska. In this case, in addition to get the packages metadata, Nebraska will also download the package payload itself so that it can serve it to your instances when serving updates.

This functionality is turned off by default. So to make Nebraska host the Flatcar Linux packages payload, the following options have to be passed to it:

    nebraska -host-flatcar-packages=true -flatcar-packages-path=/PATH/TO/STORE/PACKAGES -nebraska-url=http://your.Nebraska.host:port

## Managing updates for your own applications

In addition to managing updates for Flatcar Linux, you can use Nebraska for other applications as well.

In the `updaters/lib` directory there are some sample helpers that can be useful to create your own updaters that talk to Nebraska or even embed them into your own applications.

In the `updaters/examples` you'll find a sample minimal application built using [grace](https://github.com/facebookgo/grace) that is able to update itself using Nebraska in a graceful way.

## Contributing

Nebraska is an Open Source project and contributions are welcome. It is usually a good idea to discuss new features or major changes before submitting any code. For doing that, please file a new [issue](https://github.com/Nebraska/Nebraska/issues).

To build the whole project (backend + frontend), you can run:

    make

Below you will find some introductory notes that should point you in the right direction to start playing with the Nebraska source code.

### Backend

The Nebraska backend is written in Go. The backend source code is structured as follows:

- **`pkg/api`**: provides functionality to do CRUD operations on all elements found in Nebraska (applications, groups, channels, packages, etc), abstracting the rest of the components from the underlying datastore (PostgreSQL). It also controls the groups' roll-out policy logic and the instances/events registration.

- **`pkg/omaha`**: provides functionality to validate, handle, process and reply to Omaha updates and events requests received from the Omaha clients. It relies on the `api` package to get update packages, store events, or register instances when needed.

- **`pkg/syncer`**: provides some functionality to synchronize packages available in the official Flatcar Linux channels, storing the references to them in your Nebraska datastore and even downloading packages payloads when configured to do so. It's basically in charge of keeping up to date your the Flatcar Linux application in your Nebraska installation.

- **`cmd/nebraska`**: is the main backend process, exposing the functionality described above in the different packages through its http server. It provides several http endpoints used to drive most of the functionality of the dashboard as well as handling the Omaha updates and events requests received from your servers and applications.

- **`cmd/initdb`**: is just a helper to reset your database, and causing the migrations to be re-run. `nebraska` will apply all database migrations automatically, so this process should only be used to wipe out all your data and start from a clean state (you should probably never need it).

To build the backend you can run:

    make backend

#### Backend database

Nebraska uses `PostgreSQL`. You can install it locally or use the docker image available in the quay.io registry (`quay.io/flatcar/nebraska-postgres`). If you don't use the docker image provided, you'll have to set up the database yourself. By default Nebraska uses a database with the name `nebraska`, and `nebraska_tests`
for the tests, respectively. For the main database, the full URL (with a different database name if desired) can be overridden by the environment variable
`NEBRASKA_DB_URL`.

Note: the timezone for the Nebraska database is expected to be UTC (`set timezone = 'utc';`).

### Frontend

The frontend side of Nebraska (dashboard) is a web application built using `react` and `material-ui`.

To build the webapp you have to install `node.js`. After that, you can build it using:

    make frontend

It is very useful to be able to quickly build the frontend while changing it during development, and for that you can use the following make target:

    make frontend-watch

### User Access

For instructions on how to set up user access, please check the [authorization documentation](./docs/authorization.md).

## License

Nebraska is released under the terms of the [AGPL v3](https://www.gnu.org/licenses/agpl-3.0.en.html), and was forked from the [CoreRoller](https://github.com/coreroller/coreroller) project (licensed under [Apache 2.0](http://www.apache.org/licenses/LICENSE-2.0)).
