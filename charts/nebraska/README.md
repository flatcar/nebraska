# Nebraska Helm Chart

Nebraska is an update manager for Flatcar Container Linux.

## TL;DR

```console
$ helm repo add nebraska https://flatcar.github.io/nebraska
$ helm install my-nebraska nebraska/nebraska
```

## Upgrading to 4.0.0

**Breaking change: the bundled PostgreSQL is no longer the Bitnami subchart.**

Chart 4.0.0 removes the `bitnami/postgresql` dependency and replaces it with a
small PostgreSQL StatefulSet defined inside this chart, running the official
`docker.io/postgres` image.

### Why

Bitnami retired its free catalogue on 2025-08-28. Chart 1.7.0 pinned the image
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
| set `postgresql.enabled: false` and use an external database | **Almost nothing.** The database side does not touch you. But if your values still set old Bitnami-only `postgresql.*` keys (for example `postgresql.postgresqlPassword`), the chart stops and lists them; delete them. They did nothing before either. |
| run with the default `postgresql.primary.persistence.enabled: false` | **No action needed.** Your database is already ephemeral. The bundled PostgreSQL comes back empty and the chart rolls Nebraska so it recreates its schema unattended. |
| run with `postgresql.primary.persistence.enabled: true` | **Action required.** Same PostgreSQL major version: reuse the volume in place, see below. Otherwise: dump and restore. |

### Reusing the volume in place, or dump and restore

Both chart versions use the same PVC (`data-<release>-postgresql-0`). Reusing
it in place needs the **same PostgreSQL major version**: chart 3.0.0 defaulted
to `bitnamilegacy/postgresql:17.5.0` and this chart to `postgres:17-bookworm`.
If you pinned another major version, use dump and restore.

Three things differ, and the in-place values below handle all of them:

1. **Data directory.** Bitnami used `/bitnami/postgresql/data`; the default here
   is `/var/lib/postgresql/data/pgdata`. Set `postgresql.dataMountPath` and
   `postgresql.dataSubdir`.
2. **uid.** Bitnami ran as 1001; this image runs as 999. Set
   `postgresql.podSecurityContext`.
3. **No `postgresql.conf` on the volume.** Bitnami kept its config in the
   image, while the official image expects it in `PGDATA`. A one-time init
   container writes a minimal `postgresql.conf` and `pg_hba.conf`.

This was tested end to end on kind: PostgreSQL recovered from the Bitnami WAL
and served the existing database. Dump and restore is the fallback, and the only
path across major versions.

### In-place upgrade (same major version, tested)

No dump, no restore, no PVC deletion. The StatefulSet's name, selector, service
name and claim name are identical, so the same volume is attached; three value
changes make the old data directory usable. Set these in your values file and
run `helm upgrade` with `postgresql.acknowledgeDataDirMigration=true`:

