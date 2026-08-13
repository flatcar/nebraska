{{/*
Refuse to upgrade a persistent install onto the new data directory layout
without an explicit acknowledgement.

This is the most important guard in the chart. The Bitnami subchart mounted the
PVC at /bitnami/postgresql with PGDATA=/bitnami/postgresql/data; this chart
mounts the same PVC at /var/lib/postgresql/data with PGDATA=.../data/pgdata.

Every field the Kubernetes StatefulSet controller treats as immutable --
spec.selector, spec.serviceName, spec.volumeClaimTemplates -- is deliberately
byte-identical to what the subchart emitted, so `helm upgrade` is ACCEPTED and
returns 0. The pod restarts, finds no PG_VERSION at the new PGDATA, runs initdb
into a fresh sibling directory, and comes up as an empty cluster. Nebraska then
recreates its schema. Every application, group, channel and rollout disappears
from the user's point of view while the old cluster sits untouched next to it.
No error is produced anywhere.

Prose in the README does not defend against this: nobody reads a README during
a Renovate bump. So the render fails until the operator says they have dealt
with the data.

The old cluster is still on the volume, so an accidental upgrade is recoverable
with `helm rollback` -- that is worth knowing and is in the message.
*/}}
{{- define "nebraska.postgresql.validateDataDirMigration" -}}
{{- $pg := .Values.postgresql | default dict -}}
{{- /* Only an UPGRADE can destroy data this way. A fresh install has no prior
       cluster on the volume, so failing there would be pure friction for every
       new user who wants persistence. Note this means template-rendering
       workflows (Argo/Flux, `helm template`) never see the gate, because
       .Release.IsUpgrade is false there -- those users get the README and
       NOTES.txt instead. */ -}}
{{- if and .Release.IsUpgrade $pg.enabled ((($pg.primary | default dict).persistence | default dict).enabled) -}}
{{- if not $pg.acknowledgeDataDirMigration -}}
{{- fail "\n\nSTOP -- this upgrade would silently discard your database.\n\nChart 3.0.0 replaced the Bitnami postgresql subchart with the official postgres\nimage, which stores data at a different path inside the same volume:\n\n  chart 2.0.0 (Bitnami):  PVC mounted at /bitnami/postgresql       PGDATA=/bitnami/postgresql/data\n  chart 3.0.0 (official): PVC mounted at /var/lib/postgresql/data  PGDATA=/var/lib/postgresql/data/pgdata\n\nNothing in Kubernetes rejects this change, so `helm upgrade` would SUCCEED and\nPostgreSQL would initialise a brand-new empty database alongside your existing\none. Nebraska would come up looking healthy with no applications, groups or\nrollouts, and no error would be reported.\n\nYou have persistence enabled, so you must choose:\n\n  1. Migrate the data (dump/restore). Follow \"Upgrading to 3.0.0\" in the chart\n     README, then re-run with:\n         --set postgresql.acknowledgeDataDirMigration=true\n\n  2. Deliberately start from an empty database (fine for dev/test). Same flag:\n         --set postgresql.acknowledgeDataDirMigration=true\n\n  3. Stay on chart 2.0.0 for now.\n\nIf you already ran this upgrade by accident: your old cluster is still present\non the volume, untouched, in the `data/` directory. Run `helm rollback` NOW,\nbefore deleting any PVC, and it will come back.\n" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Report values that this chart does not honour.

DENY-UNKNOWN, not allow-known. The first version of this listed the Bitnami keys
someone remembered, and an adversarial review found it missed roughly 55 of them
-- including `primary.resources` (memory limits silently vanish),
`primary.podSecurityContext` (a user's fsGroup silently replaced),
`primary.persistence.existingClaim` (silently ignored, so a fresh EMPTY volume is
provisioned) and the whole `global.*` tree. An allowlist can only ever catch the
keys its author thought of, which is the wrong failure mode for a data-bearing
chart. So: enumerate the keys this chart READS, and report everything else.

DEFAULT-AWARE. A very common pattern is to vendor the upstream subchart's
values.yaml wholesale and edit a few lines. Such a file carries dozens of keys at
their Bitnami defaults -- `architecture: standalone`, `metrics.enabled: false`,
`tls.enabled: false`. Failing on those would tell users that features they never
turned on have been removed, which is noise, and would make a routine upgrade
impossible for exactly the people we least want to break. So an inert value
(a feature block that is disabled, or a setting already at the Bitnami default)
is ignored; only values that would actually have changed behaviour are reported.
*/}}

{{/* True when a removed value would not have done anything anyway.

     Recursive, because a vendored Bitnami values.yaml carries whole nested
     blocks sitting at their defaults. `metrics:` is not just `enabled: false` --
     it is fifty lines of image tags, probe timings and resource stanzas hanging
     off it. Reporting every one of those would make the upgrade unusable for
     precisely the people the README promises it will work for.

     A value is inert when it is empty, false, an empty collection, a feature
     block that is switched off, or a map whose every member is itself inert. */}}
{{- define "nebraska.postgresql.isInertValue" -}}
{{- $v := .value -}}
{{- $k := .key -}}
{{- if kindIs "invalid" $v -}}inert
{{- else if kindIs "bool" $v -}}
  {{- if not $v -}}inert{{- end -}}
{{- else if kindIs "string" $v -}}
  {{- if eq $v "" -}}inert
  {{- else if and (eq $k "architecture") (eq $v "standalone") -}}inert{{- end -}}
{{- else if kindIs "slice" $v -}}
  {{- if not $v -}}inert{{- end -}}
{{- else if kindIs "map" $v -}}
  {{- if not $v -}}inert
  {{- else if and (hasKey $v "enabled") (not $v.enabled) -}}inert
  {{- else if and (hasKey $v "create") (not $v.create) -}}inert
  {{- else -}}
    {{- $allInert := true -}}
    {{- range $ck, $cv := $v -}}
      {{- if not (include "nebraska.postgresql.isInertValue" (dict "key" $ck "value" $cv)) -}}
        {{- $allInert = false -}}
      {{- end -}}
    {{- end -}}
    {{- if $allInert -}}inert{{- end -}}
  {{- end -}}
{{- end -}}
{{- end -}}

{{- define "nebraska.postgresql.validateUnknownValues" -}}
{{- $pg := .Values.postgresql | default dict -}}
{{- $found := list -}}

{{/* Keys this chart actually reads. Anything else is reported. */}}
{{- $known := list
  "enabled" "acknowledgeDataDirMigration" "nameOverride" "fullnameOverride"
  "extraPodSpec"
  "auth" "image" "service" "dataMountPath" "dataSubdir" "primary"
  "serviceAccount" "podSecurityContext" "containerSecurityContext"
  "terminationGracePeriodSeconds"
  "resources"
  "extraEnv" "extraVolumes" "extraVolumeMounts"
  "args" "shmSizeLimit" "startupProbe"
-}}

{{/* Specific guidance where a generic message would not be enough. */}}
{{- $guide := dict
  "architecture"                     "streaming standbys are not provided by this chart. Note Nebraska cannot use a read-only standby anyway -- every Omaha check-in writes, and it opens a single DSN. If you are after the distributed topology in RFC #1375, that uses one-way LOGICAL replication, which this chart can do: set postgresql.args to include -c wal_level=logical. For managed HA, use an operator such as CloudNativePG with postgresql.enabled=false."
  "replication"                      "streaming replication is not provided by this chart. Logical replication is reachable via postgresql.args (-c wal_level=logical); for managed HA use an operator with postgresql.enabled=false."
  "readReplicas"                     "read replicas are not provided by this chart, and Nebraska opens a single DSN so it has no read/write split to use them."
  "metrics"                          "the postgres-exporter sidecar, ServiceMonitor and PrometheusRule are gone. Run the exporter via postgresql.sidecars, and supply the ServiceMonitor through the top-level extraObjects."
  "tls"                              "in-chart TLS termination is gone. Set server options via postgresql.args (-c ssl=on -c ssl_cert_file=...) with the cert supplied through postgresql.extraVolumes, or terminate at a proxy/mesh."
  "ldap"                             "LDAP auth is gone. It was never used by Nebraska."
  "audit"                            "pgAuditLog/pgAuditLogCatalog need the pgaudit extension, which the official image does not ship. The other audit settings (logConnections, logDisconnections, logHostname, logLinePrefix, logTimezone, clientMinMessages) are plain PostgreSQL settings -- set them via postgresql.args, e.g. -c log_connections=on."
  "postgresqlSharedPreloadLibraries" "set this via postgresql.args (-c shared_preload_libraries=...). Note the official image does not ship pgaudit."
  "postgresqlDataDir"                "renamed. Use postgresql.dataMountPath plus postgresql.dataSubdir; PGDATA must be a strict subdirectory of the mount."
  "volumePermissions"                "podSecurityContext.fsGroup handles ownership on CSI drivers that honour it. On storage that ignores fsGroup (some NFS), reproduce the chown with postgresql.initContainers -- see the example in values.yaml."
  "networkPolicy"                    "NetworkPolicy is not rendered by this chart. Supply your own through the top-level extraObjects."
  "rbac"                             "no Role/RoleBinding is needed; the pod does not talk to the API server."
  "psp"                              "PodSecurityPolicy was removed from Kubernetes in 1.25."
  "shmVolume"                        "/dev/shm is always mounted as a Memory-backed emptyDir. Use postgresql.shmSizeLimit to bound it."
  "containerPorts"                   "renamed. Use postgresql.service.port."
  "extraDeploy"                      "renamed. Use the top-level extraObjects."
  "commonLabels"                     "renamed. Use the top-level extraLabels."
  "commonAnnotations"                "renamed. Use the top-level extraAnnotations."
  "clusterDomain"                    "not used; the chart addresses the database by Service name."
  "diagnosticMode"                   "not supported. Use postgresql.args to change the server command line."
-}}

{{- range $k, $v := $pg -}}
{{- if not (has $k $known) -}}
{{- if not (include "nebraska.postgresql.isInertValue" (dict "key" $k "value" $v)) -}}
{{- $help := index $guide $k | default "not read by this chart; check charts/nebraska/values.yaml for the current key." -}}
{{- $found = append $found (printf "postgresql.%s -- %s" $k $help) -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{- $knownAuth := list "username" "database" "postgresPassword" "existingSecret" "secretKeys" -}}
{{- $guideAuth := dict
  "password"            "the official image has a single superuser password. Use postgresql.auth.postgresPassword (secret key postgres-password)."
  "enablePostgresUser"  "POSTGRES_USER is always the superuser initdb creates; there is no separate postgres role to toggle."
  "replicationUsername" "replication is not provided by this chart."
  "replicationPassword" "replication is not provided by this chart."
  "usePasswordFiles"    "not supported; the password is injected with secretKeyRef."
-}}
{{- range $k, $v := ($pg.auth | default dict) -}}
{{- if not (has $k $knownAuth) -}}
{{- if not (include "nebraska.postgresql.isInertValue" (dict "key" $k "value" $v)) -}}
{{- $found = append $found (printf "postgresql.auth.%s -- %s" $k (index $guideAuth $k | default "not read by this chart.")) -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/* Most `primary.*` keys just moved up a level, so derive that message from
     the known-keys list rather than writing fourteen near-identical table
     entries by hand. Only keys whose replacement is NOT a simple rename need a
     bespoke message below. `extraEnvVars` is the one rename that changed name
     as well as level, so it is listed explicitly. */}}
{{/* Recurse one level into auth.secretKeys: only adminPasswordKey is read, and
     Bitnami's siblings (userPasswordKey, replicationPasswordKey) were passing
     silently -- exactly the class of gap the deny-unknown design exists to
     close, missed because the recursion stopped at auth.* */}}
{{- range $k, $v := (($pg.auth | default dict).secretKeys | default dict) -}}
{{- if ne $k "adminPasswordKey" -}}
{{- if not (include "nebraska.postgresql.isInertValue" (dict "key" $k "value" $v)) -}}
{{- $found = append $found (printf "postgresql.auth.secretKeys.%s -- not read by this chart. The official image has a single superuser, so only adminPasswordKey applies." $k) -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{- $guidePrimary := dict
  "configuration"             "custom postgresql.conf is not rendered. Set server options with postgresql.args, e.g. -c max_connections=200."
  "extendedConfiguration"     "set server options with postgresql.args."
  "existingConfigmap"         "set server options with postgresql.args, or mount your own file and point at it with postgresql.args."
  "existingExtendedConfigmap" "set server options with postgresql.args."
  "pgHbaConfiguration"        "custom pg_hba.conf is not rendered. The official image's initdb writes pg_hba.conf inside PGDATA."
  "initdb"                    "mount a ConfigMap at /docker-entrypoint-initdb.d with postgresql.extraVolumes and postgresql.extraVolumeMounts. Note it only runs on a first-time init of an empty data directory."
  "standby"                   "standby/streaming replication is not provided by this chart."
  "extraEnvVars"              "moved to postgresql.extraEnv."
  "extraEnvVarsCM"            "not supported; the official image reads only a few POSTGRES_* variables, and only on first init. Use postgresql.extraEnv."
  "extraEnvVarsSecret"        "not supported; use postgresql.extraEnv with a secretKeyRef."
  "podAntiAffinityPreset"     "meaningless for a single replica; there is no second pod to schedule away from."
  "updateStrategy"            "fixed to RollingUpdate; with a single replica there is nothing else to choose."
  "priorityClassName"         "set it through postgresql.extraPodSpec."
  "schedulerName"             "set it through postgresql.extraPodSpec."
  "hostAliases"               "set it through postgresql.extraPodSpec."
  "topologySpreadConstraints" "set it through postgresql.extraPodSpec."
  "command"                   "not supported. Use postgresql.args to pass arguments to postgres."
  "service"                   "only the port is configurable, as postgresql.service.port."
  "resources"                 "moved to postgresql.resources. Left here your CPU/memory limits would be silently dropped."
  "podSecurityContext"        "moved to postgresql.podSecurityContext. Left here your runAsUser/fsGroup would be silently dropped -- which matters, because the image's uid changed."
  "livenessProbe"             "probes are fixed by this chart. postgresql.startupProbe tunes the first-start budget."
  "readinessProbe"            "probes are fixed by this chart. postgresql.startupProbe tunes the first-start budget."
  "lifecycleHooks"            "not supported; the chart sets a preStop hook for clean shutdown."
  "sidecars"                  "not supported. A metrics exporter or backup agent does not need to share the pod -- run it as its own Deployment against the Service. If you need one badly enough, use postgresql.enabled=false and a real database."
  "initContainers"            "set them through postgresql.extraPodSpec."
  "nodeSelector"              "set it through postgresql.extraPodSpec."
  "tolerations"               "set them through postgresql.extraPodSpec."
  "affinity"                  "set it through postgresql.extraPodSpec."
  "podLabels"                 "use the top-level extraLabels, which apply to every object."
  "podAnnotations"            "use the top-level extraAnnotations, which apply to every object."
-}}
{{- range $k, $v := ($pg.primary | default dict) -}}
{{- if ne $k "persistence" -}}
{{- if not (include "nebraska.postgresql.isInertValue" (dict "key" $k "value" $v)) -}}
{{- $help := index $guidePrimary $k -}}
{{- if not $help -}}
{{- $help = ternary (printf "moved to postgresql.%s." $k) "not read by this chart." (has $k $known) -}}
{{- end -}}
{{- $found = append $found (printf "postgresql.primary.%s -- %s" $k $help) -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{- $knownPersist := list "enabled" "storageClass" "accessModes" "size" "labels" "annotations" -}}
{{- $guidePersist := dict
  "existingClaim" "not supported. The chart always uses a volumeClaimTemplate named 'data', so leaving this set would provision a NEW, EMPTY volume and leave your claim untouched."
  "mountPath"     "renamed. Use postgresql.dataMountPath."
  "subPath"       "not supported. Use postgresql.dataSubdir, which is the PGDATA subdirectory inside the volume."
-}}
{{- range $k, $v := (($pg.primary | default dict).persistence | default dict) -}}
{{- if not (has $k $knownPersist) -}}
{{- if not (include "nebraska.postgresql.isInertValue" (dict "key" $k "value" $v)) -}}
{{- $found = append $found (printf "postgresql.primary.persistence.%s -- %s" $k (index $guidePersist $k | default "not read by this chart.")) -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/* The subchart honoured global.*; nothing here does except global.imageRegistry
     and global.storageClass. global.postgresql.auth.* is Bitnami's own documented
     way to set the password, so silently ignoring it would rewrite a live Secret
     to the chart default while the database keeps the old password. */}}
{{- $global := .Values.global | default dict -}}
{{- if $global.postgresql -}}
{{- $found = append $found "global.postgresql.* -- not read by this chart. Use postgresql.auth.* and postgresql.image.* instead." -}}
{{- end -}}
{{- range $k, $v := $global -}}
{{- if not (has $k (list "imageRegistry" "storageClass" "postgresql")) -}}
{{- if not (include "nebraska.postgresql.isInertValue" (dict "key" $k "value" $v)) -}}
{{- $found = append $found (printf "global.%s -- not read by this chart; only global.imageRegistry and global.storageClass are honoured." $k) -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{- if $found -}}
{{- fail (printf "\n\nThese values are not read by this chart and would have been silently ignored:\n\n  - %s\n\nChart 3.0.0 replaced the Bitnami postgresql subchart with an in-chart StatefulSet\nrunning the official postgres image. See \"Upgrading to 3.0.0\" in the chart README.\nValues left at their Bitnami defaults are not reported, so everything listed above\nwould really have changed behaviour.\n\nIf you need something this chart does not provide, set postgresql.enabled=false and\npoint config.database.* at a database you operate.\n" (join "\n  - " $found)) -}}
{{- end -}}
{{- end -}}

{{/*
Refuse a Bitnami-family image.

The whole point of 3.0.0 is to stop shipping an image that never gets CVE fixes.
A values file carrying `postgresql.image.repository: bitnamilegacy/postgresql`
from 2.0.0 still renders, and the resulting pod fails in a way that looks like a
chart bug (the Bitnami entrypoint does not understand this chart's PGDATA layout)
rather than a stale value. Fail with an explanation instead.
*/}}
{{- define "nebraska.postgresql.validateImage" -}}
{{- $img := (.Values.postgresql | default dict).image | default dict -}}
{{- /* Match the FULL reference, not just the repository. Checking `repository`
       alone was bypassable by pushing the vendor name into the registry:
         --set postgresql.image.registry=docker.io/bitnamilegacy
         --set postgresql.image.repository=postgresql
       rendered silently. A guard with a trivial bypass is worse than none,
       because it advertises protection it does not provide. */ -}}
{{- $ref := printf "%s/%s" ($img.registry | default "" | toString) ($img.repository | default "" | toString) -}}
{{- if regexMatch "bitnami" (lower $ref) -}}
{{- fail (printf "\n\nThe PostgreSQL image resolves to %q, which is a Bitnami-family image.\n\nChart 3.0.0 runs the official postgres image and configures it accordingly\n(PGDATA layout, uid, env var names). A Bitnami image will not start correctly\nhere, and bitnamilegacy/* is a frozen archive that receives no security updates\n-- which is the reason this chart stopped using it.\n\nRemove the postgresql.image override to use the chart default, or set\npostgresql.enabled=false and run your own database.\n" $ref) -}}
{{- end -}}
{{- end -}}
