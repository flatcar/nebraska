{{/*
Refuse to upgrade a persistent install onto the new data directory layout
without an explicit acknowledgement.

The Bitnami subchart used PGDATA=/bitnami/postgresql/data; this chart uses
/var/lib/postgresql/data/pgdata on the same PVC. The StatefulSet's immutable
fields match the subchart's, so `helm upgrade` succeeds and PostgreSQL starts an
empty cluster next to the old one, with no error anywhere. The old cluster stays
on the volume, so `helm rollback` recovers it.

The scan loops in validateUnknownValues look repetitive, but an earlier attempt to
merge them broke the chart. Re-run the CI guard checks after any change here.
*/}}
{{- define "nebraska.postgresql.validateDataDirMigration" -}}
{{- $pg := .Values.postgresql | default dict -}}
{{- /* Only an upgrade can lose data this way. `helm template` and GitOps never
       set IsUpgrade, so they rely on the README and NOTES.txt. Compare the flag
       as a string, so that --set-string ...=false is not read as true. */ -}}
{{- /* Compare as a string: `--set-string ...=false` passes the STRING "false",
       which is truthy, so a truthiness check reads it as an acknowledgement and
       the gate stays down. Only an actual true (bool or string) acknowledges. */ -}}
{{- /* Skip only when the live StatefulSet comes from this chart (label
       "nebraska-...", Bitnami's is "postgresql-...") and already uses the same
       PGDATA. Without a cluster, lookup is empty and the gate fires. */ -}}
{{- $pgdata := printf "%s/%s" ($pg.dataMountPath | default "") ($pg.dataSubdir | default "") -}}
{{- $sameLayout := false -}}
{{- $live := (lookup "apps/v1" "StatefulSet" .Release.Namespace (include "nebraska.postgresql.fullname" .)) | default dict -}}
{{- $liveChart := include "nebraska.postgresql.liveChart" . -}}
{{- $persist := (($pg.primary | default dict).persistence | default dict).enabled -}}
{{- if and .Release.IsUpgrade $pg.enabled (not $persist) ($live.spec | default dict).volumeClaimTemplates -}}
{{- fail "\n\npostgresql.primary.persistence.enabled is false, but the running PostgreSQL\nStatefulSet has a data volume. Kubernetes cannot remove the volume from a\nStatefulSet, so this upgrade would fail half-way, after Helm has already\nchanged other objects such as the password Secret.\n\nSet postgresql.primary.persistence.enabled=true. Helm does not reuse the old\nvalues once you pass any value, so a values file without this setting turns\npersistence off.\n" -}}
{{- end -}}
{{- if hasPrefix (printf "%s-" .Chart.Name) $liveChart -}}
{{- range ((($live.spec | default dict).template | default dict).spec | default dict).containers | default list -}}
  {{- range .env | default list -}}
    {{- if and (eq .name "PGDATA") (eq (.value | default "") $pgdata) -}}
      {{- $sameLayout = true -}}
    {{- end -}}
  {{- end -}}
{{- end -}}
{{- end -}}
{{- if and .Release.IsUpgrade $pg.enabled $persist (not $sameLayout) -}}
{{- if ne ($pg.acknowledgeDataDirMigration | toString) "true" -}}
{{- fail (printf "\n\nSTOP. This upgrade would silently discard your database.\n\nChart 4.0.0 replaced the Bitnami postgresql subchart with the official postgres\nimage, which stores data at a different path inside the same volume:\n\n  chart 3.0.0 (Bitnami):  PVC mounted at /bitnami/postgresql  PGDATA=/bitnami/postgresql/data\n  this upgrade renders:   PVC mounted at %s  PGDATA=%s\n\nNothing in Kubernetes rejects this change, so `helm upgrade` would SUCCEED and\nPostgreSQL would initialise a brand-new empty database alongside your existing\none. Nebraska would come up looking healthy with no applications, groups or\nrollouts, and no error would be reported.\n\nYou have persistence enabled, so you must choose:\n\n  1. Reuse the volume in place (same PostgreSQL major version, no dump). Follow\n     \"In-place upgrade\" in the chart README, then re-run with:\n         --set postgresql.acknowledgeDataDirMigration=true\n\n  2. Migrate the data (dump/restore). Follow \"Upgrading to 4.0.0\" in the chart\n     README; its last step sets the same flag.\n\n  3. Deliberately start from an empty database (fine for dev/test). Same flag.\n\n  4. Stay on chart 3.0.0 for now.\n\nIf this release already runs chart 4.0.0:\n  - A client-side --dry-run or `helm template --is-upgrade` cannot see the\n    cluster, so it stops here. Use --dry-run=server, and do not set the flag.\n  - If you changed postgresql.dataMountPath or postgresql.dataSubdir by mistake,\n    set them back; the flag would start an empty database.\n  - If you are moving to a new PostgreSQL major version and already took a dump\n    (\"Upgrade PostgreSQL\" in the chart README), the flag is expected.\n\nIf you already ran this upgrade by accident: your old cluster is still present\non the volume, untouched. Run `helm rollback` NOW, before deleting any PVC, and\nit will come back.\n" ($pg.dataMountPath | toString) $pgdata) -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Report values that this chart does not read, instead of ignoring them. This is a
deny-list of everything unknown, not an allow-list of old Bitnami keys, because
an allow-list only catches the keys its author remembered.
Inert values (empty, false, or a feature block that is switched off, such as
`metrics.enabled: false`) are ignored, because they changed nothing before either.
*/}}

