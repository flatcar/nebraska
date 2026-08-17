{{/*
Expand the name of the chart.
*/}}
{{- define "nebraska.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "nebraska.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "nebraska.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "nebraska.labels" -}}
helm.sh/chart: {{ include "nebraska.chart" . }}
{{ include "nebraska.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "nebraska.selectorLabels" -}}
app.kubernetes.io/name: {{ include "nebraska.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "nebraska.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "nebraska.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Fully qualified name for the bundled PostgreSQL objects.

[PARITY] This reproduces the Bitnami subchart's `common.names.fullname` semantics exactly,
including honouring fullnameOverride and the `contains` short-circuit that skips
the suffix when the release name already contains the component name. Those are
not stylistic details: any divergence renames the StatefulSet, which orphans the
PVC and hands the user a new empty database. A release named `nebraska` with
nameOverride `nebraska` produced `nebraska` before and would produce
`nebraska-nebraska` under a naive implementation.
*/}}
{{- define "nebraska.postgresql.fullname" -}}
{{- if .Values.postgresql.fullnameOverride -}}
{{- .Values.postgresql.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default "postgresql" .Values.postgresql.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Standard metadata for the bundled PostgreSQL objects.

Factored out because the same labels/annotations stanza appeared on each of
them. Used by four of the five: the Secret inlines its own copy because it also
carries `helm.sh/resource-policy: keep`, and merging one fixed annotation into
this helper would cost more indirection than the six duplicated lines. Emits the
`annotations:` key only when there is something to put under it, so it stays
valid when extraAnnotations is empty.
*/}}
{{- define "nebraska.postgresql.metadata" -}}
labels:
  {{- include "nebraska.postgresql.labels" . | nindent 2 }}
  {{- with .Values.extraLabels }}
  {{- toYaml . | nindent 2 }}
  {{- end }}
{{- with .Values.extraAnnotations }}
annotations:
  {{- toYaml . | nindent 2 }}
{{- end }}
{{- end -}}

{{/*
The readiness/liveness/startup check. One definition, three call sites.
*/}}
{{- define "nebraska.postgresql.probeCommand" -}}
exec:
  command:
    - /bin/sh
    - -c
    - exec pg_isready -U {{ .Values.postgresql.auth.username | quote }} -d {{ printf "dbname=%s" .Values.postgresql.auth.database | quote }} -h 127.0.0.1 -p {{ int .Values.postgresql.service.port }}
{{- end -}}

{{/*
[PARITY] ServiceAccount used by the PostgreSQL pod. Honours an explicit name, as the
Bitnami subchart did, so an existing install pinning one keeps working.
*/}}
{{- define "nebraska.postgresql.serviceAccountName" -}}
{{- if .Values.postgresql.serviceAccount.create -}}
{{- default (include "nebraska.postgresql.fullname" .) .Values.postgresql.serviceAccount.name -}}
{{- else -}}
{{- default "default" .Values.postgresql.serviceAccount.name -}}
{{- end -}}
{{- end -}}

{{/*
Name of the headless Service backing the StatefulSet.

[EXTRA, not required by the migration] A pre-existing bug that the Bitnami
subchart had too, handled with exact parity wherever a working install can exist.

The subchart computed `printf "%s-hl" fullname | trunc 63 | trimSuffix "-"`.
At release-name length 50 the fullname is 61 chars, so that yields
`<fullname>-h`, distinct from the main Service, a working install, and an
immutable `spec.serviceName` that this chart has to reproduce exactly. So the
Bitnami expression is used verbatim whenever its result differs from the main
Service name.

At lengths 51-52 the truncation drops the whole "-hl" (and trimSuffix drops
the dangling "-"), so the headless name COLLIDES with the main Service name and
the apiserver rejects the release: no working install can exist to stay
compatible with there. Only for that case the base is truncated to 60 before
appending, which is guaranteed distinct and <=63. (Appending without truncating
would give 66 characters, equally rejected.)
*/}}
{{- define "nebraska.postgresql.headlessName" -}}
{{- $fullname := include "nebraska.postgresql.fullname" . -}}
{{- $bitnami := printf "%s-hl" $fullname | trunc 63 | trimSuffix "-" -}}
{{- if ne $bitnami $fullname -}}
{{- $bitnami -}}
{{- else -}}
{{- printf "%s-hl" ($fullname | trunc 60 | trimSuffix "-") -}}
{{- end -}}
{{- end -}}

{{/*
Resolve the StorageClass for the data volume.

[PARITY] Honours global.storageClass like the subchart did, and reproduces its "-"
sentinel, which means "render an empty storageClassName" (bind to a pre-created
PV / disable dynamic provisioning) rather than "use a StorageClass literally
named -". Emits nothing when unset, so the cluster default applies.
*/}}
{{- define "nebraska.postgresql.storageClass" -}}
{{- $sc := .Values.postgresql.primary.persistence.storageClass | default (.Values.global | default dict).storageClass -}}
{{- if $sc -}}
{{- if eq $sc "-" -}}
storageClassName: ""
{{- else -}}
storageClassName: {{ $sc | quote }}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Selector labels for the bundled PostgreSQL StatefulSet.

[MIGRATION] These are the same labels the Bitnami postgresql 11.9.1 subchart
used. A StatefulSet spec.selector cannot be changed after it is created, so
keeping them the same lets `helm upgrade` patch the existing StatefulSet instead
of failing. Do not change them without a major chart bump.

Why they are kept as they are:

If we renamed them to match the rest of the chart (for example
`app.kubernetes.io/name: nebraska` with `component: postgresql`) the labels
would read better, but every existing install would fail to upgrade. The user
would have to run `kubectl delete statefulset --cascade=orphan` by hand first.

3.0.0 is a breaking release anyway, so this was the cheapest moment to rename
them. We decided not to. Upgrading in place without manual steps is worth more
than nicer labels. The price is that the chart keeps a selector that looks
wrong, `name: postgresql` in a chart called nebraska, and we cannot fix that
until the next major version.

If this is revisited later, the change is small. It only affects this
StatefulSet selector and the labels built from it. It does not affect the
Nebraska Deployment, the Secret, or the data.
*/}}
{{- define "nebraska.postgresql.selectorLabels" -}}
{{- /* The name label follows nameOverride (NOT fullnameOverride), because the
       subchart's `common.names.name` is `default .Chart.Name .Values.nameOverride`
       and the StatefulSet's immutable selector embeds it. A 2.0.0 install with
       postgresql.nameOverride=pg has `name: pg` in its selector; hardcoding
       "postgresql" here would change an immutable field and get the upgrade
       REJECTED by the apiserver, verified against the vendored subchart. */ -}}
app.kubernetes.io/name: {{ default "postgresql" .Values.postgresql.nameOverride }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/component: primary
{{- end -}}

{{/*
Common labels for the bundled PostgreSQL objects.
*/}}
{{- define "nebraska.postgresql.labels" -}}
helm.sh/chart: {{ include "nebraska.chart" . }}
{{ include "nebraska.postgresql.selectorLabels" . }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: {{ include "nebraska.name" . }}
{{- end -}}

{{/*
Name of the secret holding the PostgreSQL superuser password.
*/}}
{{- define "nebraska.postgresql.secretName" -}}
{{- if .Values.postgresql.auth.existingSecret -}}
{{- $name := tpl .Values.postgresql.auth.existingSecret . -}}
{{- /* Naming the chart's own secret suppresses its template while leaving both
       workloads referencing it, so Helm deletes it on upgrade and everything
       fails with CreateContainerConfigError, with the password gone. */ -}}
{{- if eq $name (include "nebraska.postgresql.fullname" .) -}}
{{- fail (printf "postgresql.auth.existingSecret must not name the secret this chart manages (%s): Helm would stop rendering it and then delete it, taking the password with it. Either drop existingSecret and set postgresql.auth.postgresPassword, or point it at a separately-managed secret." $name) -}}
{{- end -}}
{{- $name -}}
{{- else -}}
{{- include "nebraska.postgresql.fullname" . -}}
{{- end -}}
{{- end -}}

{{/*
Return the appropriate apiVersion for ingress
*/}}
{{- define "nebraska.ingress.apiVersion" -}}
{{- if semverCompare "<1.14-0" .Capabilities.KubeVersion.Version -}}
{{- print "extensions/v1beta1" -}}
{{- else if semverCompare "<1.19-0" .Capabilities.KubeVersion.Version -}}
{{- print "networking.k8s.io/v1beta1" -}}
{{- else -}}
{{- print "networking.k8s.io/v1" -}}
{{- end -}}
{{- end -}}

{{- define "nebraska.ingressScheme" -}}
http{{ if $.Values.ingress.tls }}s{{ end }}
{{- end -}}

{{/*
Return the proper image name.
This honours global overrides for the image registry the same way the former
postgresql subchart did.
*/}}
{{- define "nebraska.image" -}}
{{- $registryName := .imageRoot.registry -}}
{{- $repositoryName := .imageRoot.repository -}}
{{- $tag := "" -}}
{{- if .context }}
{{- $tag = .imageRoot.tag | default .context.Chart.AppVersion | toString -}}
{{- else }}
{{- $tag = .imageRoot.tag | toString -}}
{{- end -}}
{{- if .global }}
    {{- if .global.imageRegistry }}
     {{- $registryName = .global.imageRegistry -}}
    {{- end -}}
{{- end -}}
{{- /* A digest is content-addressed, so it pins the image regardless of what the
       tag later points at. When set it replaces the tag entirely. */ -}}
{{- if .imageRoot.digest -}}
{{- printf "%s/%s@%s" $registryName $repositoryName (.imageRoot.digest | toString) -}}
{{- else -}}
{{- printf "%s/%s:%s" $registryName $repositoryName $tag -}}
{{- end -}}
{{- end -}}
