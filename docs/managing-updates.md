---
title: Managing Flatcar updates
weight: 10
---

Once you have Nebraska up and running, a common use-case is to manage Flatcar Container Linux updates.

By default, your Flatcar Container Linux instances use the public servers to get updates, so you have to point them to your Nebraska deployment for it to
manage their updates. The process for doing so is slightly different depending on whether you have existing machines or new ones.

## New machines

For new machines, you can set up the updates server in their Container Linux Configuration. Here is a small example of how to do it:


```yaml
storage:
  files:
    - path: /etc/flatcar/update.conf
      filesystem: root
      mode: 0644
      contents:
        inline: |
          GROUP=stable
          SERVER=http://your.nebraska.host:port/v1/update/
```

In addition to the default `stable`, `beta` and `alpha` groups, you can also create and use **custom groups** for greater control over the updates. In that case, you **must** use the group id (not the name) you will find next to the group name in the dashboard.


```yaml
storage:
  files:
    - path: /etc/flatcar/update.conf
      filesystem: root
      mode: 0644
      contents:
        inline: |
          GROUP=ab51a000-02dc-4fc7-a6b0-c42881c89856
          SERVER=http://your.nebraska.host:port/v1/update/
```

**Note**: The sample Nebraska containers provided use the `port 8000` by default (**plain HTTP, no SSL**). Please adjust the update URL setup in your servers to match your Nebraska deployment.

## Existing machines

To update the update server in existing instances please edit `/etc/flatcar/update.conf` and update the `SERVER` value (and optionally `GROUP` if needed):

	SERVER=https://your.nebraska.host/v1/update/

Again, when using custom groups instead of the official ones (stable, beta, alpha) the group id **must** be used, not the group name:

    GROUP=ab51a000-02dc-4fc7-a6b0-c42881c89856

To apply these changes run:

	sudo systemctl restart update-engine

In may take a few minutes to see an update request coming through. If you want to see it sooner, you can force it running this command:

	update_engine_client -update

## Flatcar Container Linux packages in Nebraska

Nebraska is able to periodically poll the public Flatcar Container Linux update servers and create new packages to update the corresponding channels. So if Nebraska is connected to the internet, new packages will show up automatically for the official Flatcar Container Linux. This functionality is optional, and turned off by default. If you
prefer to use it, you should pass the option `-enable-syncer=true` when running Nebraska.

Notice that by default Nebraska only stores metadata about the Flatcar Container Linux updates, not the updates payload. This means that the updates served to your instances contain instructions to download the packages payload from the public Flatcar Container Linux update servers directly, so your servers need access to the Internet to download them.

It is also possible to host the Flatcar Container Linux packages payload in Nebraska. In this case, in addition to get the packages metadata, Nebraska will also download the package payload itself so that it can serve it to your instances when serving updates.

This functionality is turned off by default. So to make Nebraska host the Flatcar Container Linux packages payload, the following options have to be passed to it:

    nebraska -host-flatcar-packages=true -flatcar-packages-path=/PATH/TO/STORE/PACKAGES -nebraska-url=http://your.Nebraska.host:port

## Managing updates for your own applications

In addition to managing updates for Flatcar Container Linux, you can use Nebraska for other applications as well.

In the `updaters/lib` directory there are some sample helpers that can be useful to create your own updaters that talk to Nebraska or even embed them into your own applications.

In the `updaters/examples` you'll find a sample minimal application built using [grace](https://github.com/facebookgo/grace) that is able to update itself using Nebraska in a graceful way.