```yaml
postgresql:
  dataMountPath: /bitnami/postgresql   # mount the PVC where Bitnami mounted it
  dataSubdir: data                     # PGDATA = /bitnami/postgresql/data
  podSecurityContext:                  # every file on the volume is uid/gid 1001
    runAsNonRoot: true
    runAsUser: 1001
    runAsGroup: 1001
    fsGroup: 1001
    fsGroupChangePolicy: OnRootMismatch
    seccompProfile: { type: RuntimeDefault }
  extraPodSpec:
    initContainers:                    # one-time config bootstrap, see reason 3
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

Caveats, all verified in the same run:

* **Snapshot or retain the PV first.** In-place means PostgreSQL writes to the
  original directory, the safety net is the volume, not an untouched `data/`.
* Expect a short WAL redo on first start (the 3.0.0 pod had no preStop hook, so
  it was SIGKILLed), normal crash recovery, not corruption.
* The first start may log `chmod: changing permissions of '/var/run/postgresql':
  Operation not permitted`, cosmetic, the server continues.
* **Check how your passwords are hashed before you use the `pg_hba.conf` above.**
  It ends with `scram-sha-256`, which is correct for PostgreSQL 14 and newer, so
  it matches a 17.5 Bitnami install. If your cluster was upgraded from an older
  major version, the stored passwords may still be md5, and then every login
  fails with a confusing error. To check, before you start:

  ```console
  $ kubectl exec <postgres-pod> -- psql -U postgres -tAc \
      "select distinct substring(rolpassword for 4) from pg_authid where rolpassword is not null"
  ```

  `SCRAM` means you can keep the line as it is. `md5` means change that last
  line to `md5`, or reset the password after the upgrade.
* This recreates no config you had under Bitnami beyond the defaults above. If
  you relied on custom Bitnami settings (`shared_preload_libraries`, custom
  `pg_hba` rules), port them yourself. Note the official image does not ship
  `pgaudit`.
* After it works you may drop the init container on a later change; the config
  files now live in PGDATA, where the official image expects them. Keep
  `dataMountPath`, `dataSubdir` and `podSecurityContext` for as long as you use
  this volume. Without them PostgreSQL points at an empty directory, or runs as
  uid 999 against files owned by 1001 and refuses to start. The chart refuses
  the first case on upgrade and tells you to set the values back.

The official image does not ship `pgaudit`, which the Bitnami chart preloaded.
If you rely on audit logging, see "Other behaviour changes" below.

### Migration (persistence enabled)

Do **not** run `helm upgrade` first, the dump has to come out of the old pod.

The commands below use the current namespace. If the release lives in another
namespace, add `-n <namespace>` to every `kubectl` and `helm` command.

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
#    rejected by Kubernetes, it would succeed and silently start an empty
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
#    unless you supply them again, and 4.0.0 defaults persistence to false, so
#    omitting -f here renders PostgreSQL with no PVC at all and you would restore
#    the dump into an emptyDir that disappears on the next restart.
#    Do NOT use --reuse-values: it would resurrect the 3.0.0 bitnamilegacy image.
$ helm upgrade my-nebraska nebraska/nebraska --version 4.0.0 \
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

# 8. Refresh planner statistics. A plain SQL restore does not do this, and
#    without it the first queries run against empty stats.
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
$ helm upgrade my-nebraska nebraska/nebraska --version 4.0.0 \
    -f my-values.yaml \
    --set postgresql.primary.persistence.enabled=true \
    --set postgresql.acknowledgeDataDirMigration=true
```

Keep `nebraska.sql` until you have confirmed the new instance is serving
correctly, and only then remove the retained PV.

**If you upgraded by accident and lost your data:** don't delete anything. The
Bitnami cluster is still on the volume in the `data/` directory, untouched. The
new empty cluster was created next to it in `pgdata/`. Run
`helm rollback my-nebraska` immediately and it comes back. `helm rollback`
replays the stored 3.0.0 manifest and does not re-resolve the Bitnami chart
repository, so it works even though that repository is deprecated.

> **Helm 4:** if the Secret was ever changed outside Helm (for example with
> `kubectl patch`), the rollback fails with `conflict with "kubectl-patch"`.
> Retry with `helm rollback my-nebraska 1 --force-conflicts`. Helm 3 is not
> affected.

A Nebraska pod running against an empty database still shows `Ready`, because
its probes do not check the database; the API then returns errors such as
`relation "application" does not exist`. So run the verification query; do not
rely on a Ready pod. (The chart restarts Nebraska once on the upgrade, through
the `nebraska.flatcar.org/bundled-db-generation` annotation, so it runs its
schema migrations against the new database.)

### GitOps (Argo CD, Flux)

Two things to know:

* The chart preserves an existing password by reading the live Secret. That
  lookup returns nothing during a dry-run or a bare `helm template`, so a
  rendered manifest shows a *freshly generated* password on every render. This
  is worse than cosmetic drift: if your tooling *applies* that manifest, the
  cluster Secret is overwritten while PostgreSQL keeps the password it was
  initialised with, the next pod restart fails authentication. Use
  `postgresql.auth.existingSecret` with a secret you manage (SOPS, External
  Secrets, Sealed Secrets) and the chart will not render a Secret at all.
  Renaming `secretKeys.adminPasswordKey` on an existing install has the same
  lockout effect, because the lookup only reads the new key.
* The `postgresql.acknowledgeDataDirMigration` gate only fires on a real
  `helm upgrade`. Template-rendering workflows never trigger it, so if you are
  moving a persistent install from 3.0.0 to 4.0.0 under GitOps, do the dump and
  restore deliberately, nothing will stop you. The gate does not fire again
  once the live StatefulSet was made by this chart with the same data path, so
  later upgrades need no flag. It does fire if you change `dataMountPath` or
  `dataSubdir`, because that also points PostgreSQL at an empty directory.

### Values that changed

