---
title: Deployment & Authorization
weight: 10
---

Nebraska uses either a noop authentication or OIDC to authenticate and authorize users.

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

## Tuning PostgreSQL auto-vacuum

Autovacuum and autoanalyze in PostgreSQL are effectively disabled when tables
are very large. This is because the default is 20% of a table (and 10%
of a table for analyze).

We advise to change the mentioned configuration in order to have autovacuum
and autoanalyse run when about 5,000 rows change. This value was chosen
based on getting the autovacuum to run every day, as it's large enough
to not cause the autovac to run all the time, but about the right size
to make a difference for query statistics and reducing table bloat.

You can verify (and eventually use) this [SQL file](./autovacuum-tune.sql)
where we have set up these changes.

The analyze threshold was chosen at half the autovacuum threshold
because the defaults are set at half.

# Deploying Nebraska for testing on local computer (noop authentication)

- Go to the nebraska project directory and run `make`

- Start the database (see the section above if you need a quick setup).

- Start the Nebraska backend:

  - `nebraska -auth-mode noop -http-static-dir $PWD/frontend/build
    -http-log`

- In the browser, access `http://localhost:8000`

# Deploying Nebraska with OIDC authentication mode

- Go to the nebraska project directory and run `make`

- Start the database (see the section above if you need a quick setup).

- Setup OIDC provider (Keycloak Recommended).

- Start the Nebraska backend:

  - `nebraska --auth-mode oidc --oidc-admin-roles nebraska_admin  --oidc-viewer-roles nebraska_member --oidc-client-id nebraska --oidc-issuer-url http://localhost:8080/auth/realms/master --oidc-client-secret <Your_Client_Secret>`

  Note: If roles array in the token is not in `roles` key, one can specify a custom JSON path using the `oidc-roles-path` flag.

- In the browser, access `http://localhost:8000`

# Preparing Keycloak as an OIDC provider for Nebraska

- Run `Keycloak` using docker:
    - `docker run -p 8080:8080 -e KEYCLOAK_USER=admin -e KEYCLOAK_PASSWORD=admin -d quay.io/keycloak/keycloak:13.0.1`

- Open http://localhost:8080 in your browser to access keycloak UI and login with the username admin and password as admin.

## Creating Roles

### Member Role

1. Click on `Roles` menu option and select `Add Role`.
2. Provide a name for the member role, here we will use `nebraska_member`.
3. Click `Save`.

### Admin Role

1. Click on `Roles` menu option and select `Add Role`.
2. Provide a name for the admin role, here we will use `nebraska_admin`.
3. Click `Save`.
4. After the admin role is created enable composite role to ON. In the Composite Roles section select the member role, In our case it is nebraska_member and click Add Selected.

Now the member and admin roles are created, the admin role is a composite role which comprises of member role.

<p align="center">
  <img width="100%"  src="./images/keycloak-roles.gif">
</p>

## Creating a client

1. Click on `Clients` menu option and click `Create`.
2. Set the client name as `nebraska` and click `Save`.
3. Change the `Access Type` to `Confidential`
4. Set `Valid Redirect URIs` to `http://localhost:8000/login/cb`.

<p align="center">
  <img width="100%" src="./images/keycloak-client.gif">
</p>

## Adding roles scope to token

1. Click on `Mappers` tab in Client Edit View. Click on `Create`.
2. Set the name as `roles`, Select the `Mapper Type` as `User Realm Role`, `Token Claim Name` as `roles` and Select `Claim JSON Type` as String.
3. Click `Save`

<p align="center">
  <img width="100%" src="./images/keycloak-token.gif">
</p>

## Attaching Roles to User

1. Click on `Users` menu option and click `View all users`.
2. Once the user list appears select the user and click on `Edit`.
3. Go to `Role Mapping` tab and select `nebraska_admin` role and click on add selected to attach role to user. If you want to provide only member access access select the member role.

<p align="center">
  <img width="100%" src="./images/keycloak-user.gif">
</p>

# Preparing Auth0 as an OIDC provider for Nebraska
## Create and configure new application

1. Click on `Create Application`.
2. Provide the name as `nebraska`, select `Regular Web Application`.
3. Click `Create`
4. Click on the `settings` tab.
5. Under `Application URIs` section provide the `Allowed Callback URLs` as `http://localhost:8000/login/cb`.
6. Click on `Save Changes`

<p align="center">
  <img width="100%" src="./images/auth0-application.gif">
</p>


## Adding roles scope to token

