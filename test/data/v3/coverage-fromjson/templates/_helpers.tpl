{{- define "coverage-fromjson.db" -}}
{{- $d := dict "scheme" "mysql" -}}
{{- if .enabled -}}
{{- $d = merge $d (dict "mode" "primary") -}}
{{- else -}}
{{- $d = merge $d (dict "mode" "fallback") -}}
{{- end -}}
{{- mustToJson (dict "db" $d) -}}
{{- end -}}