| 3.0.0 | 4.0.0 | Note |
|-------|-------|------|
| `postgresql.image.repository: bitnamilegacy/postgresql` | `postgresql.image.repository: postgres` | |
| `postgresql.image.tag: 17.5.0` | `postgresql.image.tag: 17-bookworm` | Same PostgreSQL major version, and the same glibc as the Bitnami image, so collation is unchanged. |
| *(n/a)* | `postgresql.auth.existingSecret` | New: bring your own secret. |
| *(n/a)* | `postgresql.auth.secretKeys.adminPasswordKey` | New; defaults to the previous key name `postgres-password`. |
| *(n/a)* | `postgresql.dataMountPath`, `postgresql.dataSubdir` | New; see below. |
| *(n/a)* | `postgresql.args` | New: arguments for the postgres server. The only declarative, first-boot way to set start-time settings such as `wal_level` or `log_connections` (`ALTER SYSTEM` works but needs pod access and a restart). |
| *(n/a)* | `postgresql.image.digest` | New: pin the image by content rather than by tag. |
| *(n/a)* | `postgresql.startupProbe`, `postgresql.shmSizeLimit`, `postgresql.extraPodSpec` | New; see `values.yaml`. |
| *(n/a)* | `postgresql.podSecurityContext`, `postgresql.containerSecurityContext`, `postgresql.resources`, `postgresql.extraEnv`, `postgresql.extraVolumes`, `postgresql.extraVolumeMounts` | New; previously supplied by the subchart under `postgresql.primary.*`. Scheduling fields (`nodeSelector`, `tolerations`, `affinity`, `priorityClassName`) are set through `postgresql.extraPodSpec` rather than one key each. |
| any other `postgresql.*` key from the Bitnami subchart | **rejected at render time** | The chart reports any value it does not read rather than ignoring it. Keys switched off or left empty (`metrics.enabled: false`, `tls: {}`, `architecture: standalone`) are accepted silently. Keys carrying a real value are reported with the setting they moved to, where there is one. |

If you copied the whole upstream Bitnami `values.yaml`, expect a longer list,
because it sets many keys to non-empty defaults (probe timings, secret key
names, the `image:` block). Delete those keys; this chart does not read them.
A key you really changed, say `primary.resources`, is reported with the key it
moved to, because silently dropping your resource limits is exactly the failure
this guard exists to prevent.

Reported settings that are easy to miss, because they sit under keys this chart
*does* read:

* Typos and Bitnami-only keys under `postgresql.image.*` (for example a
  misspelt `digest`, which would otherwise fall back to the floating tag) and
  under `postgresql.service.*` (only `port` is read; the Services are always
  ClusterIP).
* `postgresql.primary.podSecurityContext` / `containerSecurityContext` with
  `enabled: false`. Move it to `postgresql.podSecurityContext.enabled: false` /
  `postgresql.containerSecurityContext.enabled: false`, which, as before, render
  no security context, for platforms that assign uids themselves.
* `postgresql.serviceAccount.annotations`, not applied to the PostgreSQL
  ServiceAccount; use the top-level `extraAnnotations`, which reach every object.
* `postgresql.primary.livenessProbe` / `readinessProbe`, including an explicit
  `enabled: false`. Probes are fixed by this chart; `postgresql.startupProbe`
  tunes the first-start budget.
* `postgresql.extraPodSpec.containers` / `volumes`, these would duplicate a key
  the chart renders itself and produce an invalid pod spec. Use `postgresql.extraEnv`,
  `extraVolumes` and `extraVolumeMounts`, or `extraPodSpec.initContainers`.

`postgresql.enabled`, `postgresql.auth.username`, `postgresql.auth.database`,
`postgresql.auth.postgresPassword`, `postgresql.primary.persistence.*`,
`postgresql.serviceAccount.*` and `postgresql.nameOverride` keep their previous
names and meaning. Object names (`<release>-postgresql`,
`<release>-postgresql-hl`), the secret key `postgres-password` and the
`config.database.*` contract are all unchanged, so an external secret manager or
a `config.database.passwordExistingSecret` pointing at them keeps working.

### Other behaviour changes

* **The secret has one key, not two.** Only `postgres-password` is produced.
  Helm **removes** the old `password` key from the existing Secret on upgrade,
  so repoint anything that reads it *before* upgrading.
* **An existing password is preserved.** If the Secret already exists in the
  cluster, its current value wins over `postgresql.auth.postgresPassword`. A
  password you rotated by hand is not reverted to the chart default by a later
  `helm upgrade`.
