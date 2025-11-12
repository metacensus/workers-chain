{{/*
Helpers for fabric-spc chart
Each chart includes shared helpers for independence and clarity
*/}}

{{/* === Shared across all fabric charts === */}}

{{- define "fabric.commonLabels" -}}
app.kubernetes.io/name: fabric
app.kubernetes.io/managed-by: helm
app.kubernetes.io/part-of: fabric
{{- end }}

{{- define "fabric.namespace" -}}
{{- if .Values.namespace }}
{{ .Values.namespace }}
{{- else }}
{{ .Release.Namespace }}
{{- end }}
{{- end }}

{{/* === SPC-specific helpers === */}}

{{- define "fabric-spc.resourceName" -}}
{{- if .Values.resourceName }}
{{ .Values.resourceName }}
{{- else }}
{{ .Values.organization }}-{{ .Values.identity }}-spc
{{- end }}
{{- end }}

{{- define "fabric-spc.vaultPath" -}}
secret/data/fabric/{{ .Values.organization }}/msp/{{ .Values.identity }}
{{- end }}

{{- define "fabric-spc.labels" -}}
{{ include "fabric.commonLabels" . }}
app.kubernetes.io/component: secretproviderclass
fabric.metacensus.org/organization: {{ .Values.organization }}
fabric.metacensus.org/identity: {{ .Values.identity }}
{{- end }}

