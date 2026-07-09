{{/*
Shared helper functions for Certificate Authority chart
*/}}

{{- define "certificate-authority.name" -}}
certificate-authority-{{ .Values.certType }}-{{ .Values.caId }}
{{- end }}

{{- define "certificate-authority.namespace" -}}
{{ .Values.namespace }}
{{- end }}

{{- define "certificate-authority.serviceName" -}}
{{ include "certificate-authority.name" . }}
{{- end }}

{{- define "certificate-authority.commonLabels" -}}
app.kubernetes.io/name: {{ include "certificate-authority.name" . }}
app.kubernetes.io/component: certificate-authority
app.kubernetes.io/managed-by: {{ .Values.commonLabels.managedBy }}
app.kubernetes.io/part-of: {{ .Values.commonLabels.partOf }}
{{- end }}

{{- define "certificate-authority.selectorLabels" -}}
app.kubernetes.io/name: {{ include "certificate-authority.name" . }}
app.kubernetes.io/component: certificate-authority
{{- end }}




