# Nebraska Helm Chart

Nebraska is an update manager for Flatcar Container Linux.

## TL;DR

```console
$ helm repo add nebraska https://flatcar.github.io/nebraska
$ helm install my-nebraska nebraska/nebraska
```

## Upgrading to 2.0.0

**Breaking Changes for OIDC Users**

Helm chart version 2.0.0 (app version 3.0.0) includes breaking changes to OIDC authentication. If you are using OIDC authentication, you **must** migrate your configuration.

### What Changed

The OIDC implementation has been refactored to use Authorization Code Flow with PKCE for improved security. The backend is now stateless and the frontend handles OIDC authentication directly.

**Removed Configuration Options:**
- `config.auth.oidc.clientSecret` - No longer needed (public client)
- `config.auth.oidc.validRedirectURLs` - Frontend handles redirects
- `config.auth.oidc.sessionAuthKey` - Backend is stateless
- `config.auth.oidc.sessionCryptKey` - Backend is stateless

**Added Configuration Options:**
- `config.auth.oidc.audience` - Optional, required for Auth0
- `config.auth.oidc.useUserInfo` - Use UserInfo endpoint for role extraction (for providers that don't include roles in access token)

**All Other Options Remain:** `clientID`, `issuerURL`, `managementURL`, `logoutURL`, `adminRoles`, `viewerRoles`, `rolesPath`, `scopes`

### Migration Steps

1. **Update your OIDC provider configuration:**
   - Change client type from "Confidential" to "Public" (SPA)
   - Remove client secret
   - Update redirect URI to: `https://your-domain.com/auth/callback`
   - Enable CORS for your Nebraska domain

2. **Update your helm values:**
   ```yaml
   config:
     auth:
       mode: oidc
       oidc:
         clientID: "your-public-client-id"
         issuerURL: "https://your-oidc-provider.com"
         adminRoles: "nebraska-admin"
         viewerRoles: "nebraska-viewer"
         # Remove: clientSecret, validRedirectURLs, sessionAuthKey, sessionCryptKey
         # Optional: audience (required for Auth0)
   ```

3. **Upgrade the helm chart:**
   ```bash
   helm upgrade my-nebraska nebraska/nebraska --version 2.0.0
   ```

**For detailed migration instructions, see the [OIDC Migration Guide](https://github.com/flatcar/nebraska/blob/main/docs/oidc-migration-guide.md).**

**Note:** If you are using `mode: noop` (default) or `mode: github`, no changes are required.

## Upgrade PostgreSQL

When there is a major upgrade of PostgreSQL, a manual intervention might be required with a downtime. It is possible to automate things with operators, but here's a simple example:

1. Scale down Nebraska deployment:
```
$ kubectl scale --replicas=0 deployment/nebraska
deployment.apps/nebraska scaled
```

2. Backup PostgreSQL data:
```
$ kubectl exec -ti pod/nebraska-postgresql-0 -- pg_dumpall > backup.sql
```

3. Scale down Nebraska statefulset:
```
$ kubectl scale --replicas=0 statefulset/nebraska-postgresql
statefulset.apps/nebraska-postgresql scaled
```

4. Backup and remove the data from the bound volume (depending on the storage class)

3. Upgrade PostgreSQL version, e.g:
```diff
-          image: docker.io/bitnami/postgresql:13.8.0-debian-11-r18
+          image: docker.io/bitnamilegacy/postgresql:17.5.0
```

5. Apply the changes and scale up Nebraska statefulset to its original value

6. Inject the backup and assert that everything looks good in the database:
```
$ kubectl exec -ti pod/nebraska-postgresql-0 -- psql < backup.sql
```

7. Scale up Nebraska deployment and assert that everything is back to normal

## Parameters

### Global parameters

| Parameter                 | Description                                                                           | Default |
|---------------------------|---------------------------------------------------------------------------------------|---------|
| `global.imageRegistry`    | Global Container image registry                                                       | `nil`   |
| `extraObjects`            | List of extra manifests to deploy. Will be passed through `tpl` to support templating | `[]`    |

### Nebraska parameters

| Parameter                               | Description                                                                                                                              | Default                               |
|-----------------------------------------|------------------------------------------------------------------------------------------------------------------------------------------|---------------------------------------|
| `replicaCount`                          | Number of desired pods                                                                                                                   | `1`                                   |
| `image.registry`                        | Container image registry                                                                                                                 | `ghcr.io`                             |
| `image.repository`                      | Container image name                                                                                                                     | `flatcar/nebraska`                    |
| `image.tag`                             | Container image tag                                                                                                                      | `""` (use appVersion in `Chart.yaml`) |
| `image.pullPolicy`                      | Image pull policy. One of `Always`, `Never`, `IfNotPresent`                                                                              | `IfNotPresent`                        |
| `image.pullSecrets`                     | An optional list of references to secrets in the same namespace to use for pulling any of the images used                                | `[]`                                  |
| `nameOverride`                          | Overrides the name of the chart                                                                                                          | `""`                                  |
| `fullnameOverride`                      | Overrides the full name of the chart                                                                                                     | `""`                                  |
| `serviceAccount.create`                 | Specifies whether a service account should be created                                                                                    | `false`                               |
| `serviceAccount.annotations`            | Annotations to add to the service account                                                                                                | `{}`                                  |
| `serviceAccount.name`                   | The name of the service account to use. (If not set and create is true, a name is generated using the fullname template)                 | `{}`                                  |
| `strategy.type`                         | Type of deployment. Can be `Recreate` or `RollingUpdate`                                                                                 | `Recreate`                            |
| `strategy.rollingUpdate.maxSurge`       | The maximum number of pods that can be scheduled above the desired number of pods (Only applies when `strategy.type` is `RollingUpdate`) | `nil`                                 |
| `strategy.rollingUpdate.maxUnavailable` | The maximum number of pods that can be unavailable during the update (Only applies when `strategy.type` is `RollingUpdate`)              | `nil`                                 |
| `podAnnotations`                        | Annotations for pods                                                                                                                     | `nil`                                 |
| `podLabels`                             | Labels for pods                             |                                                                                            | `nil`                                 |
| `extraLabels`                           | Additional labels that will be applied to all objects |                                                                                  | `nil`                                 |
| `extraAnnotations`                      | Additional annotations that will be applied to all objects |                                                                             | `nil`                                 |
| `podSecurityContext`                    | Holds pod-level security attributes and common container settings                                                                        | Check `values.yaml` file              |
| `securityContext`                       | Security options the container should run with                                                                                           | `nil`                                 |
| `service.type`                          | Kubernetes Service type                                                                                                                  | `ClusterIP`                           |
| `service.port`                          | Kubernetes Service port                                                                                                                  | `80`                                  |
| `ingress.enabled`                       | Enable ingress controller resource                                                                                                       | `true`                                |
| `ingress.annotations`                   | Annotations for Ingress resource                                                                                                         | `{}`                                  |
| `ingress.hosts`                         | Hostname(s) for the Ingress resource                                                                                                     | `["flatcar.example.com"]`             |
| `ingress.ingressClassName`              | Ingress controller which implements the resource. This replaces the deprecated `kubernetes.io/ingress.class` annotation on K8s > 1.19    | `""`                                  |
| `ingress.tls`                           | Ingress TLS configuration                                                                                                                | `[]`                                  |
| `ingress.update.enabled`                | Create a separate ingress for the `/v1/update` and `/flatcar` paths, with it's own annotations.                                          | `false`                               |
| `ingress.update.annotations`            | Annotations for Ingress resource                                                                                                         | `{}`                                  |
| `ingress.update.ingressClassName`       | Ingress controller which implements the resource. This replaces the deprecated `kubernetes.io/ingress.class` annotation on K8s > 1.19    | `""`                                  |
| `resources`                             | CPU/Memory resource requests/limits                                                                                                      | `{}`                                  |
| `nodeSelector`                          | Node labels for pod assignment                                                                                                           | `{}`                                  |
| `tolerations`                           | Toleration labels for pod assignment                                                                                                     | `[]`                                  |
| `affinity`                              | Affinity settings for pod assignment                                                                                                     | `{}`                                  |
| `livenessProbe`                         | Liveness Probe settings                                                                                                                  | Check `values.yaml` file              |
| `readinessProbe`                        | Readiness Probe settings                                                                                                                 | Check `values.yaml` file              |

### Nebraska Configuration

| Parameter                                             | Description                                                                                                                          | Default                                                                 |
|-------------------------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------|-------------------------------------------------------------------------|
| `config.app.logoPath`                                 | Client app logo, should be a path to svg file                                                                                        | `""`                                                                    |
| `config.app.title`                                    | Client app title                                                                                                                     | `""                    `                                                |
| `config.app.headerStyle`                              | Client app header style, should be either `dark` or `light`                                                                          | `""`                                                                    |
| `config.app.httpStaticDir`                            | Path to frontend static files                                                                                                        | `/nebraska/static`                                                      |
| `config.syncer.enabled`                               | Enable Flatcar packages syncer                                                                                                       | `true`                                                                  |
| `config.syncer.interval`                              | Sync check interval (the minimum depends on the number of channels to sync, e.g., `8m` for 8 channels incl. different architectures) | `nil` (uses app defaults of `1h`)                                       |
| `config.syncer.updateURL`                             | Flatcar update URL to sync from (default "https://public.update.flatcar-linux.net/v1/update/")                                       | `nil` (uses app defaults)                                               |
| `config.hostFlatcarPackages.enabled`                  | Host Flatcar packages in Nebraska                                                                                                    | `false`                                                                 |
| `config.hostFlatcarPackages.packagesPath`             | Path where Flatcar packages files should be stored                                                                                   | `/mnt/packages`                                                         |
| `config.hostFlatcarPackages.nebraskaURL`              | Nebraska URL (`http://host:port`)                                                                                                    | `nil` (defaults to first ingress host)                                  |
| `config.hostFlatcarPackages.persistence.enabled`      | Enable persistence using PVC                                                                                                         | `false`                                                                 |
| `config.hostFlatcarPackages.persistence.labels        | Additional labels to be applied to the PVC                                |                                                          | `nil`                                                                   |
| `config.hostFlatcarPackages.persistence.annotations   | Additional annotations to be applied to the PVC                           |                                                          | `nil`                                                                   |
| `config.hostFlatcarPackages.persistence.storageClass` | PVC Storage Class for PostgreSQL volume                                                                                              | `nil`                                                                   |
| `config.hostFlatcarPackages.persistence.accessModes`  | PVC Access Mode for PostgreSQL volume                                                                                                | `["ReadWriteOnce"]`                                                     |
| `config.hostFlatcarPackages.persistence.size`         | PVC Storage Request for PostgreSQL volume                                                                                            | `10Gi`                                                                  |
| `config.auth.mode`                                    | Authentication mode, available modes: `noop`, `github`, `oidc`                                                                               | `noop`                                                                  |
| `config.auth.github.clientID`                         | GitHub client ID used for authentication                                                                                             | `nil`                                                                   |
| `config.auth.github.clientSecret`                     | GitHub client secret used for authentication                                                                                         | `nil`                                                                   |
| `config.auth.github.existingSecret`                    | existingSecret will mount a given secret to the container. Be sure to match the expected keys in [deployment.yaml](./templates/deployment.yaml) |`nil`                                                                               |                                                                   |
| `config.auth.github.sessionAuthKey`                   | Session secret used for authenticating sessions in cookies used for storing GitHub info , will be generated if none is passed        | `nil`                                                                   |
| `config.auth.github.sessionCryptKey`                  | Session key used for encrypting sessions in cookies used for storing GitHub info, will be generated if none is passed                | `nil`                                                                   |
| `config.auth.github.webhookSecret`                    | GitHub webhook secret used for validing webhook messages                                                                             | `nil`                                                                   |
| `config.auth.github.readWriteTeams`                   | comma-separated list of read-write GitHub teams in the org/team format                                                               | `nil`                                                                   |
| `config.auth.github.readOnlyTeams`                    | comma-separated list of read-only GitHub teams in the org/team format                                                                | `nil`                                                                   |
| `config.auth.github.enterpriseURL`                    | Base URL of the enterprise instance if using GHE                                                                                     | `nil`    |
| `config.auth.oidc.clientID`                           | OIDC client ID used for authentication (public client)  | `nil`  |
| `config.auth.oidc.existingSecret`                      | existingSecret will mount a given secret to the container. Be sure to match the expected keys in [deployment.yaml](./templates/deployment.yaml). |`nil`                                                                               |                                                                   |
| `config.auth.oidc.issuerURL`                          | OIDC issuer URL used for authentication | `nil`  |
| `config.auth.oidc.managementURL`                      | OIDC management url for managing the account  | `nil`  |
| `config.auth.oidc.logoutURL`                          | URL to logout the user from current session  | `nil`  |
| `config.auth.oidc.adminRoles`                         | comma-separated list of accepted roles with admin access | `nil`  |
| `config.auth.oidc.viewerRoles`                        | comma-separated list of accepted roles with viewer access | `nil`  |
| `config.auth.oidc.rolesPath`                          | json path in which the roles array is present in the id token  | `nil`  |
| `config.auth.oidc.scopes`                             | comma-separated list of scopes to be used in OIDC | `nil`  |
| `config.auth.oidc.audience`                           | OIDC audience (required for Auth0, optional for others) | `nil`  |
| `config.auth.oidc.useUserInfo`                        | Use UserInfo endpoint for role extraction (for providers that don't include roles in access token) | `false`  |
| `config.database.host`                                | The host name of the database server                                                                                                 | `""` (use postgresql from Bitnami subchart)                             |
| `config.database.port`                                | The port number the database server is listening on                                                                                  | `5432`                                                                  |
| `config.database.sslMode`                             | The mode of the database connection                                                                                                  | `disable`                                                               |
| `config.database.dbname`                              | The database name                                                                                                                    | `{{ .Values.postgresql.auth.database }}` (evaluated as a template)      |
| `config.database.username`                            | PostgreSQL user                                                                                                                      | `{{ .Values.postgresql.postgresqlUsername }}` (evaluated as a template)                                    |
| `config.database.password`                            | PostgreSQL user password                                                                                                             | `""` (evaluated as a template)                                          |
| `config.database.passwordExistingSecret.enabled`      | Enables setting PostgreSQL user password via an existing secret                                                                      | `true`                                                                  |
| `config.database.passwordExistingSecret.name`         | Name of the existing secret                                                                                                          | `{{ .Release.Name }}-postgresql` (evaluated as a template)              |
| `config.database.passwordExistingSecret.key`          | Key inside the existing secret containing the PostgreSQL user password                                                               | `postgres-password`                                                     |
| `extraArgs`                                           | Extra arguments to pass to Nebraska binary                                                                                           | `[]`                                                                    |
| `extraEnvVars`                                        | Any extra environment variables you would like to pass on to the pod                                                                 | `{ "TZ": "UTC" }`                                                       |
| `extraEnv`                                        | Any extra environment variables in the form of env spec to pass into the deployment pod                                                                 | `[]`                                                       |

### Postgresql dependency

| Parameter                                                | Description                                                                                                   | Default                |
|----------------------------------------------------------|---------------------------------------------------------------------------------------------------------------|------------------------|
| `postgresql.enabled`                                     | Enable Bitnami postgresql subchart and deploy database within this helm release                               | `true`                 |
| `postgresql.auth.database`                               | PostgreSQL database                                                                                           | `nebraska`             |
| `postgresql.auth.postgresPassword`                       | PostgreSQL password of user "postgres" **Recommended to change it to something secure for security reasons.** | `changeIt`             |
| `postgresql.image.tag`                                   | PostgreSQL Image tag                                                                                          | `13.8.0-debian-11-r18` |
| `postgresql.primary.persistence.enabled`                 | Enable persistence using PVC                                                                                  | `false`                |
| `postgresql.primary.persistence.storageClass`            | PVC Storage Class for PostgreSQL volume                                                                       | `nil`                  |
| `postgresql.primary.persistence.accessModes`             | PVC Access Mode for PostgreSQL volume                                                                         | `["ReadWriteOnce"]`    |
| `postgresql.primary.persistence.size`                    | PVC Storage Request for PostgreSQL volume                                                                     | `1Gi`                  |
| `postgresql.serviceAccount.create`                       | Enable creation of ServiceAccount for PostgreSQL pod                                                          | `true`                 |
| `postgresql.serviceAccount.automountServiceAccountToken` | Can be set to false if pods using this serviceAccount do not need to use K8s API                              | `false`                |

... for more options see https://github.com/bitnami/charts/tree/master/bitnami/postgresql
