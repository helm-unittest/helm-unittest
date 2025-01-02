{{- if .Values.case1 }}
{{- fail (printf "`runAsNonRoot` is set to `true` but `runAsUser` is set to `0` (root)") }}
{{- end }}

{{- if .Values.case2 }}
{{- fail (printf "\n`runAsNonRoot` start with new line and backtick") }}
{{- end }}

{{- if .Values.case3 }}
{{- fail (printf "error contains single \\ escape character ") }}
{{- end }}

{{- if .Values.case4 }}
{{- fail (printf "\nerror contains single \\ escape character ") }}
{{- end }}
