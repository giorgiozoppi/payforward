{{/*
Expand the name of the chart.
*/}}
{{- define "payforward-backend.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "payforward-backend.fullname" -}}
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
{{- define "payforward-backend.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "payforward-backend.labels" -}}
helm.sh/chart: {{ include "payforward-backend.chart" . }}
{{ include "payforward-backend.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "payforward-backend.selectorLabels" -}}
app.kubernetes.io/name: {{ include "payforward-backend.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "payforward-backend.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "payforward-backend.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Neo4j URI
*/}}
{{- define "payforward-backend.neo4jUri" -}}
{{- if .Values.neo4j.enabled }}
{{- .Values.neo4j.uri }}
{{- else }}
{{- .Values.externalNeo4j.uri }}
{{- end }}
{{- end }}

{{/*
Neo4j Secret Name
*/}}
{{- define "payforward-backend.neo4jSecretName" -}}
{{- if .Values.neo4j.enabled }}
{{- .Values.neo4j.existingSecret }}
{{- else }}
{{- .Values.externalNeo4j.existingSecret }}
{{- end }}
{{- end }}
