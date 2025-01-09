{{/*
Kubernetes standard annotations
{{ include "common.annotations.standard" . }}
*/}}
{{- define "common.annotations.standard" -}}
meta.helm.sh/release-name: {{ include "common.names.name" . | quote }}
meta.helm.sh/release-namespace: {{ include "common.names.namespace" . | quote }}
{{- end -}}