1. Click on `Rules` sub-menu from `Auth Pipeline` menu option.
2. Click on `Empty Rule` option.
3. Provide the name as `roles`.
4. Paste the following snippet in `Script` text box.
```js
function (user, context, callback) {
  const namespace = 'http://kinvolk.io';
  const assignedRoles = (context.authorization || {}).roles;

  let idTokenClaims = context.idToken || {};
  let accessTokenClaims = context.accessToken || {};

  idTokenClaims[`${namespace}/roles`] = assignedRoles;
  accessTokenClaims[`${namespace}/roles`] = assignedRoles;

  context.idToken = idTokenClaims;
  context.accessToken = accessTokenClaims;
  callback(null, user, context);
}
```
Now the rule to add the roles to the token is setup, the roles will be available in the key `http://kinvolk.io/roles`.

Note: The `oidc-roles-path` argument accepts a JSONPath to fetch roles from the token, in this case set the value to `http://kinvolk\.io/roles`.

<p align="center">
  <img width="100%" src="./images/auth0-token.gif">
</p>

# Preparing Dex with github connector as an OIDC provider for Nebraska

## Setting up a Github App to be used as a connector for Dex

- Create a new `organization` in Github.

- Now you need a Github app, go to `https://github.com/organizations/<ORG>/settings/apps/new` and fill
  the following fields:

  - `GitHub App name` - just put some fancy name.

  - `Homepage URL` - `http://localhost:8000`

  - `User authorization callback URL` - `http://0.0.0.0:5556/dex/callback`

  - `Permissions` - `Access: Read-only` to `Organization members`

  - `User permissions` - none needed

  - `Subscribe to events` - tick `Membership`, `Organization` and `Team`

  - `Where can this GitHub App be installed?` - `Only on this account`

- Press `Create GitHub App` button

- Next thing you'll get is `OAuth credentials` at the bottom of the
  page of the app you just created, we will need both `Client ID` and
  `Client secret`

- You also need to install the app you just created

  - Go to `https://github.com/organizations/<ORG>/settings/apps`

  - Click `Edit` button for your new app

  - Choose `Install App` on the left of the page and perform the
    installation

## Creating Github Teams

- Create two teams in your organization with the following names
  `admin` and `viewer`.

- Add the admin users to both `admin` and `viewer` team. Add the non-admin users to
`viewer` team.

## Configuring and Running Dex IDP

- Create a configuration for Dex based on the example.

> example.yaml

```yaml
issuer: http://0.0.0.0:5556/dex

storage:
  type: sqlite3
  config:
    file: /var/dex/dex.db

web:
  http: 0.0.0.0:5556

staticClients:
  - id: nebraska
    redirectURIs:
      - 'http://localhost:8000/login/cb'
    name: 'nebraska'
    secret: <ClientSecret> // Random Hash

connectors:
- type: github
  id: github
  name: GitHub
  config:
    clientID: <Client ID>
    clientSecret: <Client Secret>
    redirectURI: http://0.0.0.0:5556/dex/callback
    loadAllGroups: true
    teamNameField: slug
    useLoginAsID: true

enablePasswordDB: true
```

- Run Dex using docker with the example configuration.

> docker run -p 5556:5556 -v ${PWD}/example.yaml:/etc/dex/example.yaml -v ${PWD}/dex.db:/var/dex/dex.db ghcr.io/dexidp/dex:v2.28.1 dex serve /etc/dex/example.yaml


## Running nebraska

> go run ./cmd/nebraska --auth-mode oidc --oidc-admin-roles <organization>:admin  --oidc-viewer-roles <organization>:viewer --oidc-client-id nebraska --oidc-issuer-url http://127.0.0.1:5556/dex --oidc-client-secret <ClientSecret> --oidc-roles-path groups --oidc-scopes groups,openid,profile

# Deploying on Kubernetes using the Helm Chart

We maintain a Helm Chart for deploying a Nebraska instance to Kubernetes. The
Helm Chart offers flexible configuration options such as:

- Deploy a single-replica `PostgreSQL` database together with Nebraska. We use
  the container image and also the Helm Chart (as a subchart) from
  [Bitnami](https://github.com/bitnami/charts/tree/master/bitnami/postgresql)

- Enabling / disabling, and configuring persistence for both Nebraska and PostgreSQL
  (persistence is disabled by default)

- Common deployment parameters (exposing through `Ingress`, replica count, etc.)

- All [Nebraska application configuration](https://github.com/kinvolk/nebraska/tree/main/charts/nebraska#nebraska-configuration)

For the complete list of all available customization options, please read the
[Helm Chart README](https://github.com/kinvolk/nebraska/blob/main/charts/nebraska/README.md).

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
