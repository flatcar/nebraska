# Nebraska  <img align="right" width=384 src="docs/nebraska-logo.svg">

Nebraska is an update manager for [Flatcar Container Linux](https://www.flatcar.org/).

## Overview

Nebraska offers an easy way to monitor and manage the rollout of updates to applications that use
the [Omaha](https://code.google.com/p/omaha/) protocol, with special functionality for Flatcar Container Linux updates.

## Features

- Monitor and control application updates;
- Optimized for serving updates to Flatcar Container Linux;
- Automatically fetch new Flatcar Container Linux updates;
- Store and serve Flatcar Container Linux payloads (optional);
- Compatible with any applications that use the Omaha protocol;
- Define groups, channels, and packages;
- Control what updates are rolled out for which instance groups, as well as when and how they are updates;
- Pause/resume updates at any time;
- Statistics about the versions installed for instances, status history, and updates progress, etc.;
- Activity timeline to quickly see important events or errors;

## Published Container Images

**UPDATE:** New container images are now only published under `ghcr.io/flatcar/nebraska`.

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

Please report any issues in [here](https://github.com/flatcar/nebraska/issues).

## Code of Conduct

We follow the [CNCF Code of Conduct](https://github.com/cncf/foundation/blob/main/code-of-conduct.md).

Please contact [private Maintainer mailing list](maintainers@flatcar-linux.org) or the Cloud Native Foundation mediator, conduct@cncf.io, to report an issue.

## Contributing

If you want to start contributing to Nebraska, please check out the [contributing](https://www.flatcar.org/docs/latest/contribute/) documentation.

### Development

For a quickstart on setting up a development environment, please check the [development documentation](https://www.flatcar.org/docs/latest/nebraska/development/).

### User Access

For instructions on how to set up user access, please check the [authorization documentation](https://www.flatcar.org/docs/latest/nebraska/authorization/).

## License

Nebraska is released under the terms of the [Apache 2.0](http://www.apache.org/licenses/LICENSE-2.0), and was forked from the [CoreRoller](https://github.com/coreroller/coreroller) project.