{{/* True when a removed value would not have done anything anyway: empty,
     false, an empty collection, `architecture: standalone`, a feature block
     that is switched off, or a map whose every member is itself inert. */}}
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
  "password"            "the official image has a single superuser password. Use postgresql.auth.postgresPassword (secret key postgres-password). After an in-place upgrade, a custom auth.username keeps this password; if it differs from postgresPassword, let Nebraska log in as the superuser: config.database.username=postgres with secret key postgres-password."
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

{{/* Recurse into image and service. Both are known keys, so a typo under them
     (`digests`, `pullPolicies`) or a Bitnami-only child passed the top-level
     check and was dropped. For image that matters most: a misspelt digest
     silently falls back to the floating tag. */}}
{{- $knownImage := list "registry" "repository" "tag" "digest" "pullPolicy" "pullSecrets" -}}
{{- range $k, $v := ($pg.image | default dict) -}}
{{- if not (has $k $knownImage) -}}
{{- if not (include "nebraska.postgresql.isInertValue" (dict "key" $k "value" $v)) -}}
{{- $found = append $found (printf "postgresql.image.%s: %s" $k (ternary "Bitnami image debug logging is gone. For verbose server logs use postgresql.args, e.g. -c log_min_messages=debug1." (printf "not read by this chart; the image keys are %s." (join ", " $knownImage)) (eq $k "debug"))) -}}
{{- end -}}
{{- end -}}
{{- end -}}
{{- range $k, $v := ($pg.service | default dict) -}}
{{- if ne $k "port" -}}
{{- if not (include "nebraska.postgresql.isInertValue" (dict "key" $k "value" $v)) -}}
{{- $found = append $found (printf "postgresql.service.%s: not read by this chart; only postgresql.service.port is configurable. The Services are always ClusterIP." $k) -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/* extraPodSpec is rendered next to the pod spec keys the chart sets itself.
     Setting one of them there gives a duplicate key. securityContext and
     imagePullSecrets are rendered only in some cases, but they have their own
     values, so they are rejected always. */}}
{{- $podOwned := dict
  "containers"                    "use postgresql.extraEnv and postgresql.extraVolumeMounts."
  "volumes"                       "use postgresql.extraVolumes."
  "serviceAccountName"            "use postgresql.serviceAccount.name."
  "automountServiceAccountToken"  "use postgresql.serviceAccount.automountServiceAccountToken."
  "terminationGracePeriodSeconds" "use postgresql.terminationGracePeriodSeconds."
  "imagePullSecrets"              "use postgresql.image.pullSecrets."
  "securityContext"               "use postgresql.podSecurityContext."
-}}
{{- if $pg.enabled -}}
{{- range $k, $v := ($pg.extraPodSpec | default dict) -}}
{{- if hasKey $podOwned $k -}}
{{- $found = append $found (printf "postgresql.extraPodSpec.%s: the chart sets this key itself, so it would be a duplicate key in the pod spec. Instead, %s" $k (index $podOwned $k)) -}}
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
{{- if has $k (list "livenessProbe" "readinessProbe" "podSecurityContext" "containerSecurityContext") -}}
{{- /* The subchart defaulted these to enabled: true, so `enabled: false` is a
       real change that isInertValue would swallow. Report any non-empty value. */ -}}
{{- if $v -}}
{{- $found = append $found (printf "postgresql.primary.%s: %s" $k (index $guidePrimary $k | default (printf "moved to postgresql.%s." $k))) -}}
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

