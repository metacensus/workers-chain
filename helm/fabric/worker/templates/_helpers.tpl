{{/*
Shared helper functions for Fabric Worker chart
*/}}

{{- define "fabric-worker.name" -}}
fabric-worker-{{ .Values.org }}
{{- end }}

{{- define "fabric-worker.namespace" -}}
{{ .Values.namespace }}
{{- end }}

{{- define "fabric-worker.serviceAccountName" -}}
fabric-worker-{{ .Values.org }}
{{- end }}

{{- define "fabric-worker.commonLabels" -}}
app.kubernetes.io/name: {{ include "fabric-worker.name" . }}
app.kubernetes.io/component: worker
app.kubernetes.io/managed-by: {{ .Values.commonLabels.managedBy }}
app.kubernetes.io/part-of: {{ .Values.commonLabels.partOf }}
{{- end }}

{{- define "fabric-worker.selectorLabels" -}}
app.kubernetes.io/name: {{ include "fabric-worker.name" . }}
app.kubernetes.io/component: worker
{{- end }}



