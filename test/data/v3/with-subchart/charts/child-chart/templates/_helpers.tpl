{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "child-chart.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
*/}}
{{- define "child-chart.fullname" -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "child-chart.ingressClass" -}}
  {{- $ingressClass := .Values.ingress.class -}}
  {{- if .Values.global -}}
    {{- if .Values.global.ingressClass -}}
      {{- $ingressClass = .Values.global.ingressClass -}}
    {{- end -}}
  {{- end -}}
  {{- printf "%s" $ingressClass -}}
{{- end -}}

{{- define "sec.containerSeccompProfile" -}}
{{- $profile := . -}}
{{/*- fail (printf "%s-%s" "my-error: " ($profile.type)) -*/}}
{{- if and $profile $profile.type -}}
seccompProfile:
  type: {{ $profile.type }}
{{- if eq $profile.type "Localhost" }}
{{- if (empty $profile.localhostProfile) }}
  {{- fail "The 'Localhost' seccomp profile requires a profile name to be provided in localhostProfile parameter." -}}
{{- else }}
  localhostProfile: {{ $profile.localhostProfile }}
{{- end }}
{{- end -}}
{{- end -}}
{{- end -}}
