{{/*
Expand the name of the chart.
*/}}
{{- define "g-be-example.name" -}}
{{- default .Chart.Name .Values.g_be_example.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "g-be-example.fullname" -}}
{{- if .Values.g_be_example.fullnameOverride }}
{{- .Values.g_be_example.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.g_be_example.nameOverride }}
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
{{- define "g-be-example.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "g-be-example.labels" -}}
helm.sh/chart: {{ include "g-be-example.chart" . }}
{{ include "g-be-example.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "g-be-example.selectorLabels" -}}
app.kubernetes.io/name: {{ include "g-be-example.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "g-be-example.serviceAccountName" -}}
{{- if .Values.g_be_example.serviceAccount.create }}
{{- default (include "g-be-example.fullname" .) .Values.g_be_example.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.g_be_example.serviceAccount.name }}
{{- end }}
{{- end }}
