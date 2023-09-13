{{/* Checks for `.Release.name` */}}
{{- if gt (len .Release.Name) 3 }}
  {{ required "The `.Release.name`: must be <= 3 characters!" .Values.versionAlwaysFail }}
{{- end }}

