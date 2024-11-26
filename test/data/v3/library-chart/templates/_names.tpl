{{/*
Expand the name of the chart.
{{ include "common.names.name" . -}}
*/}}
{{- define "common.names.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
{{ include "common.names.chart" . -}}
*/}}
{{- define "common.names.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Allow the release namespace to be overridden for multi-namespace deployments in combined charts.
{{ include "common.names.namespace" . -}}
*/}}
{{- define "common.names.namespace" -}}
{{- default .Release.Namespace .Values.namespace | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
List the service name.
{{ include "common.names.service-name" . -}}
*/}}
{{- define "common.names.service-name" -}}
{{- default (include "common.names.name" . ) .Values.serviceName | trunc 63 | trimSuffix "-" -}}
{{- end -}}
