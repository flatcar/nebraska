---
title: Deployment & Authorization
weight: 10
---

Nebraska uses either a noop authentication, OAuth or Bearer tokens
to authenticate and authorize users. Currently, only GitHub is
supported as an OAuth authentication backend.  [GitHub personal access
tokens](https://github.com/settings/tokens) can be used as bearer
token. If no bearer token is sent, Nebraska will authorize users
through the GitHub Oauth flow and create a persistent session.

# Preparing the database for Nebraska

Nebraska uses the `PostgreSQL` database, and expects the used database to
be set to the UTC timezone.

For a quick setup of `PostgreSQL` for Nebraska's development, you can use
the `postgres` container as follows:

- Start `Postgres`:
    - `docker run --rm -d --name nebraska-postgres-dev -p 5432:5432 -e POSTGRES_PASSWORD=nebraska postgres`

- Create the database for Nebraska (by default it is `nebraska`):
    - `psql postgres://postgres:nebraska@localhost:5432/postgres -c 'create database nebraska;'`

- Set the timezone to Nebraska's database:
    - `psql postgres://postgres:nebraska@localhost:5432/nebraska -c 'set timezone = "utc";'`

# Deploying Nebraska for testing on local computer (noop authentication)

- Go to the nebraska project directory and run `make`

- Start the database (see the section above if you need a quick setup).

- Start the Nebraska backend:

  - `nebraska -auth-mode noop -http-static-dir $PWD/frontend/build
    -http-log`

- In the browser, access `http://localhost:8000`

# Deploying Nebraska for testing/development on local computer (GitHub authentication)

- Go to https://smee.io/ and press the `Start a new channel` button,
  you'll get redirected to something like
  `https://smee.io/asfdFA7B87SD98F7`

- On the computer create some directory for nodejs stuff (say
  `~/nodejsstuff`), then in that directory do `npm install
  smee-client`, after that run `~/nodejsstuff/node_modules/.bin/smee
  -u https://smee.io/asfdFA7B87SD98F7 -p 8000 -P /login/webhook`

- Now you need a github app, go to
  `https://github.com/settings/apps/new` (or, if doing it in some
  organization, to
  `https://github.com/organizations/<ORG>/settings/apps/new`) and fill
  the following fields:

  - `GitHub App name` - just put some fancy name, must be unique I
    think

  - `Homepage URL` - `https://smee.io/asfdFA7B87SD98F7`

  - `User authorization callback URL` - `http://localhost:8000/login/cb`

  - `Webhook URL` - `https://smee.io/asfdFA7B87SD98F7`

  - `Webhook secret` - another secret stuff, this time for webhook
    message validation

  - Enable SSL verification

  - `Permissions` - `Access: Read-only` to `Organization members`

  - `User permissions` - none needed

  - `Subscribe to events` - tick `Membership`, `Organization` and `Team`

  - `Where can this GitHub App be installed?` - `Only on this account`

- Press `Create GitHub App` button

- Next thing you'll get is `OAuth credentials` at the bottom of the
  page of the app you just created, we will need both `Client ID` and
  `Client secret`

- You also need to install the app you just created

  - Go to `https://github.com/settings/apps` (or, in case of an
    organization,
    `https://github.com/organizations/<ORG>/settings/apps`)

  - Click `Edit` button for your new app

  - Choose `Install App` on the left of the page and perform the
    installation

- Now go to the nebraska project directory and run `make all`

- Start the database: `docker run --privileged -d -p
  127.0.0.1:5432:5432 nebraska/postgres`

- Start the Nebraska backend:

  - `nebraska -auth-mode github -gh-client-id <CLIENT_ID>
    -gh-client-secret <CLIENT_SECRET> -gh-ro-teams <READ_ONLY_TEAMS>
    -gh-rw-teams <READ_WRITE_TEAMS> -gh-webhook-secret <WEBHOOK_SECRET>
    -http-static-dir $PWD/frontend/build -http-log`

    - Use the `-gh-rw-teams` and `-gh-ro-teams` to specify the list of
      read-write and read-only access teams.
    - Nebraska uses this list to set the access level of the user accordingly
      and users in read-only teams can only perform `GET` and `HEAD` requests.
    - Nebraska then logs into Github and fetches the list of the Github teams
      of the logged in user and tries to match them against the list of teams
      passed through the CLI
    - If user is not part of any of the teams in the list then Nebraska denies
      access completely and sets the permissions of the session accordingly.
    - Nebraska doesn't support groups but instead assumes that there is only
      one nebraska team in the database. We plan to setup a new nebraska
      instance for separation needed between different customers or projects
    - Run `nebraska -h` to learn about env vars you can use instead of
      flags

- In the browser, access `http://localhost:8000`

# Deploying on remote server (GitHub authentication)

Assuming that our Nebraska instance is reachable through a
`nebraska.example.com` address, then there is no need for using `smee` and
`localhost:8000` that were mentioned above. In that case, the following fields in
the app configuration can have different values:

  - `Homepage URL` - `https://nebraska.example.com`

  - `User authorization callback URL` - `https://nebraska.example.com/login/cb`

  - `Webhook URL` - `https://nebraska.example.com/login/webhook`

Rest of the steps is the same.

# Deploying on Kubernetes using the Helm Chart

We maintain a Helm Chart for deploying a Nebraska instance to Kubernetes. The
Helm Chart offers flexible configuration options such as:

- Deploy a single-replica `PostgreSQL` database together with Nebraska. We use
  the container image and also the Helm Chart (as a subchart) from
  [Bitnami](https://github.com/bitnami/charts/tree/master/bitnami/postgresql)

- Enabling / disabling, and configuring persistence for both Nebraska and PostgreSQL
  (persistence is disabled by default)

- Common deployment parameters (exposing through `Ingress`, replica count, etc.)

- All [Nebraska application configuration](https://github.com/kinvolk/nebraska/tree/master/charts/nebraska#nebraska-configuration)

For the complete list of all available customization options, please read the
[Helm Chart README](https://github.com/kinvolk/nebraska/blob/master/charts/nebraska/README.md).

To install the Helm Chart using the default configuration (noop authentication),
you can execute:

```console
$ helm repo add nebraska https://kinvolk.github.io/nebraska/
$ helm install my-nebraska nebraska/nebraska
```

You probably need to customize the installation, then use a Helm values file.
Eg.:

```yaml
# nebraska-values.yaml
config:
  app:
    title: Nebraska

  auth:
    mode: github
    github:
      clientID: <your clientID obtained during GitHub App registration>
      clientSecret: <your clientSecret obtained during GitHub App registration>
      sessionAuthKey: <64 random hexadecimal characters>
      sessionCryptKey: <32 random hexadecimal characters>
      webhookSecret: <random Secret used in GitHub App registration>

ingress:
  annotations:
    kubernetes.io/ingress.class: <your ingress class>
  hosts:
    - nebraska.example.com

postgresql:
  postgresqlPassword: <A secure password>
  persistence:
    enabled: true
    storageClass: <A block storage-class>
    accessModes:
      - ReadWriteOnce
    size: 1Gi
```

Then execute:

```console
$ helm install my-nebraska nebraska/nebraska --values nebraska-values.yaml
```

# Personal Access Tokens (GitHub authentication)

How a Nebraska instance does authentication depends on existence of the
`Authorization` header in the first request. Basically if the header
does not exist then Nebraska will go with the "Login through GitHub"
route, which means redirecting to github, authenticating there if you
weren't logged in before, authorizing the app if it wasn't authorized
before. This is not exactly friendly for some services (for example,
prometheus won't work with that).

If the `Authorization` header exists, it must be like `Authorization:
bearer <TOKEN>`. The `<TOKEN>` part should be personal access token
generated on github (`https://github.com/settings/tokens`).

Personal access token requires just one scope: `read:org`.

# GitHub Enterprise

If you want to authenticate against GitHub Enterprise, set `-gh-enterprise-url`
to the URL of your GitHub Enterprise instance, for example
`https://github.myorganization.net`.

After that, follow the same procedures as with GitHub.

# Troubleshooting:

- I'm getting a blank page!

  - You likely visited nebraska frontend website before, so browser
    likely has cached the `index.html` page, so it won't get it from
    Nebraska, but instead start asking for some CSS and javascript
    stuff outright, which it won't get. That results in a blank
    page. Force the browser to get `index.html` from Nebraska by
    either doing a force refresh (ctrl+f5 on firefox), or by cleaning
    the cache for localhost (or the server where the Nebraska instance
    is deployed). We will try to improve this in the future.
