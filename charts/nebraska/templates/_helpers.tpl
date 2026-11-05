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
Fully qualified name for the bundled PostgreSQL objects. Same rules as the
Bitnami `common.names.fullname`, including the `contains` short-cut: a different
name would rename the StatefulSet and orphan its PVC.
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
Labels and annotations for the PostgreSQL objects. The Secret has its own copy
because it also needs `helm.sh/resource-policy: keep`.
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
The helm.sh/chart label of the live PostgreSQL StatefulSet, or "" when there is
none or no cluster (`helm template`, client dry-run). "postgresql-..." means the
Bitnami subchart; "nebraska-..." means this chart.
*/}}
{{- define "nebraska.postgresql.liveChart" -}}
{{- $live := (lookup "apps/v1" "StatefulSet" .Release.Namespace (include "nebraska.postgresql.fullname" .)) | default dict -}}
{{- index (($live.metadata | default dict).labels | default dict) "helm.sh/chart" | default "" -}}
{{- end -}}

{{/*
The probe command. No shell, so `$` or quotes in names are passed unchanged. -d
takes a connection string, so the name is quoted with \ and ' escaped for libpq.
*/}}
{{- define "nebraska.postgresql.probeCommand" -}}
exec:
  command:
    - pg_isready
    - -U
    - {{ .Values.postgresql.auth.username | toString | quote }}
    - -d
    - {{ printf "dbname='%s'" (.Values.postgresql.auth.database | toString | replace "\\" "\\\\" | replace "'" "\\'") | quote }}
    - -h
    - 127.0.0.1
    - -p
    - {{ int .Values.postgresql.service.port | quote }}
{{- end -}}

{{/*
ServiceAccount of the PostgreSQL pod. Honours an explicit name, as the subchart did.
*/}}
{{- define "nebraska.postgresql.serviceAccountName" -}}
{{- if .Values.postgresql.serviceAccount.create -}}
{{- default (include "nebraska.postgresql.fullname" .) .Values.postgresql.serviceAccount.name -}}
{{- else -}}
{{- default "default" .Values.postgresql.serviceAccount.name -}}
{{- end -}}
{{- end -}}

{{/*
Name of the headless Service. Same formula as Bitnami, because
spec.serviceName is immutable. Only when that formula gives the same name as
the main Service (a 62-63 character name), the base is cut to 60 characters
first. Chart 3.0.0 could not run a PostgreSQL pod at that length, so no working
release changes.
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
StorageClass of the data volume. global.storageClass wins over
primary.persistence.storageClass, and "-" means an empty storageClassName, both
as in the subchart. The order matters: volumeClaimTemplates is immutable, so a
different class fails the upgrade from 3.0.0. Unset means the cluster default.
*/}}
{{- define "nebraska.postgresql.storageClass" -}}
{{- $sc := (.Values.global | default dict).storageClass | default .Values.postgresql.primary.persistence.storageClass -}}
{{- if $sc -}}
{{- if eq $sc "-" -}}
storageClassName: ""
{{- else -}}
storageClassName: {{ $sc | quote }}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Selector labels of the PostgreSQL StatefulSet. They are the Bitnami 11.9.1 labels,
because spec.selector is immutable: changing them makes `helm upgrade` fail until
the StatefulSet is deleted by hand. Change them only in a major version.
*/}}
{{- define "nebraska.postgresql.selectorLabels" -}}
{{- /* Follows nameOverride, as Bitnami's `common.names.name` did, so a 3.0.0
       install with postgresql.nameOverride keeps the same selector. */ -}}
app.kubernetes.io/name: {{ default "postgresql" .Values.postgresql.nameOverride | trunc 63 | trimSuffix "-" }}
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
Name of the secret holding the PostgreSQL superuser password. With an external
database it stays `<release>-postgresql`, the name chart 3.0.0 used, because
users created that Secret themselves.
*/}}
{{- define "nebraska.postgresql.secretName" -}}
{{- if not .Values.postgresql.enabled -}}
{{- printf "%s-postgresql" .Release.Name -}}
{{- else if .Values.postgresql.auth.existingSecret -}}
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
