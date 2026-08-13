# Nebraska Helm Chart

Nebraska is an update manager for Flatcar Container Linux.

## TL;DR

```console
$ helm repo add nebraska https://flatcar.github.io/nebraska
$ helm install my-nebraska nebraska/nebraska
```

## Upgrading to 3.0.0

**Breaking change: the bundled PostgreSQL is no longer the Bitnami subchart.**

Chart 3.0.0 removes the `bitnami/postgresql` dependency and replaces it with a
small PostgreSQL StatefulSet defined inside this chart, running the official
`docker.io/postgres` image.

### Why

Bitnami retired its free catalogue on 2025-08-28. Chart 2.0.0 pinned the image
to `docker.io/bitnamilegacy/postgresql:17.5.0` to keep installs working
(flatcar/nebraska#1227), but the `bitnamilegacy` registry is a frozen archive:
it is never rebuilt, so every PostgreSQL and OS-package CVE published since then
is unpatched in every install running chart defaults (flatcar/nebraska#1574).
The chart also still depended on `https://charts.bitnami.com/bitnami`, a
deprecated endpoint with no committed shutdown date (flatcar/nebraska#1148).
Vendoring the StatefulSet resolves both: the chart now has no external chart
dependency and no Bitnami-family image.

### Does this affect me?

| If you... | Then... |
|-----------|---------|
| set `postgresql.enabled: false` and use an external database | **No action needed.** Nothing in this change touches you. |
| run with the default `postgresql.primary.persistence.enabled: false` | Your database is already ephemeral. Upgrade, and Nebraska will recreate its schema on the new empty database. |
| run with `postgresql.primary.persistence.enabled: true` | **Action required — dump and restore.** See below. The existing PVC cannot be reused as-is. |

### Why persistent data cannot be reused in place

Four independent reasons, any one of which is sufficient:

1. **Different data directory.** Bitnami stored the cluster at
   `/bitnami/postgresql/data`; the official image uses
   `/var/lib/postgresql/data/pgdata`.
2. **Different uid.** Bitnami ran as uid 1001. The Alpine-based official image
   runs as uid 70 (the Debian-based variants use 999). Every file in the data
   directory is owned by the wrong user.
3. **Different C library.** Bitnami images are Debian/glibc; `postgres:17-alpine`
   is musl. Collation ordering differs between the two, and a btree index on
   `text`/`varchar` built under one collation is silently wrong under another —
   queries can fail to find rows that are present. See
   [Locale data changes](https://wiki.postgresql.org/wiki/Locale_data_changes).
   `pg_dump`/restore is explicitly *not* affected by this, which is why it is
   the supported path.
4. **The volume contains no `postgresql.conf`.** Bitnami kept its server config
   inside the *image* at `/opt/bitnami/postgresql/conf/` and passed it with
   `--config-file`, and its entrypoint deleted `postgresql.conf` and
   `pg_hba.conf` from the data directory on every start. The official image
   expects both to live inside `PGDATA`, so pointing it at a Bitnami volume
   fails with `could not access the server configuration file`.

Separately, and regardless of migration path: the Bitnami chart set
`shared_preload_libraries = 'pgaudit'`, and `pgaudit` is not present in the
official image. Because that setting lived in the image's config file rather
than on the volume, it does not block anything — but **if you rely on audit
logging today, it goes away with this upgrade.** Use an image that ships the
extension if you need it.

### Migration (persistence enabled)

Do **not** run `helm upgrade` first — the dump has to come out of the old pod.

```console
# 1. Stop writes.
$ kubectl scale --replicas=0 deployment/my-nebraska

# 2. Dump from the still-running Bitnami pod. Use -U/-d explicitly: a wrong
#    name produces a valid-looking but empty dump.
$ PGPW=$(kubectl get secret my-nebraska-postgresql -o jsonpath='{.data.postgres-password}' | base64 -d)
$ kubectl exec my-nebraska-postgresql-0 --     env PGPASSWORD="$PGPW" pg_dump -U postgres -d nebraska > nebraska.sql
$ grep -c '^COPY public\.' nebraska.sql   # must be >0; `ls -l` cannot detect an empty dump

# 3. Retain the volume BEFORE deleting anything, so a bad dump is survivable.
#    Most StorageClasses use reclaimPolicy: Delete, which destroys the disk with
#    the PVC.
$ PV=$(kubectl get pvc data-my-nebraska-postgresql-0 -o jsonpath='{.spec.volumeName}')
$ kubectl patch pv "$PV" -p '{"spec":{"persistentVolumeReclaimPolicy":"Retain"}}'

# 4. Delete the old StatefulSet and PVC. Note this chart deliberately keeps the
#    StatefulSet's immutable fields (selector, serviceName, volumeClaimTemplates)
#    identical to the Bitnami subchart's, so an in-place `helm upgrade` is NOT
#    rejected by Kubernetes -- it would succeed and silently start an empty
#    database. That is why the chart refuses to render without
#    postgresql.acknowledgeDataDirMigration=true.
$ kubectl delete statefulset my-nebraska-postgresql --cascade=orphan
$ kubectl delete pod my-nebraska-postgresql-0
$ kubectl delete pvc data-my-nebraska-postgresql-0

# 5. Upgrade, keeping Nebraska scaled to zero. Without --set replicaCount=0 the
#    upgrade scales Nebraska straight back up against the new EMPTY database; it
#    would run its migrations and bootstrap rows, and the restore in step 7 would
#    then collide with them.
#
#    PASS YOUR OWN VALUES FILE. `helm upgrade` resets values to chart defaults
#    unless you supply them again, and 3.0.0 defaults persistence to false -- so
#    omitting -f here renders PostgreSQL with no PVC at all and you would restore
#    the dump into an emptyDir that disappears on the next restart.
#    Do NOT use --reuse-values: it would resurrect the 2.0.0 bitnamilegacy image.
$ helm upgrade my-nebraska nebraska/nebraska --version 3.0.0 \
    -f my-values.yaml \
    --set replicaCount=0 \
    --set postgresql.primary.persistence.enabled=true \
    --set postgresql.acknowledgeDataDirMigration=true \
    --wait --timeout 10m

# 6. Wait for the new database to be ready.
$ kubectl wait --for=condition=Ready pod/my-nebraska-postgresql-0 --timeout=300s

# 7. Restore. ON_ERROR_STOP + single-transaction means a partial restore rolls
#    back instead of leaving a half-populated database.
$ kubectl exec -i my-nebraska-postgresql-0 -- \
    env PGPASSWORD="$PGPW" psql -U postgres -d nebraska \
      -v ON_ERROR_STOP=1 --single-transaction < nebraska.sql

# 8. Refresh planner statistics. pg_restore does not do this, and without it the
#    first queries run against empty stats.
$ kubectl exec -i my-nebraska-postgresql-0 -- \
    env PGPASSWORD="$PGPW" psql -U postgres -d nebraska -c 'ANALYZE'

# 9. Verify BEFORE scaling Nebraska back up.
$ kubectl exec -i my-nebraska-postgresql-0 -- \
    env PGPASSWORD="$PGPW" psql -U postgres -d nebraska -tAc \
    'select (select count(*) from application) as apps,
            (select count(*) from groups) as groups,
            (select count(*) from package) as packages'
#    Compare against the same query run before the upgrade.

# 10. Scale Nebraska back up.
$ helm upgrade my-nebraska nebraska/nebraska --version 3.0.0 \
    -f my-values.yaml \
    --set postgresql.primary.persistence.enabled=true \
    --set postgresql.acknowledgeDataDirMigration=true
```

Keep `nebraska.sql` until you have confirmed the new instance is serving
correctly, and only then remove the retained PV.

**If you upgraded by accident and lost your data:** don't delete anything. The
Bitnami cluster is still on the volume in the `data/` directory, untouched —
the new empty cluster was created beside it in `pgdata/`. Run
`helm rollback my-nebraska` immediately and it comes back. `helm rollback`
replays the stored 2.0.0 manifest and does not re-resolve the Bitnami chart
repository, so it works even though that repository is deprecated.

Be aware the symptom is not obvious. The Nebraska Deployment's pod spec is
unchanged between 2.0.0 and 3.0.0, so an in-place upgrade does **not** restart
Nebraska, and Nebraska only runs its schema migrations at process start. The pod
therefore stays `Ready` with its liveness probe green while the API returns
errors (`relation "application" does not exist` in the logs). The empty schema
only materialises the next time Nebraska restarts. Do not read a Ready pod as
confirmation that the upgrade went well — run the verification query.

### GitOps (Argo CD, Flux)

Two things to know:

* The chart preserves an existing password by reading the live Secret. That
  lookup returns nothing during a dry-run or a bare `helm template`, so a
  rendered manifest shows the *values.yaml* password and your tooling will report
  permanent drift on the Secret. Use `postgresql.auth.existingSecret` with a
  secret you manage (SOPS, External Secrets, Sealed Secrets) and the chart will
  not render a Secret at all.
* The `postgresql.acknowledgeDataDirMigration` gate only fires on a real
  `helm upgrade`. Template-rendering workflows never trigger it, so if you are
  moving a persistent install from 2.0.0 to 3.0.0 under GitOps, do the dump and
  restore deliberately — nothing will stop you.

### Values that changed

| 2.0.0 | 3.0.0 | Note |
|-------|-------|------|
| `postgresql.image.repository: bitnamilegacy/postgresql` | `postgresql.image.repository: postgres` | |
| `postgresql.image.tag: 17.5.0` | `postgresql.image.tag: 17-alpine` | Same PostgreSQL major version. |
| *(n/a)* | `postgresql.auth.existingSecret` | New: bring your own secret. |
| *(n/a)* | `postgresql.auth.secretKeys.adminPasswordKey` | New; defaults to the previous key name `postgres-password`. |
| *(n/a)* | `postgresql.dataMountPath`, `postgresql.dataSubdir` | New; see below. |
| *(n/a)* | `postgresql.args` | New: arguments for the postgres server. The only way to set start-time settings such as `wal_level` or `log_connections`, which `ALTER SYSTEM` cannot change. |
| *(n/a)* | `postgresql.image.digest` | New: pin the image by content rather than by tag. |
| *(n/a)* | `postgresql.startupProbe`, `postgresql.shutdownTimeoutSeconds`, `postgresql.shmSizeLimit` | New; see `values.yaml`. |
| *(n/a)* | `postgresql.podSecurityContext`, `postgresql.containerSecurityContext`, `postgresql.resources`, `postgresql.extraEnv`, `postgresql.extraVolumes`, `postgresql.extraVolumeMounts`, `postgresql.nodeSelector`, `postgresql.tolerations`, `postgresql.affinity` | New; previously supplied by the subchart under `postgresql.primary.*`. |
| any other `postgresql.*` key from the Bitnami subchart | **rejected at render time** | The chart reports any value it does not read rather than ignoring it. Keys switched off or left empty (`metrics.enabled: false`, `tls: {}`, `architecture: standalone`) are accepted silently. Keys carrying a real value are reported with the setting they moved to — see below. |

If you vendored the upstream Bitnami `values.yaml` wholesale, expect roughly
twenty reports on the first upgrade. That is intentional: about half of them
(`primary.resources`, `primary.podSecurityContext`, `primary.persistence.mountPath`)
carry real configuration that simply moved, and silently dropping your resource
limits or your `fsGroup` is exactly the failure this guard exists to prevent.
Each message names the replacement key. It is a one-time cleanup, and the values
that genuinely did nothing are already ignored for you.

`postgresql.enabled`, `postgresql.auth.username`, `postgresql.auth.database`,
`postgresql.auth.postgresPassword`, `postgresql.primary.persistence.*`,
`postgresql.serviceAccount.*` and `postgresql.nameOverride` keep their previous
names and meaning. Object names (`<release>-postgresql`,
`<release>-postgresql-hl`), the secret key `postgres-password` and the
`config.database.*` contract are all unchanged, so an external secret manager or
a `config.database.passwordExistingSecret` pointing at them keeps working.

### Other behaviour changes

* **The secret has one key, not two.** The Bitnami subchart emitted both
  `postgres-password` and `password`. Only `postgres-password` is now produced.
  Note the `password` key is actively **removed** from the existing Secret on
  upgrade, not merely left unused — `Secret.data` has no merge patch strategy,
  so Helm nulls the absent key. If you referenced `password` from your own
  manifests, repoint them *before* upgrading.
* **An existing password is preserved.** If the Secret already exists in the
  cluster, its current value wins over `postgresql.auth.postgresPassword`. A
  password you rotated by hand is not reverted to the chart default by a later
  `helm upgrade`.
* **`postgresql.auth.username` now actually works.** Under the Bitnami subchart
  a non-`postgres` username created a *non-superuser* whose password lived under
  the `password` key, while the chart went on connecting with the
  `postgres-password` value — so anything other than `postgres` was broken. With
  the official image `POSTGRES_USER` *is* the superuser initdb creates, and its
  password is the one in `postgres-password`. Note this also means the chart no
  longer offers a way to run Nebraska as a least-privilege, non-superuser role.
* **No `pgaudit`, and no connection logging by default.** Set
  `postgresql.args: [postgres, -c, log_connections=on, -c, log_disconnections=on]`
  if you need an access trail.
* **Sort order may change.** The Bitnami image collated with Debian glibc; the
  Alpine image's musl collation is effectively byte order. `ORDER BY` on text
  columns can return a different order for mixed-case or non-ASCII values. Set
  `postgresql.image.tag` to a Debian variant such as `17-bookworm` (and
  `runAsUser`/`runAsGroup`/`fsGroup` to `999`) if you need glibc collation —
  bookworm carries the same glibc 2.36 as the Bitnami image did.
* **Clean shutdowns.** The pod now has a `preStop` hook running
  `pg_ctl -m fast`. Kubernetes sends SIGTERM, which PostgreSQL reads as "smart
  shutdown" and which makes it wait indefinitely for Nebraska's pooled
  connections to close; previously the pod was SIGKILLed at the end of the grace
  period and the next start did crash recovery.
* **`readOnlyRootFilesystem: true` by default,** with `emptyDir`s at
  `/var/run/postgresql` (the socket directory — PostgreSQL will not start
  without it), `/tmp` and `/dev/shm`. This stops an attacker tampering with the
  binaries; it does not stop code execution, because those mounts are writable
  and a PostgreSQL superuser has `COPY ... TO PROGRAM` regardless.
* **Do not upgrade with `--force-replace`** (Helm 3's `--force`). It deletes and
  recreates the Services, which changes the ClusterIP and breaks every pooled
  connection Nebraska is holding.

### Security notes

* **The superuser password is generated on first install** and preserved across
  upgrades. Retrieve it with:
  ```console
  $ kubectl get secret my-nebraska-postgresql -o jsonpath='{.data.postgres-password}' | base64 -d
  ```
  Chart 2.0.0 shipped a fixed default of `changeIt`; that is gone. Set
  `postgresql.auth.postgresPassword`, or `postgresql.auth.existingSecret`, if you
  manage credentials yourself.
* **No NetworkPolicy is rendered**, matching the Bitnami subchart's default. The
  database is a ClusterIP Service, so anything on the pod network can reach port
  5432 — it just needs the password now, rather than a published default. To
  restrict it, add one through `extraObjects`:
  ```yaml
  extraObjects:
    - apiVersion: networking.k8s.io/v1
      kind: NetworkPolicy
      metadata:
        name: '{{ .Release.Name }}-postgresql'
        namespace: '{{ .Release.Namespace }}'
      spec:
        podSelector:
          matchLabels:
            app.kubernetes.io/name: postgresql
            app.kubernetes.io/instance: '{{ .Release.Name }}'
        policyTypes: [Ingress]
        ingress:
          - from:
              - podSelector:
                  matchLabels:
                    app.kubernetes.io/name: nebraska
                    app.kubernetes.io/instance: '{{ .Release.Name }}'
            ports:
              - port: 5432
                protocol: TCP
  ```
  Remember to allow any backup jobs as well. Requires a CNI that enforces
  NetworkPolicy.
* **No memory limit is set by default**, as in chart 2.0.0 and the Bitnami
  subchart — a wrong limit OOM-kills a database mid-transaction, so the chart
  will not guess one. Set `postgresql.resources.limits` once you know your
  working set. `/dev/shm` is bounded at 256Mi by default, which removes the
  node-pressure vector that an unbounded memory-backed volume would create.
* Traffic to the database is unencrypted by default (`sslmode=disable`). To
  enable TLS, mount a certificate with `postgresql.extraVolumes` and set
  `postgresql.args: [postgres, -c, ssl=on, -c, ssl_cert_file=..., -c, ssl_key_file=...]`,
  then set `config.database.sslMode: verify-full`.
* `postgresql.image.tag` is a floating tag pulled with `IfNotPresent`, so a node
  that has already cached it will not pick up a rebuilt image. Set
  `postgresql.image.digest` to pin by content, and keep it updated.

### Is the bundled database production-ready?

No, and it is not meant to be. It is a single replica with no backups, no
failover and no automated major-version upgrades — the same scope the Bitnami
subchart had in this chart. It exists so that `helm install` produces a working
Nebraska.

For anything you care about, set `postgresql.enabled: false` and point
`config.database.*` at a database you operate, or at an operator such as
[CloudNativePG](https://cloudnative-pg.io/).

That applies with particular force to the distributed topology in
[RFC #1375](https://github.com/flatcar/nebraska/issues/1375). That design gives
each region its own **writable** database kept in sync by one-way *logical*
replication, with least-privilege roles separating the admin and runtime write
surfaces. The bundled StatefulSet can technically participate — set
`postgresql.args: [postgres, -c, wal_level=logical]` and persistence on — but it
provisions no roles, manages no publications or subscriptions, and defaults to
ephemeral storage, which would destroy replication slots on every pod
replacement. An operator is the right tool there. Streaming replication and read
replicas are deliberately not offered: every Omaha check-in writes, and Nebraska
opens a single connection pool, so a read-only standby has nowhere to send
traffic.

### Backups

Anything that backs up over the network — `pg_dump`/`pg_dumpall` against the
`<release>-postgresql` Service — is unaffected. The wire protocol, port, Service
name, database name and credentials are all unchanged.

Two things do change:

* **Volume-snapshot backups are not portable across this upgrade.** A snapshot
  taken from the Bitnami PVC cannot be restored into the new StatefulSet, for
  the same four reasons listed above. Take a logical dump before upgrading, and
  treat any pre-upgrade snapshots as restorable only onto chart 2.0.0.
* **`kubectl exec ... pg_dumpall` without credentials still works, but for a
  different reason.** The official image's `initdb` leaves `local` connections
  on `trust`, and the container runs as the `postgres` OS user, so a dump over
  the unix socket needs no password. This depends on the `/var/run/postgresql`
  mount being present; if you override `postgresql.extraVolumeMounts` in a way
  that removes it, socket connections — and the server itself — stop working.

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
-    tag: 17-alpine
+    tag: 18-alpine
```
   Note that PostgreSQL 18 relocates both `PGDATA` and the image's declared
   volume — and note the two must move together. The image declares a `VOLUME`,
   and if the PVC is mounted at an *ancestor* of it the runtime mounts an empty
   volume over the top and everything underneath becomes invisible inside the
   container. For 17 the VOLUME is `/var/lib/postgresql/data`, so mounting at
   `/var/lib/postgresql` silently hides your data; for 18 it is
   `/var/lib/postgresql`, so that mount point becomes the correct one. Set
   `postgresql.dataMountPath: /var/lib/postgresql` and
   `postgresql.dataSubdir: 18/docker` to match.

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
| `config.caFile`                                       | Path to a PEM-encoded CA certificate file to trust for TLS verification (additive to system CAs, used for OIDC and syncer) | `nil`  |
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
| `config.database.host`                                | The host name of the database server                                                                                                 | `""` (use the PostgreSQL bundled with this chart)                             |
| `config.database.port`                                | The port number the database server is listening on                                                                                  | `5432`                                                                  |
| `config.database.sslMode`                             | The mode of the database connection                                                                                                  | `disable`                                                               |
| `config.database.dbname`                              | The database name                                                                                                                    | `{{ .Values.postgresql.auth.database }}` (evaluated as a template)      |
| `config.database.username`                            | PostgreSQL user                                                                                                                      | `{{ .Values.postgresql.auth.username }}` (evaluated as a template)                                    |
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
| `postgresql.enabled`                                     | Deploy the PostgreSQL StatefulSet bundled with this chart                                                     | `true`                 |
| `postgresql.auth.database`                               | PostgreSQL database                                                                                           | `nebraska`             |
| `postgresql.auth.postgresPassword`                       | PostgreSQL password of user "postgres" **Recommended to change it to something secure for security reasons.** | `changeIt`             |
| `postgresql.image.repository`                             | PostgreSQL image repository                                                                                   | `postgres`             |
| `postgresql.image.tag`                                   | PostgreSQL Image tag                                                                                          | `17-alpine`            |
| `postgresql.auth.existingSecret`                         | Use an existing secret for the password instead of rendering one (evaluated as a template)                    | `""`                   |
| `postgresql.auth.secretKeys.adminPasswordKey`            | Key inside the secret holding the password                                                                    | `postgres-password`    |
| `postgresql.dataMountPath`                               | Where the data volume is mounted                                                                              | `/var/lib/postgresql/data` |
| `postgresql.dataSubdir`                                  | Subdirectory of the mount used as `PGDATA` (must not be the mount root)                                       | `pgdata`               |
| `postgresql.podSecurityContext`                          | Pod security context; uid/gid 70 matches the Alpine image (Debian variants use 999)                           | see `values.yaml`      |
| `postgresql.containerSecurityContext`                    | Container security context; `readOnlyRootFilesystem` is on by default                                         | see `values.yaml`      |
| `postgresql.resources`                                   | Resource requests/limits for the PostgreSQL container                                                         | `250m` / `256Mi` requests |
| `postgresql.primary.persistence.enabled`                 | Enable persistence using PVC                                                                                  | `false`                |
| `postgresql.primary.persistence.storageClass`            | PVC Storage Class for PostgreSQL volume                                                                       | `nil`                  |
| `postgresql.primary.persistence.accessModes`             | PVC Access Mode for PostgreSQL volume                                                                         | `["ReadWriteOnce"]`    |
| `postgresql.primary.persistence.size`                    | PVC Storage Request for PostgreSQL volume                                                                     | `1Gi`                  |
| `postgresql.serviceAccount.create`                       | Enable creation of ServiceAccount for PostgreSQL pod                                                          | `true`                 |
| `postgresql.serviceAccount.automountServiceAccountToken` | Can be set to false if pods using this serviceAccount do not need to use K8s API                              | `false`                |

This is a deliberately minimal, single-replica PostgreSQL meant to make `helm install` work out of
the box. It does no backups, no failover and no automated major-version upgrades. For production,
set `postgresql.enabled: false` and point `config.database.*` at a database you operate, or at an
operator such as [CloudNativePG](https://cloudnative-pg.io/).