* **`postgresql.auth.username` now works.** Under Bitnami, a name other than
  `postgres` created a non-superuser while the chart kept using the superuser
  password, so it was broken. Now `POSTGRES_USER` is the superuser, with the
  password in `postgres-password`. There is no way to run Nebraska as a
  non-superuser role with the bundled database.
* **No `pgaudit`.** The official image does not ship it. For connection
  logging, set
  `postgresql.args: [postgres, -c, log_connections=on, -c, log_disconnections=on]`.
* **Sort order is unchanged,** because `17-bookworm` has the same glibc 2.36 as
  the Bitnami image. A trixie tag (`17`, `17-trixie`, newer glibc) changes it,
  so run `REINDEX DATABASE` and `ALTER DATABASE ... REFRESH COLLATION VERSION`
  after switching. An Alpine tag (musl) changes it too, and also needs uid/gid 70.
* **Clean shutdowns.** A `preStop` hook runs `pg_ctl -m fast`. Before, the pod
  waited for Nebraska's pooled connections, was SIGKILLed, and the next start
  did crash recovery.
* **`readOnlyRootFilesystem: true`,** with `emptyDir`s at `/var/run/postgresql`,
  `/tmp` and `/dev/shm`.
* **The metrics exporter and the volumePermissions init container are gone,**
  and setting either is refused. Run postgres-exporter as its own Deployment
  against the Service (for example through `extraObjects`). If your storage
  ignores `fsGroup`, use the init container example in `values.yaml`.
