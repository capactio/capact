{{/*
Expand the name of the chart.
*/}}
{{- define "hub.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "hub.fullname" -}}
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
{{- define "hub.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "hub.labels" -}}
helm.sh/chart: {{ include "hub.chart" . }}
{{ include "hub.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "hub.selectorLabels" -}}
app.kubernetes.io/name: {{ include "hub.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "hub.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "hub.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the manifest source paths separated by the comma
*/}}
{{- define "populator.manifestSources" -}}
{{- $length := len .Values.populator.manifestsLocations }}
{{- range $index, $values := .Values.populator.manifestsLocations -}}
{{- include "populator.manifestSource" . }}
{{- $position := add $index 1 -}}
{{- if ne $position $length -}},{{- end -}}
{{- end }}
{{- end }}

{{/*
Create the manifest source path
*/}}
{{- define "populator.manifestSource" -}}
{{- if .local -}}
/hub-manifests
{{- else }}
{{- .repository -}}
?ref={{ .branch -}}
{{- if .sshKey -}}
&sshkey={{ .sshKey -}}
{{- end }}
{{- end }}
{{- end }}