{{/* The subchart honoured global.*; nothing here does except global.imageRegistry,
     global.imagePullSecrets and global.storageClass. global.postgresql.auth.* is Bitnami's own documented
     way to set the password, so silently ignoring it would rewrite a live Secret
     to the chart default while the database keeps the old password. */}}
{{- $global := .Values.global | default dict -}}
{{- if not (include "nebraska.postgresql.isInertValue" (dict "key" "postgresql" "value" $global.postgresql)) -}}
{{- $found = append $found "global.postgresql.*: not read by this chart. Use postgresql.auth.* and postgresql.image.* instead." -}}
{{- end -}}
{{- range $k, $v := $global -}}
{{- if not (has $k (list "imageRegistry" "imagePullSecrets" "storageClass" "postgresql")) -}}
{{- if not (include "nebraska.postgresql.isInertValue" (dict "key" $k "value" $v)) -}}
{{- $found = append $found (printf "global.%s: not read by this chart; only global.imageRegistry, global.imagePullSecrets and global.storageClass are honoured." $k) -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{- if $found -}}
{{- fail (printf "\n\nThese values are not read by this chart and would have been silently ignored:\n\n  - %s\n\nChart 4.0.0 replaced the Bitnami postgresql subchart with an in-chart StatefulSet\nrunning the official postgres image. See \"Upgrading to 4.0.0\" in the chart README.\nDisabled or empty values are not reported. Delete each key listed above, or move\nit to the replacement named in its message.\n\nIf you need something this chart does not provide, set postgresql.enabled=false and\npoint config.database.* at a database you operate.\n" (join "\n  - " $found)) -}}
{{- end -}}
{{- end -}}

{{/*
Refuse a data mount above the image's declared VOLUME. The runtime would mount an
empty volume over it, hiding the data on the PVC (seen on kind). The VOLUME moved
in PostgreSQL 18, so the major version is read from the tag; an unknown tag is
not checked.
*/}}
{{- define "nebraska.postgresql.validateMountPath" -}}
{{- $pg := .Values.postgresql | default dict -}}
{{- if $pg.enabled -}}
{{- $raw := $pg.dataMountPath | default "" | toString -}}
{{- $mount := $raw | clean | trimSuffix "/" -}}
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
{{- fail (printf "\n\npostgresql.dataMountPath %q is not usable: it must be an absolute path\n(e.g. /var/lib/postgresql/data).\n" $raw) -}}
{{- end -}}
{{- range $seg := (splitList "/" $raw) -}}
{{- if eq $seg ".." -}}
{{- fail (printf "\n\npostgresql.dataMountPath %q contains '..'.\n" $raw) -}}
{{- end -}}
{{- end -}}
{{- $major := regexFind "^[0-9]+" ((($pg.image | default dict).tag | default "" | toString)) -}}
{{- if and $major $mount -}}
{{- $vol := ternary "/var/lib/postgresql" "/var/lib/postgresql/data" (ge (int $major) 18) -}}
{{- if hasPrefix (printf "%s/" $mount) $vol -}}
{{- fail (printf "\n\npostgresql.dataMountPath is %q, which is above the VOLUME the image declares (%q\nfor PostgreSQL %s).\n\nMounting the data volume above the image's VOLUME makes the runtime lay an empty\nvolume over the top: the pod starts, PostgreSQL initialises into what looks like\nempty space, and everything already on your PVC becomes invisible from inside the\ncontainer. It is still on the disk, and nothing will tell you.\n\nSet postgresql.dataMountPath to %q (and keep postgresql.dataSubdir as the\nsubdirectory inside it).\n" $raw $vol $major $vol) -}}
{{- end -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Refuse a Nebraska password key that the chart's own Secret does not have.
*/}}
{{- define "nebraska.postgresql.validatePasswordKey" -}}
{{- $pg := .Values.postgresql | default dict -}}
{{- $ref := .Values.config.database.passwordExistingSecret -}}
{{- if and $pg.enabled $ref.enabled (not $pg.auth.existingSecret) -}}
{{- $name := tpl $ref.name . -}}
{{- $key := tpl $ref.key . -}}
{{- $adminKey := $pg.auth.secretKeys.adminPasswordKey -}}
{{- if and (eq $name (include "nebraska.postgresql.fullname" .)) (ne $key $adminKey) -}}
{{- fail (printf "\n\nconfig.database.passwordExistingSecret.key is %q, but the Secret\n%s that this chart creates only has the key %q,\nso Nebraska would not start.\n\nSet config.database.passwordExistingSecret.key=%s.\nAfter an in-place upgrade from chart 3.0.0 with a custom auth.username, also\nset config.database.username=postgres, because that user keeps its old\npassword.\n" $key $name $adminKey $adminKey) -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Refuse a Bitnami-family image. A 3.0.0 values file still sets
bitnamilegacy/postgresql, and its entrypoint does not work with this chart.
*/}}
{{- define "nebraska.postgresql.validateImage" -}}
{{- $pg := .Values.postgresql | default dict -}}
{{- $img := $pg.image | default dict -}}
{{- /* Check the full reference, including global.imageRegistry, so the vendor
       name cannot slip through in the registry field. */ -}}
{{- $registry := ($img.registry | default "" | toString) -}}
{{- with .Values.global -}}{{- with .imageRegistry -}}{{- $registry = . -}}{{- end -}}{{- end -}}
{{- $ref := printf "%s/%s" $registry ($img.repository | default "" | toString) -}}
{{- /* Gated on enabled: with postgresql.enabled=false the image is never
       pulled, so a stale override is inert and refusing would contradict the
       README's "no action needed" for external-database users. */ -}}
{{- if and $pg.enabled (regexMatch "bitnami" (lower $ref)) -}}
{{- fail (printf "\n\nThe PostgreSQL image resolves to %q, which is a Bitnami-family image.\n\nChart 4.0.0 runs the official postgres image and configures it accordingly\n(PGDATA layout, uid, env var names). A Bitnami image will not start correctly\nhere, and bitnamilegacy/* is a frozen archive that receives no security updates\n-- which is the reason this chart stopped using it.\n\nRemove the postgresql.image override to use the chart default, or set\npostgresql.enabled=false and run your own database.\n" $ref) -}}
{{- end -}}
{{- /* An empty tag with no digest renders `postgres:`, unpullable, and the
       appVersion fallback is deliberately disabled for this image (postgres
       has no 4.x matching the chart's). Fail with a message instead. */ -}}
{{- if and $pg.enabled (not ($img.tag | default "" | toString)) (not ($img.digest | default "" | toString)) -}}
{{- fail "\n\npostgresql.image.tag is empty and no postgresql.image.digest is set, so the image\nwould render as 'postgres:' with no tag, unpullable. (The chart's appVersion\nfallback applies to the Nebraska image only; there is no postgres:4.0.0.)\nSet a tag (e.g. 17-bookworm) or a digest.\n" -}}
{{- end -}}
{{- end -}}
