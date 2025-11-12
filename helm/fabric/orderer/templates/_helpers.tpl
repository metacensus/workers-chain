{{/*
Helpers for fabric-orderer chart
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

{{/* === Shared across fabric node types (peer, orderer, ca) === */}}

{{- define "fabric.csiVolumeMount" -}}
- name: fabric-msp
  mountPath: {{ .Values.msp.basePath }}
  readOnly: true
{{- end }}

{{- define "fabric.mspVolumeMount" -}}
- name: fabric-msp
  mountPath: {{ .Values.msp.basePath }}
{{- end }}

{{- define "fabric.csiVolume" -}}
- name: fabric-msp
  csi:
    driver: secrets-store.csi.k8s.io
    readOnly: true
    volumeAttributes:
      secretProviderClass: {{ . | quote }}
{{- end }}

{{- define "fabric.mspVolume" -}}
- name: fabric-msp
  secret:
    secretName: {{ . }}
{{- end }}

{{- define "fabric.imagePullSecrets" -}}
{{- if .Values.image.pullSecrets }}
imagePullSecrets:
  {{- toYaml .Values.image.pullSecrets | nindent 2 }}
{{- end }}
{{- end }}

{{/* === Orderer-specific helpers === */}}

{{- define "fabric-orderer.name" -}}
{{ .Values.organization.id }}-{{ .Values.component }}-{{ .Values.number }}
{{- end }}

{{- define "fabric-orderer.selectorLabels" -}}
app.kubernetes.io/name: fabric-orderer
app.kubernetes.io/instance: {{ .Release.Name }}
fabric.metacensus.org/organization: {{ .Values.organization.id }}
fabric.metacensus.org/component: {{ .Values.component }}
fabric.metacensus.org/number: {{ .Values.number | quote }}
fabric.metacensus.org/type: orderer
{{- end }}

{{- define "fabric-orderer.labels" -}}
{{ include "fabric-orderer.selectorLabels" . }}
{{ include "fabric.commonLabels" . }}
app.kubernetes.io/component: orderer
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}

