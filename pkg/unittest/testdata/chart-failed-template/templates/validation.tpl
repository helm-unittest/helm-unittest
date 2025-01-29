{{- if .Values.case1 }}
{{- fail (printf "`runAsNonRoot` is set to `true` but `runAsUser` is set to `0` (root)") }}
{{- end }}

{{- if .Values.case2 }}
{{- fail (printf "\n`in-backtick` start with new line following backtick") }}
{{- end }}

{{- if .Values.case3 }}
{{- fail (printf "error contains single \\ escape character ") }}
{{- end }}

{{- if .Values.case4 }}
{{- fail (printf "\nerror contains single \\ escape character ") }}
{{- end }}

