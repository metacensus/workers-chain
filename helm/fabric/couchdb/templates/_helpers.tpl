{{/*
Helpers for fabric-couchdb chart
Each chart includes shared helpers for independence and clarity
*/}}

{{/* === Shared across all fabric charts === */}}

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

{{- define "fabric.imagePullSecrets" -}}
{{- if .Values.image.pullSecrets }}
imagePullSecrets:
  {{- toYaml .Values.image.pullSecrets | nindent 2 }}
{{- end }}
{{- end }}

{{/* === CouchDB-specific helpers === */}}

{{- define "fabric-couchdb.name" -}}
{{ .Values.organization.id }}-{{ .Values.component }}-{{ .Values.number }}
{{- end }}

{{- define "fabric-couchdb.selectorLabels" -}}
app.kubernetes.io/name: fabric-couchdb
app.kubernetes.io/instance: {{ .Release.Name }}
fabric.metacensus.org/organization: {{ .Values.organization.id }}
fabric.metacensus.org/component: {{ .Values.component }}
fabric.metacensus.org/number: {{ .Values.number | quote }}
fabric.metacensus.org/type: couchdb
{{- end }}

{{- define "fabric-couchdb.labels" -}}
{{ include "fabric-couchdb.selectorLabels" . }}
{{ include "fabric.commonLabels" . }}
app.kubernetes.io/component: couchdb
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}

