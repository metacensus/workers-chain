{{/*
Shared helper functions for Fabric Peer chart
*/}}

{{- define "peer.name" -}}
{{ .Values.peerId }}
{{- end }}

{{- define "peer.namespace" -}}
{{ .Values.namespace }}
{{- end }}

{{- define "peer.serviceName" -}}
{{ .Values.peerId }}
{{- end }}

{{- define "peer.commonLabels" -}}
app.kubernetes.io/name: {{ include "peer.name" . }}
app.kubernetes.io/component: peer
app.kubernetes.io/managed-by: {{ .Values.commonLabels.managedBy }}
app.kubernetes.io/part-of: {{ .Values.commonLabels.partOf }}
{{- end }}

{{- define "peer.selectorLabels" -}}
app.kubernetes.io/name: {{ include "peer.name" . }}
app.kubernetes.io/component: peer
{{- end }}

