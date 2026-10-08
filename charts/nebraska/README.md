# Nebraska Helm Chart

Nebraska is an update manager for Flatcar Container Linux.

## TL;DR

```console
$ helm repo add nebraska https://flatcar.github.io/nebraska
$ helm install my-nebraska nebraska/nebraska
```

## Upgrading to 4.0.0

**Breaking change:** the bundled PostgreSQL is no longer the Bitnami subchart.
It is now a small StatefulSet in this chart, running the official
`docker.io/postgres:17-bookworm` image. The `bitnamilegacy` image no longer gets
security fixes (flatcar/nebraska#1574), and the Bitnami chart repository is
deprecated (flatcar/nebraska#1148).

| If you use... | Then... |
|---------------|---------|
| an external database (`postgresql.enabled: false`) | No action. |
| the default `postgresql.primary.persistence.enabled: false` | No action. The database was already temporary. It starts empty, and Nebraska is restarted so that it recreates its schema. |
| `postgresql.primary.persistence.enabled: true` | Action required. Do an [in-place upgrade](#in-place-upgrade) (same PostgreSQL major version) or a [dump and restore](#dump-and-restore). Otherwise the chart refuses to upgrade. |

Other changes:

* Old Bitnami `postgresql.*` values are refused, and the error names the
  replacement value, if there is one. Delete the old keys.
* The Secret has only the `postgres-password` key. The old `password` key is
  removed, so update anything that reads it before you upgrade.
* A new install gets a generated password. On an upgrade, the password already
  in the Secret is kept. To read it:
  `kubectl get secret my-nebraska-postgresql -o jsonpath='{.data.postgres-password}' | base64 -d`
* The official image has no `pgaudit`, no metrics exporter, no `POSTGRESQL_*`
  variables and no `/opt/bitnami` scripts.
* Keep a `-bookworm` tag. Other tags change the sort order (collation) of your
  existing data, and Alpine tags also need uid/gid 70.

### In-place upgrade

Use this only for the same PostgreSQL major version (chart 3.0.x shipped 17).
Snapshot the PV or set it to `Retain` first. Add these values to your values
file, then run `helm upgrade -f <your values> --set postgresql.acknowledgeDataDirMigration=true`:

```yaml
postgresql:
  dataMountPath: /bitnami/postgresql   # mount the PVC where Bitnami mounted it
  dataSubdir: data                     # PGDATA = /bitnami/postgresql/data
  podSecurityContext:                  # the files on the volume are owned by uid/gid 1001
    runAsNonRoot: true
    runAsUser: 1001
    runAsGroup: 1001
    fsGroup: 1001
    fsGroupChangePolicy: OnRootMismatch
    seccompProfile: { type: RuntimeDefault }
  extraPodSpec:
    initContainers:                    # writes the config files that Bitnami kept in its image
      - name: pgconf-bootstrap
        image: docker.io/postgres:17-bookworm
        command: [/bin/sh, -c]
        args:
          - |
            set -e
            cd /bitnami/postgresql/data
            [ -f postgresql.conf ] || printf "listen_addresses = '*'\n" > postgresql.conf
            [ -f pg_hba.conf ] || printf "local all all trust\nhost all all 127.0.0.1/32 trust\nhost all all ::1/128 trust\nhost all all all scram-sha-256\n" > pg_hba.conf
            [ -f pg_ident.conf ] || echo "# no mappings" > pg_ident.conf
        securityContext: { runAsUser: 1001 }
        volumeMounts:
          - name: data
            mountPath: /bitnami/postgresql
```

* The last `pg_hba.conf` line expects `scram-sha-256` passwords. If this
  database was upgraded from PostgreSQL 13 or older, check this before you
  upgrade. If the command prints `md5`, change `scram-sha-256` to `md5`:
  ```console
  $ kubectl exec my-nebraska-postgresql-0 -- psql -U postgres -tAc \
      "select distinct substring(rolpassword for 4) from pg_authid where rolpassword is not null"
  ```
* Keep `dataMountPath`, `dataSubdir` and `podSecurityContext` for as long as
  you use this volume. You can remove the init container after the first start.

### Dump and restore

Run these steps before you upgrade the chart, because the dump must come from
the old pod. If the release is in another namespace, add `-n <namespace>` to
each command.

```console
# 1. Stop Nebraska.
$ kubectl scale --replicas=0 deployment/my-nebraska

# 2. Dump the database. Check the count: a wrong -U or -d gives an empty dump.
$ PGPW=$(kubectl get secret my-nebraska-postgresql -o jsonpath='{.data.postgres-password}' | base64 -d)
$ kubectl exec my-nebraska-postgresql-0 -- env PGPASSWORD="$PGPW" pg_dump -U postgres -d nebraska > nebraska.sql
$ grep -c '^COPY public\.' nebraska.sql   # must be more than 0

# 3. Keep the PV when the PVC is deleted, in case the dump is bad.
$ PV=$(kubectl get pvc data-my-nebraska-postgresql-0 -o jsonpath='{.spec.volumeName}')
$ kubectl patch pv "$PV" -p '{"spec":{"persistentVolumeReclaimPolicy":"Retain"}}'

# 4. Delete the old StatefulSet, pod and PVC.
$ kubectl delete statefulset my-nebraska-postgresql --cascade=orphan
$ kubectl delete pod my-nebraska-postgresql-0
$ kubectl delete pvc data-my-nebraska-postgresql-0

# 5. Upgrade with Nebraska kept at 0 replicas. Pass your own values file:
#    without it, persistence is off. Do not use --reuse-values, because it
#    keeps the old bitnamilegacy image.
$ helm upgrade my-nebraska nebraska/nebraska --version 4.0.0 \
    -f my-values.yaml \
    --set replicaCount=0 \
    --set postgresql.primary.persistence.enabled=true \
    --set postgresql.acknowledgeDataDirMigration=true \
    --wait --timeout 10m

# 6. Wait for the new database.
$ kubectl wait --for=condition=Ready pod/my-nebraska-postgresql-0 --timeout=300s

# 7. Restore. If any statement fails, nothing is restored.
$ kubectl exec -i my-nebraska-postgresql-0 -- \
    env PGPASSWORD="$PGPW" psql -U postgres -d nebraska \
      -v ON_ERROR_STOP=1 --single-transaction < nebraska.sql

# 8. Update the planner statistics.
$ kubectl exec -i my-nebraska-postgresql-0 -- \
    env PGPASSWORD="$PGPW" psql -U postgres -d nebraska -c 'ANALYZE'

# 9. Check the data. Compare with the same query run before the upgrade.
$ kubectl exec -i my-nebraska-postgresql-0 -- \
    env PGPASSWORD="$PGPW" psql -U postgres -d nebraska -tAc \
    'select (select count(*) from application) as apps,
            (select count(*) from groups) as groups,
            (select count(*) from package) as packages'

# 10. Start Nebraska again.
$ helm upgrade my-nebraska nebraska/nebraska --version 4.0.0 \
    -f my-values.yaml \
    --set postgresql.primary.persistence.enabled=true \
    --set postgresql.acknowledgeDataDirMigration=true
```

Keep `nebraska.sql` and the retained PV until Nebraska works.

**Did you upgrade by mistake and now see no data?** Do not delete anything. The
old database is still in the `data/` directory of the volume. Run
`helm rollback my-nebraska` to get it back. With Helm 4, if the rollback reports
a conflict with `kubectl-patch`, add `--force-conflicts`.

### GitOps

* `helm template` cannot read the existing Secret, so every render creates a
  new password. If you apply that output, PostgreSQL cannot log in after its
  next restart. Use `postgresql.auth.existingSecret` instead.
* The chart checks the data directory only during `helm upgrade` (Flux runs
  one). It cannot check it with `helm template` (Argo CD), so do the steps
  above yourself.

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
- `config.auth.oidc.audience` - Required API/resource audience expected in access tokens
- `config.auth.oidc.skipAudienceCheck` - Insecure migration escape hatch; disables audience validation
- `config.auth.oidc.useUserInfo` - Use UserInfo endpoint for role extraction (for providers that don't include roles in access token)
- `config.caFile` - Path to a PEM-encoded CA certificate file to trust for TLS verification

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
         audience: "your-nebraska-api-identifier"
         # Remove: clientSecret, validRedirectURLs, sessionAuthKey, sessionCryptKey
   ```

3. **Upgrade the helm chart:**
   ```bash
   helm upgrade my-nebraska nebraska/nebraska --version 2.0.0
   ```

**For detailed migration instructions, see the [OIDC Migration Guide](https://github.com/flatcar/nebraska/blob/main/docs/oidc-migration-guide.md).**

**Note:** If you are using `mode: noop` (default) or `mode: github`, no changes are required.

## Upgrade PostgreSQL

> **⚠️ Important:** Nebraska requires **PostgreSQL 14 or newer**. PostgreSQL 17.x is recommended. If you are running PostgreSQL 13 or older, you must upgrade your database before upgrading Nebraska past 4.0.0.

When there is a major upgrade of PostgreSQL, a manual intervention might be required with a downtime. It is possible to automate things with operators, but here's a simple example:

1. Scale down Nebraska deployment:
```
$ kubectl scale --replicas=0 deployment/nebraska
deployment.apps/nebraska scaled
```

2. Backup PostgreSQL data. `-U` must be a superuser: `postgres` after an in-place upgrade from 3.0.x, otherwise `postgresql.auth.username`:
```
$ kubectl exec -i pod/nebraska-postgresql-0 -- pg_dumpall -U postgres > backup.sql
```

3. Scale down Nebraska statefulset:
```
$ kubectl scale --replicas=0 statefulset/nebraska-postgresql
statefulset.apps/nebraska-postgresql scaled
```

4. Backup and remove the data from the bound volume (depending on the storage class)

5. Upgrade PostgreSQL version, e.g:
```diff
-    tag: 17-bookworm
+    tag: 18-bookworm
```
   PostgreSQL 18 uses a different data path, so change the mount path too:

   | major | set `dataMountPath` | set `dataSubdir` |
   |-------|---------------------|------------------|
   | 17    | `/var/lib/postgresql/data` | `pgdata` |
   | 18    | `/var/lib/postgresql`      | `18/docker` |

   The chart refuses a wrong combination if it can read the major version from
   the tag. With persistence enabled, also add
   `--set postgresql.acknowledgeDataDirMigration=true`. The new path starts
   empty, and you restore the dump in step 7.

6. Apply the changes and scale up Nebraska statefulset to its original value

7. Inject the backup and assert that everything looks good in the database:
```
$ kubectl exec -i pod/nebraska-postgresql-0 -- psql -U postgres -d postgres < backup.sql
```

8. Scale up Nebraska deployment and assert that everything is back to normal

## Parameters

### Global parameters

| Parameter                 | Description                                                                           | Default |
|---------------------------|---------------------------------------------------------------------------------------|---------|
| `global.imageRegistry`    | Global Container image registry                                                       | `nil`   |
| `global.imagePullSecrets` | Pull secrets added to the PostgreSQL pod, as the Bitnami subchart did                 | `nil`   |
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
| `ingress.update.enabled`                | Create a separate ingress for the `/v1/update` and `/flatcar` paths, with its own annotations.                                           | `false`                               |
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
| `config.hostFlatcarPackages.persistence.labels`       | Additional labels to be applied to the PVC                                                                                           | `nil`                                                                   |
| `config.hostFlatcarPackages.persistence.annotations`  | Additional annotations to be applied to the PVC                                                                                      | `nil`                                                                   |
| `config.hostFlatcarPackages.persistence.storageClass` | PVC Storage Class for the Flatcar packages volume                                                                                    | `nil`                                                                   |
| `config.hostFlatcarPackages.persistence.accessModes`  | PVC Access Mode for the Flatcar packages volume                                                                                      | `["ReadWriteOnce"]`                                                     |
| `config.hostFlatcarPackages.persistence.size`         | PVC Storage Request for the Flatcar packages volume                                                                                  | `10Gi`                                                                  |
| `config.caFile`                                       | Path to a PEM-encoded CA certificate file to trust for TLS verification (additive to system CAs, used for OIDC and syncer) | `nil`  |
| `extraVolumes`                                        | Additional volumes for the Nebraska pod, e.g. a Secret holding the CA file pointed at by `config.caFile` | `[]` |
| `extraVolumeMounts`                                   | Additional volume mounts for the Nebraska container | `[]` |
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
| `config.auth.oidc.rolesPath`                          | json path in which the roles array is present in the access token  | `nil`  |
| `config.auth.oidc.scopes`                             | comma-separated list of scopes to be used in OIDC | `nil`  |
| `config.auth.oidc.audience`                           | Required API/resource audience expected in OIDC access tokens | `nil`  |
| `config.auth.oidc.skipAudienceCheck`                  | Disable access-token audience validation (insecure migration escape hatch) | `false` |
| `config.auth.oidc.useUserInfo`                        | Use UserInfo endpoint for role extraction (for providers that don't include roles in access token) | `false`  |
| `config.database.host`                                | The host name of the database server                                                                                                 | `""` (use the PostgreSQL bundled with this chart)                             |
| `config.database.port`                                | The port number the database server is listening on                                                                                  | `""` (follows `postgresql.service.port` when bundled, else 5432)        |
| `config.database.sslMode`                             | The mode of the database connection                                                                                                  | `disable`                                                               |
| `config.database.dbname`                              | The database name                                                                                                                    | `{{ .Values.postgresql.auth.database }}` (evaluated as a template)      |
| `config.database.username`                            | PostgreSQL user                                                                                                                      | `{{ .Values.postgresql.auth.username }}` (evaluated as a template)                                    |
| `config.database.password`                            | PostgreSQL user password                                                                                                             | `""` (evaluated as a template)                                          |
| `config.database.passwordExistingSecret.enabled`      | Enables setting PostgreSQL user password via an existing secret                                                                      | `true`                                                                  |
| `config.database.passwordExistingSecret.name`         | Name of the existing secret                                                                                                          | `{{ include "nebraska.postgresql.secretName" . }}` (bundled DB: follows `existingSecret`/name overrides; external DB: `<release>-postgresql`) |
| `config.database.passwordExistingSecret.key`          | Key inside the existing secret containing the PostgreSQL user password                                                               | `{{ .Values.postgresql.auth.secretKeys.adminPasswordKey }}` with the bundled DB, `postgres-password` with an external DB |
| `extraArgs`                                           | Extra arguments to pass to Nebraska binary                                                                                           | `[]`                                                                    |
| `extraEnvVars`                                        | Any extra environment variables you would like to pass on to the pod                                                                 | `{ "TZ": "UTC" }`                                                       |
| `extraEnv`                                        | Any extra environment variables in the form of env spec to pass into the deployment pod                                                                 | `[]`                                                       |

### Bundled PostgreSQL parameters

| Parameter                                                | Description                                                                                                   | Default                |
|----------------------------------------------------------|---------------------------------------------------------------------------------------------------------------|------------------------|
| `postgresql.enabled`                                     | Deploy the PostgreSQL StatefulSet bundled with this chart                                                     | `true`                 |
| `postgresql.acknowledgeDataDirMigration` | Allow a persistent upgrade that points PostgreSQL at a new, empty data directory (see "Upgrading to 4.0.0") | `false` |
| `postgresql.nameOverride` | Override the PostgreSQL name part of the object names | `""` |
| `postgresql.fullnameOverride` | Override the full PostgreSQL object name | `""` |
| `postgresql.auth.username` | PostgreSQL superuser created on first start (`POSTGRES_USER`) | `postgres` |
| `postgresql.auth.database`                               | PostgreSQL database                                                                                           | `nebraska`             |
| `postgresql.auth.postgresPassword`                       | Password of the superuser `postgresql.auth.username` | `""` (a random password is generated on first install)             |
| `postgresql.image.repository`                             | PostgreSQL image repository                                                                                   | `postgres`             |
| `postgresql.image.registry` | PostgreSQL image registry | `docker.io` |
| `postgresql.image.tag`                                   | PostgreSQL Image tag                                                                                          | `17-bookworm`            |
| `postgresql.image.digest` | PostgreSQL image digest; replaces the tag when set | `""` |
| `postgresql.image.pullPolicy` | PostgreSQL image pull policy | `IfNotPresent` |
| `postgresql.image.pullSecrets`                           | Image pull secrets. Accepts the Bitnami string form (`[regcred]`) and the object form (`[{name: regcred}]`)    | `[]`                   |
| `postgresql.auth.existingSecret`                         | Use an existing secret for the password instead of rendering one (evaluated as a template)                    | `""`                   |
| `postgresql.auth.secretKeys.adminPasswordKey`            | Key inside the secret holding the password                                                                    | `postgres-password`    |
| `postgresql.service.port` | PostgreSQL service and container port | `5432` |
| `postgresql.dataMountPath`                               | Where the data volume is mounted                                                                              | `/var/lib/postgresql/data` |
| `postgresql.dataSubdir`                                  | Subdirectory of the mount used as `PGDATA` (must not be the mount root)                                       | `pgdata`               |
| `postgresql.podSecurityContext`                          | Pod security context; uid/gid 999 matches the default Debian image (use 70 for Alpine tags)                           | see `values.yaml`      |
| `postgresql.containerSecurityContext`                    | Container security context; `readOnlyRootFilesystem` is on by default                                         | see `values.yaml`      |
| `postgresql.resources`                                   | Resource requests/limits for the PostgreSQL container                                                         | `250m` / `256Mi` requests |
| `postgresql.primary.persistence.enabled`                 | Enable persistence using PVC                                                                                  | `false`                |
| `postgresql.primary.persistence.storageClass`            | PVC Storage Class for PostgreSQL volume                                                                       | `nil`                  |
| `postgresql.primary.persistence.accessModes`             | PVC Access Mode for PostgreSQL volume                                                                         | `["ReadWriteOnce"]`    |
| `postgresql.primary.persistence.size`                    | PVC Storage Request for PostgreSQL volume                                                                     | `1Gi`                  |
| `postgresql.primary.persistence.labels` | Labels for the PVC template | `{}` |
| `postgresql.primary.persistence.annotations` | Annotations for the PVC template | `{}` |
| `postgresql.serviceAccount.create`                       | Enable creation of ServiceAccount for PostgreSQL pod                                                          | `true`                 |
| `postgresql.serviceAccount.name` | Name of the ServiceAccount to use or create | `""` |
| `postgresql.serviceAccount.automountServiceAccountToken` | Can be set to false if pods using this serviceAccount do not need to use K8s API                              | `false`                |
| `postgresql.terminationGracePeriodSeconds` | Seconds PostgreSQL gets to shut down before SIGKILL | `90` |
| `postgresql.args` | Arguments for the postgres server, for example `-c` settings | `[]` |
| `postgresql.shmSizeLimit` | Size limit of the memory-backed `/dev/shm` | `256Mi` |
| `postgresql.startupProbe` | First-start budget for the fsGroup chown and initdb | see `values.yaml` |
| `postgresql.extraEnv` | Extra environment variables for the PostgreSQL container | `[]` |
| `postgresql.extraVolumes` | Extra volumes for the PostgreSQL pod | `[]` |
| `postgresql.extraVolumeMounts` | Extra volume mounts for the PostgreSQL container | `[]` |
| `postgresql.extraPodSpec` | Extra PostgreSQL pod spec fields, such as `nodeSelector`, `tolerations` or `initContainers` | `{}` |

This is a deliberately minimal, single-replica PostgreSQL meant to make `helm install` work out of
the box. It does no backups, no failover and no automated major-version upgrades. To use another
database, set `postgresql.enabled: false` and point `config.database.*` at it.