* **Do not upgrade with `--force-replace`** (Helm 3's `--force`). It deletes and
  recreates the Services, which changes the ClusterIP and breaks every pooled
  connection Nebraska is holding.

### Security notes

* **The superuser password is generated on first install** and preserved across
  upgrades. Retrieve it with:
  ```console
  $ kubectl get secret my-nebraska-postgresql -o jsonpath='{.data.postgres-password}' | base64 -d
  ```
  Chart 3.0.0 used the fixed `changeIt`. On an upgrade from 3.0.0 without
  persistence, the database starts empty anyway, so a stored `changeIt` is
  replaced with a generated password. The upgrade notes say so and print the
  command above; update any tool outside Nebraska that used `changeIt`, such as
  a backup script. With persistence the password is kept, because the database
  still uses it, and the upgrade notes show how to change it. Set
  `postgresql.auth.postgresPassword` or `postgresql.auth.existingSecret` if you
  manage credentials yourself.
* **No NetworkPolicy is rendered**, as in the Bitnami subchart, so anything on
  the pod network can reach port 5432 (it now needs the generated password). Add
  your own NetworkPolicy through `extraObjects` if you need one.
* **No memory limit is set by default**, as in chart 3.0.0. Set
  `postgresql.resources.limits` once you know your working set.
* Traffic to the database is unencrypted by default (`sslmode=disable`). To
  enable TLS, mount a certificate with `postgresql.extraVolumes` and set
  `postgresql.args: [postgres, -c, ssl=on, -c, ssl_cert_file=..., -c, ssl_key_file=...]`,
  then set `config.database.sslMode: verify-full`.
* `postgresql.image.tag` is a floating tag pulled with `IfNotPresent`, so a node
  that has already cached it will not pick up a rebuilt image. Set
  `postgresql.image.digest` to pin by content, and keep it updated.

### Is the bundled database production-ready?

No. It is a single replica with no backups, no failover and no automated
major-version upgrades, like the Bitnami subchart was. For anything you care
about, set `postgresql.enabled: false` and point `config.database.*` at a
database you operate, or use an operator such as
[CloudNativePG](https://cloudnative-pg.io/). That includes the distributed
topology in [RFC #1375](https://github.com/flatcar/nebraska/issues/1375).

### Backups

Anything that backs up over the network, `pg_dump`/`pg_dumpall` against the
`<release>-postgresql` Service, is unaffected. The wire protocol, port, Service
name, database name and credentials are all unchanged.

Three things do change:

* **Volume snapshots taken before the upgrade need the in-place values.** With
  the default values, a snapshot of the Bitnami PVC looks empty to the new
  StatefulSet, for the three reasons listed above. Restore it onto chart 3.0.0,
  or onto 4.0.0 with the values from "In-place upgrade" (same PostgreSQL major
  version only). Take a logical dump before upgrading as well.
* **Backup scripts that source the Bitnami environment break.** The official
  image has no `POSTGRESQL_*` variables (`POSTGRESQL_PASSWORD`,
  `POSTGRESQL_DATABASE`, ...) and no `/opt/bitnami/scripts/*`. Anything that
  `exec`s into the pod and relies on those needs rewriting against
  `POSTGRES_*`, or better, pointed at the Service over the network, which is
  unaffected.
* **`kubectl exec ... pg_dumpall` without a password still works,** because the
  official image trusts local socket connections. It needs the
  `/var/run/postgresql` mount, which the chart always adds.

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

When there is a major upgrade of PostgreSQL, a manual intervention might be required with a downtime. It is possible to automate things with operators, but here's a simple example:

1. Scale down Nebraska deployment:
```
$ kubectl scale --replicas=0 deployment/nebraska
deployment.apps/nebraska scaled
```

2. Backup PostgreSQL data:
```
$ kubectl exec -i pod/nebraska-postgresql-0 -- pg_dumpall > backup.sql
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
   **The mount path must move with the major version.** PostgreSQL 18 relocates
   both `PGDATA` and the image's declared `VOLUME`:

   | major | set `dataMountPath` | set `dataSubdir` |
   |-------|---------------------|------------------|
   | 17    | `/var/lib/postgresql/data` | `pgdata` |
   | 18    | `/var/lib/postgresql`      | `18/docker` |

   Get this wrong and the failure is silent: mounting the PVC *above* the
   image's `VOLUME` makes the runtime lay an empty volume over the top, so
   everything already on your disk becomes invisible inside the container. The
   chart refuses the combination rather than letting it happen, but only when
   it can read the major version from the tag.

   With persistence enabled, the new data path also triggers the chart's
   data-directory check, because PostgreSQL will start with an empty directory.
   That is expected here, since you restore the dump in step 7. Add
   `--set postgresql.acknowledgeDataDirMigration=true` to this upgrade.

6. Apply the changes and scale up Nebraska statefulset to its original value

7. Inject the backup and assert that everything looks good in the database:
```
$ kubectl exec -i pod/nebraska-postgresql-0 -- psql < backup.sql
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
| `config.database.passwordExistingSecret.key`          | Key inside the existing secret containing the PostgreSQL user password                                                               | `postgres-password`                                                     |
| `extraArgs`                                           | Extra arguments to pass to Nebraska binary                                                                                           | `[]`                                                                    |
| `extraEnvVars`                                        | Any extra environment variables you would like to pass on to the pod                                                                 | `{ "TZ": "UTC" }`                                                       |
| `extraEnv`                                        | Any extra environment variables in the form of env spec to pass into the deployment pod                                                                 | `[]`                                                       |

### Bundled PostgreSQL parameters

| Parameter                                                | Description                                                                                                   | Default                |
|----------------------------------------------------------|---------------------------------------------------------------------------------------------------------------|------------------------|
| `postgresql.enabled`                                     | Deploy the PostgreSQL StatefulSet bundled with this chart                                                     | `true`                 |
| `postgresql.auth.database`                               | PostgreSQL database                                                                                           | `nebraska`             |
| `postgresql.auth.postgresPassword`                       | PostgreSQL password of user "postgres" | `""` (a random password is generated on first install)             |
| `postgresql.image.repository`                             | PostgreSQL image repository                                                                                   | `postgres`             |
| `postgresql.image.tag`                                   | PostgreSQL Image tag                                                                                          | `17-bookworm`            |
| `postgresql.image.pullSecrets`                           | Image pull secrets. Accepts the Bitnami string form (`[regcred]`) and the object form (`[{name: regcred}]`)    | `[]`                   |
| `postgresql.auth.existingSecret`                         | Use an existing secret for the password instead of rendering one (evaluated as a template)                    | `""`                   |
| `postgresql.auth.secretKeys.adminPasswordKey`            | Key inside the secret holding the password                                                                    | `postgres-password`    |
| `postgresql.dataMountPath`                               | Where the data volume is mounted                                                                              | `/var/lib/postgresql/data` |
| `postgresql.dataSubdir`                                  | Subdirectory of the mount used as `PGDATA` (must not be the mount root)                                       | `pgdata`               |
| `postgresql.podSecurityContext`                          | Pod security context; uid/gid 999 matches the default Debian image (use 70 for Alpine tags)                           | see `values.yaml`      |
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
