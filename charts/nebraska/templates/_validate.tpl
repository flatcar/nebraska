{{/*
Refuse to upgrade a persistent install onto the new data directory layout
without an explicit acknowledgement.

This is the most important guard in the chart. The Bitnami subchart mounted the
PVC at /bitnami/postgresql with PGDATA=/bitnami/postgresql/data; this chart
mounts the same PVC at /var/lib/postgresql/data with PGDATA=.../data/pgdata.

Every field the Kubernetes StatefulSet controller treats as immutable --
spec.selector, spec.serviceName, spec.volumeClaimTemplates, is deliberately
exactly what the subchart emitted, so `helm upgrade` is ACCEPTED and
returns 0. The pod restarts, finds no PG_VERSION at the new PGDATA, runs initdb
into a fresh sibling directory, and comes up as an empty cluster. Nebraska's
pod spec is unchanged by the upgrade, so it is NOT restarted and keeps serving
from stale connections; it recreates its schema on its next restart. Every
application, group, channel and rollout disappears from the user's point of
view while the old cluster sits untouched next to it.
No error is produced anywhere.

Prose in the README does not defend against this: nobody reads a README during
a Renovate bump. So the render fails until the operator says they have dealt
with the data.

The old cluster is still on the volume, so an accidental upgrade is recoverable
with `helm rollback`. That is worth knowing, so the message says it.

This file only exists for the migration. Nothing in it prints anything when the
values are fine, so it can be removed a release or two after 3.0.0, once nobody
upgrades straight from 2.0.0 any more.

If you tidy this file, be careful with the scan loops in validateUnknownValues.
They look repetitive and easy to merge, but two bugs were found in them during
review, and one earlier attempt to simplify them broke the chart. Re-run the
guard checks after any change here.
*/}}
{{- define "nebraska.postgresql.validateDataDirMigration" -}}
{{- $pg := .Values.postgresql | default dict -}}
{{- /* Only an UPGRADE can destroy data this way. A fresh install has no prior
       cluster on the volume, so failing there would be pure friction for every
       new user who wants persistence. Note this means template-rendering
       workflows (Argo/Flux, `helm template`) never see the gate, because
       .Release.IsUpgrade is false there, those users get the README and
       NOTES.txt instead. */ -}}
{{- /* Compare as a string: `--set-string ...=false` passes the STRING "false",
       which is truthy, so a truthiness check reads it as an acknowledgement and
       the gate stays down. Only an actual true (bool or string) acknowledges. */ -}}
{{- if and .Release.IsUpgrade $pg.enabled ((($pg.primary | default dict).persistence | default dict).enabled) -}}
{{- if ne ($pg.acknowledgeDataDirMigration | toString) "true" -}}
{{- fail "\n\nSTOP. This upgrade would silently discard your database.\n\nChart 3.0.0 replaced the Bitnami postgresql subchart with the official postgres\nimage, which stores data at a different path inside the same volume:\n\n  chart 2.0.0 (Bitnami):  PVC mounted at /bitnami/postgresql       PGDATA=/bitnami/postgresql/data\n  chart 3.0.0 (official): PVC mounted at /var/lib/postgresql/data  PGDATA=/var/lib/postgresql/data/pgdata\n\nNothing in Kubernetes rejects this change, so `helm upgrade` would SUCCEED and\nPostgreSQL would initialise a brand-new empty database alongside your existing\none. Nebraska would come up looking healthy with no applications, groups or\nrollouts, and no error would be reported.\n\nYou have persistence enabled, so you must choose:\n\n  1. Migrate the data (dump/restore). Follow \"Upgrading to 3.0.0\" in the chart\n     README, then re-run with:\n         --set postgresql.acknowledgeDataDirMigration=true\n\n  2. Deliberately start from an empty database (fine for dev/test). Same flag:\n         --set postgresql.acknowledgeDataDirMigration=true\n\n  3. Stay on chart 2.0.0 for now.\n\nIf you already ran this upgrade by accident: your old cluster is still present\non the volume, untouched, in the `data/` directory. Run `helm rollback` NOW,\nbefore deleting any PVC, and it will come back.\n" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Report values that this chart does not honour.

DENY-UNKNOWN, not allow-known. The first version of this listed the Bitnami keys
someone remembered, and a review of it found that it missed about 55 of them
-- including `primary.resources` (memory limits silently vanish),
`primary.podSecurityContext` (a user's fsGroup silently replaced),
`primary.persistence.existingClaim` (silently ignored, so a fresh EMPTY volume is
provisioned) and the whole `global.*` tree. An allowlist can only ever catch the
keys its author thought of, which is the wrong failure mode for a data-bearing
chart. So: enumerate the keys this chart READS, and report everything else.

DEFAULT-AWARE. A very common pattern is to vendor the upstream subchart's
values.yaml wholesale and edit a few lines. Such a file carries dozens of keys at
their Bitnami defaults, `architecture: standalone`, `metrics.enabled: false`,
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
     block that is switched off, a map whose every member is itself inert, or
     a setting sitting at its Bitnami default (e.g. architecture: standalone). */}}
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
  "architecture"                     "streaming standbys are not provided by this chart. Note Nebraska cannot use a read-only standby anyway. Every Omaha check-in writes, and it opens a single DSN. If you are after the distributed topology in RFC #1375, that uses one-way LOGICAL replication, which this chart can do: set postgresql.args to include -c wal_level=logical. For managed HA, use an operator such as CloudNativePG with postgresql.enabled=false."
  "replication"                      "streaming replication is not provided by this chart. Logical replication is reachable via postgresql.args (-c wal_level=logical); for managed HA use an operator with postgresql.enabled=false."
  "readReplicas"                     "read replicas are not provided by this chart, and Nebraska opens a single DSN so it has no read/write split to use them."
  "metrics"                          "the postgres-exporter sidecar, ServiceMonitor and PrometheusRule are gone. Run postgres-exporter as its own Deployment against the PostgreSQL Service, and supply it plus any ServiceMonitor through the top-level extraObjects."
  "tls"                              "in-chart TLS termination is gone. Set server options via postgresql.args (-c ssl=on -c ssl_cert_file=...) with the cert supplied through postgresql.extraVolumes, or terminate at a proxy/mesh."
  "ldap"                             "LDAP auth is gone. It was never used by Nebraska."
  "audit"                            "pgAuditLog/pgAuditLogCatalog need the pgaudit extension, which the official image does not ship. The other audit settings (logConnections, logDisconnections, logHostname, logLinePrefix, logTimezone, clientMinMessages) are plain PostgreSQL settings, set them via postgresql.args, e.g. -c log_connections=on."
  "postgresqlSharedPreloadLibraries" "set this via postgresql.args (-c shared_preload_libraries=...). Note the official image does not ship pgaudit."
  "postgresqlDataDir"                "renamed. Use postgresql.dataMountPath plus postgresql.dataSubdir; PGDATA must be a strict subdirectory of the mount."
  "volumePermissions"                "podSecurityContext.fsGroup handles ownership on CSI drivers that honour it. On storage that ignores fsGroup (some NFS), reproduce the chown with postgresql.extraPodSpec.initContainers, see the worked example in values.yaml."
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
{{- $found = append $found (printf "postgresql.%s: %s" $k $help) -}}
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
{{- $found = append $found (printf "postgresql.auth.%s: %s" $k (index $guideAuth $k | default "not read by this chart.")) -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/* Recurse one level into auth.secretKeys: only adminPasswordKey is read, and
     Bitnami's siblings (userPasswordKey, replicationPasswordKey) were passing
     silently, exactly the class of gap the deny-unknown design exists to
     close, missed because the recursion stopped at auth.* */}}
{{- range $k, $v := (($pg.auth | default dict).secretKeys | default dict) -}}
{{- if ne $k "adminPasswordKey" -}}
{{- if not (include "nebraska.postgresql.isInertValue" (dict "key" $k "value" $v)) -}}
{{- $found = append $found (printf "postgresql.auth.secretKeys.%s: not read by this chart. The official image has a single superuser, so only adminPasswordKey applies." $k) -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/* Recurse into serviceAccount: only create/name/automountServiceAccountToken
     are read. annotations is the one to watch, because it passes the top level check
     serviceAccount is known, then is silently discarded. */}}
{{- $knownSA := list "create" "name" "automountServiceAccountToken" -}}
{{- $guideSA := dict
  "annotations" "not read for the PostgreSQL ServiceAccount. Use the top-level extraAnnotations, which reach every object."
-}}
{{- range $k, $v := (($pg.serviceAccount | default dict)) -}}
{{- if not (has $k $knownSA) -}}
{{- if not (include "nebraska.postgresql.isInertValue" (dict "key" $k "value" $v)) -}}
{{- $found = append $found (printf "postgresql.serviceAccount.%s: %s" $k (index $guideSA $k | default "not read by this chart.")) -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/* extraPodSpec is merged at pod-spec level; `containers` or `volumes` there
     emit a SECOND key alongside the chart's own, producing an invalid pod spec
     (the values.yaml example says as much in prose, prose is not a guard). */}}
{{- if $pg.enabled -}}
{{- range $k, $v := ($pg.extraPodSpec | default dict) -}}
{{- if has $k (list "containers" "volumes") -}}
{{- $found = append $found (printf "postgresql.extraPodSpec.%s: duplicates a key the chart renders itself, which makes the pod spec invalid. Set extraEnv/extraVolumes/extraVolumeMounts (first-class values) or use initContainers via extraPodSpec." $k) -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/* Most `primary.*` keys just moved up a level, so derive that message from
     the known-keys list rather than writing a near-identical table entry for
     each by hand. Only keys whose replacement is NOT a simple rename need a
     bespoke message below. `extraEnvVars` is the one rename that changed name
     as well as level, so it is listed explicitly. */}}
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
  "podSecurityContext"        "moved to postgresql.podSecurityContext. Left here your runAsUser/fsGroup would be silently dropped, which matters, because the image's uid changed."
  "livenessProbe"             "probes are fixed by this chart. postgresql.startupProbe tunes the first-start budget."
  "readinessProbe"            "probes are fixed by this chart. postgresql.startupProbe tunes the first-start budget."
  "lifecycleHooks"            "not supported; the chart sets a preStop hook for clean shutdown."
  "sidecars"                  "not supported. A metrics exporter or backup agent does not need to share the pod, run it as its own Deployment against the Service. If you need one badly enough, use postgresql.enabled=false and a real database."
  "initContainers"            "set them through postgresql.extraPodSpec.initContainers."
  "nodeSelector"              "set it through postgresql.extraPodSpec."
  "tolerations"               "set them through postgresql.extraPodSpec."
  "affinity"                  "set it through postgresql.extraPodSpec."
  "podLabels"                 "not offered on the PostgreSQL pod template; the top-level extraLabels apply to object metadata only and cannot be used in pod selectors. For NetworkPolicy, match the pod's existing labels through extraObjects."
  "podAnnotations"            "not offered on the PostgreSQL pod template. The top-level extraAnnotations apply to object metadata only."
-}}
{{- range $k, $v := ($pg.primary | default dict) -}}
{{- if ne $k "persistence" -}}
{{- if has $k (list "livenessProbe" "readinessProbe") -}}
{{- /* The subchart defaulted these to enabled: true, so a deliberate
       `enabled: false` is a REAL change, not an inert default, isInertValue
       would swallow it and silently turn the probe ON again. Report any
       non-empty setting; only a completely empty value means nothing. */ -}}
{{- if $v -}}
{{- $found = append $found (printf "postgresql.primary.%s: probes are fixed by this chart. postgresql.startupProbe tunes the first-start budget." $k) -}}
{{- end -}}
{{- else if not (include "nebraska.postgresql.isInertValue" (dict "key" $k "value" $v)) -}}
{{- $help := index $guidePrimary $k -}}
{{- if not $help -}}
{{- $help = ternary (printf "moved to postgresql.%s." $k) "not read by this chart." (has $k $known) -}}
{{- end -}}
{{- $found = append $found (printf "postgresql.primary.%s: %s" $k $help) -}}
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
{{- $found = append $found (printf "postgresql.primary.persistence.%s: %s" $k (index $guidePersist $k | default "not read by this chart.")) -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/* The subchart honoured global.*; nothing here does except global.imageRegistry
     and global.storageClass. global.postgresql.auth.* is Bitnami's own documented
     way to set the password, so silently ignoring it would rewrite a live Secret
     to the chart default while the database keeps the old password. */}}
{{- $global := .Values.global | default dict -}}
{{- if $global.postgresql -}}
{{- $found = append $found "global.postgresql.*: not read by this chart. Use postgresql.auth.* and postgresql.image.* instead." -}}
{{- end -}}
{{- range $k, $v := $global -}}
{{- if not (has $k (list "imageRegistry" "storageClass" "postgresql")) -}}
{{- if not (include "nebraska.postgresql.isInertValue" (dict "key" $k "value" $v)) -}}
{{- $found = append $found (printf "global.%s: not read by this chart; only global.imageRegistry and global.storageClass are honoured." $k) -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{- if $found -}}
{{- fail (printf "\n\nThese values are not read by this chart and would have been silently ignored:\n\n  - %s\n\nChart 3.0.0 replaced the Bitnami postgresql subchart with an in-chart StatefulSet\nrunning the official postgres image. See \"Upgrading to 3.0.0\" in the chart README.\nValues left at their Bitnami defaults are not reported, so everything listed above\nwould really have changed behaviour.\n\nIf you need something this chart does not provide, set postgresql.enabled=false and\npoint config.database.* at a database you operate.\n" (join "\n  - " $found)) -}}
{{- end -}}
{{- end -}}

{{/*
Refuse a data mount that sits above the image's declared VOLUME.

The postgres image declares a VOLUME. If the PVC is mounted at an ANCESTOR of
that path, the container runtime mounts an empty volume over the top and
everything the PVC holds underneath becomes invisible inside the container. The
pod starts, initdb runs into what looks like empty space, and the real data is
still on the PV where nobody can see it.

docker-library documents this for <=17 ("mount at /var/lib/postgresql/data and
NOT at /var/lib/postgresql, or data WILL NOT PERSIST"). It is widely assumed to
be a Docker-only quirk that Kubernetes ignores. It is not, this was observed
live on kind, with a Bitnami data directory present on the PV and unreadable
from inside the pod.

The VOLUME moved in 18 (/var/lib/postgresql/data -> /var/lib/postgresql), so the
correct mount point depends on the major version. The tag is parsed to pick the
right one; an unrecognisable tag is left alone rather than guessed at.

Why this guard is here at all, since most people never touch dataMountPath:
it is a value this chart exposes, and the README tells people to change it when
they move to PostgreSQL 18, because the VOLUME moved in 18. So a user can reach
this problem by following our own instructions, not only by making a typo. The
failure is silent and complete: the pod starts, looks healthy, and the data is
invisible. That is worth a guard even if it is rarely hit.

It is best effort by design. It is the only place the chart reads the image tag
to decide something, and an unknown tag is left alone instead of guessed.
*/}}
{{- define "nebraska.postgresql.validateMountPath" -}}
{{- $pg := .Values.postgresql | default dict -}}
{{- if $pg.enabled -}}
{{- $mount := $pg.dataMountPath | default "" | toString | trimSuffix "/" -}}
{{- $sub := $pg.dataSubdir | default "" | toString -}}
{{- /* PGDATA is mount/subdir, CONCATENATED AS TEXT and resolved by the kernel:
       `../../../../tmp/pgdata` lands in the ephemeral /tmp emptyDir with the
       pod looking healthy. Reject anything that can escape the volume; allow
       clean nested values like `18/docker`. */ -}}
{{- if or (not $sub) (hasPrefix "/" $sub) (eq $sub ".") -}}
{{- fail (printf "\n\npostgresql.dataSubdir %q is not usable: it must be a non-empty relative\nsubdirectory of the volume mount (e.g. pgdata, or 18/docker for PostgreSQL 18).\n" $sub) -}}
{{- end -}}
{{- range $seg := (splitList "/" $sub) -}}
{{- if eq $seg ".." -}}
{{- fail (printf "\n\npostgresql.dataSubdir %q contains '..': PGDATA is computed as\ndataMountPath/dataSubdir and resolved by the kernel, so this escapes the data\nvolume (e.g. into the ephemeral /tmp emptyDir) while the pod looks healthy.\nUse a plain relative subdirectory like pgdata or 18/docker.\n" $sub) -}}
{{- end -}}
{{- end -}}
{{- if or (not $mount) (not (hasPrefix "/" $mount)) -}}
{{- fail (printf "\n\npostgresql.dataMountPath %q is not usable: it must be an absolute path\n(e.g. /var/lib/postgresql/data).\n" $mount) -}}
{{- end -}}
{{- range $seg := (splitList "/" $mount) -}}
{{- if eq $seg ".." -}}
{{- fail (printf "\n\npostgresql.dataMountPath %q contains '..'.\n" $mount) -}}
{{- end -}}
{{- end -}}
{{- $major := regexFind "^[0-9]+" ((($pg.image | default dict).tag | default "" | toString)) -}}
{{- if and $major $mount -}}
{{- $vol := ternary "/var/lib/postgresql" "/var/lib/postgresql/data" (ge (int $major) 18) -}}
{{- if hasPrefix (printf "%s/" $mount) $vol -}}
{{- fail (printf "\n\npostgresql.dataMountPath is %q, which is above the VOLUME the image declares (%q\nfor PostgreSQL %s).\n\nMounting the data volume above the image's VOLUME makes the runtime lay an empty\nvolume over the top: the pod starts, PostgreSQL initialises into what looks like\nempty space, and everything already on your PVC becomes invisible from inside the\ncontainer. It is still on the disk, and nothing will tell you.\n\nSet postgresql.dataMountPath to %q (and keep postgresql.dataSubdir as the\nsubdirectory inside it).\n" $mount $vol $major $vol) -}}
{{- end -}}
{{- end -}}
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
{{- $pg := .Values.postgresql | default dict -}}
{{- $img := $pg.image | default dict -}}
{{- /* Match the FULL EFFECTIVE reference, not just the repository. Checking
       `repository` alone was bypassable by pushing the vendor name into the
       registry (--set postgresql.image.registry=docker.io/bitnamilegacy), and
       checking the local registry alone was bypassable by pushing it into
       global.imageRegistry, which nebraska.image applies as an override. A
       guard with a trivial bypass is worse than none, because it advertises
       protection it does not provide. */ -}}
{{- $registry := ($img.registry | default "" | toString) -}}
{{- with .Values.global -}}{{- with .imageRegistry -}}{{- $registry = . -}}{{- end -}}{{- end -}}
{{- $ref := printf "%s/%s" $registry ($img.repository | default "" | toString) -}}
{{- /* Gated on enabled: with postgresql.enabled=false the image is never
       pulled, so a stale override is inert and refusing would contradict the
       README's "no action needed" for external-database users. */ -}}
{{- if and $pg.enabled (regexMatch "bitnami" (lower $ref)) -}}
{{- fail (printf "\n\nThe PostgreSQL image resolves to %q, which is a Bitnami-family image.\n\nChart 3.0.0 runs the official postgres image and configures it accordingly\n(PGDATA layout, uid, env var names). A Bitnami image will not start correctly\nhere, and bitnamilegacy/* is a frozen archive that receives no security updates\n-- which is the reason this chart stopped using it.\n\nRemove the postgresql.image override to use the chart default, or set\npostgresql.enabled=false and run your own database.\n" $ref) -}}
{{- end -}}
{{- /* An empty tag with no digest renders `postgres:`, unpullable, and the
       appVersion fallback is deliberately disabled for this image (postgres
       has no 3.x matching the chart's). Fail with a message instead. */ -}}
{{- if and $pg.enabled (not ($img.tag | default "" | toString)) (not ($img.digest | default "" | toString)) -}}
{{- fail "\n\npostgresql.image.tag is empty and no postgresql.image.digest is set, so the image\nwould render as 'postgres:' with no tag, unpullable. (The chart's appVersion\nfallback applies to the Nebraska image only; there is no postgres:3.0.0.)\nSet a tag (e.g. 17-bookworm) or a digest.\n" -}}
{{- end -}}
{{- end -}}
