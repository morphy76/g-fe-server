{{/*
Expand the name of the chart.
*/}}
{{- define "g-fe-server.name" -}}
{{- default .Chart.Name .Values.g_fe_server.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "g-fe-server.fullname" -}}
{{- if .Values.g_fe_server.fullnameOverride }}
{{- .Values.g_fe_server.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.g_fe_server.nameOverride }}
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
{{- define "g-fe-server.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "g-fe-server.labels" -}}
helm.sh/chart: {{ include "g-fe-server.chart" . }}
{{ include "g-fe-server.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "g-fe-server.selectorLabels" -}}
app.kubernetes.io/name: {{ include "g-fe-server.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "g-fe-server.serviceAccountName" -}}
{{- if .Values.g_fe_server.serviceAccount.create }}
{{- default (include "g-fe-server.fullname" .) .Values.g_fe_server.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.g_fe_server.serviceAccount.name }}
{{- end }}
{{- end }}
