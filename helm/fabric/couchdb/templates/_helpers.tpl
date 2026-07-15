{{/*
Shared helper functions for CouchDB chart
*/}}

{{- define "couchdb.name" -}}
couchdb-{{ .Values.peerId }}
{{- end }}

{{- define "couchdb.namespace" -}}
{{ .Values.namespace }}
{{- end }}

{{- define "couchdb.serviceName" -}}
couchdb-{{ .Values.peerId }}
{{- end }}

{{- define "couchdb.commonLabels" -}}
app.kubernetes.io/name: {{ include "couchdb.name" . }}
app.kubernetes.io/component: couchdb
app.kubernetes.io/managed-by: {{ .Values.commonLabels.managedBy }}
app.kubernetes.io/part-of: {{ .Values.commonLabels.partOf }}
{{- end }}

{{- define "couchdb.selectorLabels" -}}
app.kubernetes.io/name: {{ include "couchdb.name" . }}
app.kubernetes.io/component: couchdb
{{- end }}

