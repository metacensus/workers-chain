{{/*
Expand the name of the chart.
*/}}
{{- define "fabric.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "fabric.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "fabric.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "fabric.labels" -}}
helm.sh/chart: {{ include "fabric.chart" . }}
{{ include "fabric.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- with .Values.commonLabels }}
{{ toYaml . }}
{{- end }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "fabric.selectorLabels" -}}
app.kubernetes.io/name: {{ include "fabric.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
CA specific labels
*/}}
{{- define "fabric.ca.labels" -}}
{{ include "fabric.labels" . }}
app.kubernetes.io/component: ca
{{- end }}

{{/*
Peer specific labels
*/}}
{{- define "fabric.peer.labels" -}}
{{ include "fabric.labels" . }}
app.kubernetes.io/component: peer
{{- end }}

{{/*
CouchDB specific labels
*/}}
{{- define "fabric.couchdb.labels" -}}
{{ include "fabric.labels" . }}
app.kubernetes.io/component: couchdb
{{- end }}

{{/*
Orderer specific labels
*/}}
{{- define "fabric.orderer.labels" -}}
{{ include "fabric.labels" . }}
app.kubernetes.io/component: orderer
{{- end }}

{{/*
MSP volume mount configuration
*/}}
{{- define "fabric.mspVolumeMount" -}}
- name: msp
  mountPath: {{ .Values.msp.basePath }}
  readOnly: true
{{- end }}

{{/*
MSP volume configuration from secret
*/}}
{{- define "fabric.mspVolume" -}}
- name: msp
  secret:
    secretName: {{ . }}
{{- end }}

{{/*
Create the namespace
*/}}
{{- define "fabric.namespace" -}}
{{- if .Values.namespace.name }}
{{- .Values.namespace.name }}
{{- else }}
{{- .Release.Namespace }}
{{- end }}
{{- end }}

{{/*
Image pull secrets
*/}}
{{- define "fabric.imagePullSecrets" -}}
{{- if .Values.image.pullSecrets }}
imagePullSecrets:
{{- range .Values.image.pullSecrets }}
  - name: {{ . }}
{{- end }}
{{- end }}
{{- end }}

