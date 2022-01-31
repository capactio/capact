{{/*
Expand the name of the chart.
*/}}
{{- define "capact.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "capact.fullname" -}}
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
{{- define "capact.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "capact.labels" -}}
helm.sh/chart: {{ include "capact.chart" . }}
{{ include "capact.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "capact.selectorLabels" -}}
app.kubernetes.io/name: {{ include "capact.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "capact.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "capact.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Get Dashboard URL
*/}}
{{- define "capact.dashboardURL" -}}
{{- if .Values.dashboard.ingress.enabled }}
{{- printf "https://%s.%s" .Values.dashboard.ingress.host .Values.global.domainName }}
{{- else }}
{{/*
TODO: Naive temporary implementation. After upgrade to Helm 3.7 or newer, use simply:
    {{- printf "http://%s.%s:%d" (include "dashboard.fullname" .Subcharts.dashboard) .Release.Namespace .Values.dashboard.service.port }}
    See issue: https://github.com/helm/helm/pull/9957
*/}}
{{- printf "http://%s-dashboard.%s:%d" .Release.Name .Release.Namespace .Values.dashboard.service.port }}
{{- end }}
{{- end }}
