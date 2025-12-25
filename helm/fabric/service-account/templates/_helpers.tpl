{{/*
Shared helper functions for Fabric charts
*/}}

{{- define "fabric.commonLabels" -}}
app.kubernetes.io/name: fabric
app.kubernetes.io/managed-by: {{ .Values.commonLabels.managedBy }}
app.kubernetes.io/part-of: {{ .Values.commonLabels.partOf }}
{{- end }}

{{- define "fabric.namespace" -}}
{{ .Values.namespace }}
{{- end }}

{{- define "fabric.serviceAccountName" -}}
{{ .Values.serviceAccountName }}
{{- end }}

{{- define "fabric.roleName" -}}
{{ .Values.roleName }}
{{- end }}


