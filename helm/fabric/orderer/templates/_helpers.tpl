{{/*
Shared helper functions for Fabric Orderer chart
*/}}

{{- define "orderer.name" -}}
{{ .Values.ordererId }}
{{- end }}

{{- define "orderer.namespace" -}}
{{ .Values.namespace }}
{{- end }}

{{- define "orderer.serviceName" -}}
{{ .Values.ordererId }}
{{- end }}

{{- define "orderer.commonLabels" -}}
app.kubernetes.io/name: {{ include "orderer.name" . }}
app.kubernetes.io/component: orderer
app.kubernetes.io/managed-by: {{ .Values.commonLabels.managedBy }}
app.kubernetes.io/part-of: {{ .Values.commonLabels.partOf }}
{{- end }}

{{- define "orderer.selectorLabels" -}}
app.kubernetes.io/name: {{ include "orderer.name" . }}
app.kubernetes.io/component: orderer
{{- end }}

