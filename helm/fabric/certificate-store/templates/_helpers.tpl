{{/*
Shared helper functions for Certificate Store chart
*/}}

{{- define "certificate-store.name" -}}
fabric-postgres-{{ .Values.org }}
{{- end }}

{{- define "certificate-store.namespace" -}}
fabric-{{ .Values.org }}
{{- end }}

{{- define "certificate-store.serviceName" -}}
fabric-postgres-{{ .Values.org }}
{{- end }}

{{- define "certificate-store.commonLabels" -}}
app.kubernetes.io/name: {{ include "certificate-store.name" . }}
app.kubernetes.io/component: certificate-store
app.kubernetes.io/managed-by: {{ .Values.commonLabels.managedBy }}
app.kubernetes.io/part-of: {{ .Values.commonLabels.partOf }}
{{- end }}

{{- define "certificate-store.selectorLabels" -}}
app.kubernetes.io/name: {{ include "certificate-store.name" . }}
app.kubernetes.io/component: certificate-store
{{- end }}



