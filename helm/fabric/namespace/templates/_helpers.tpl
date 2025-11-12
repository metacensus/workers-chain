{{/*
Shared helper functions for Fabric charts
Duplicated across all fabric charts for independence
*/}}

{{- define "fabric.commonLabels" -}}
app.kubernetes.io/name: fabric
app.kubernetes.io/managed-by: helm
app.kubernetes.io/part-of: fabric
{{- end }}

{{- define "fabric.namespace" -}}
{{- if .Values.namespace.name }}
{{ .Values.namespace.name }}
{{- else }}
{{ .Release.Namespace }}
{{- end }}
{{- end }}

