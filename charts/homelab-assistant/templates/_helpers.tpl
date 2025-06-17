{{/*
Expand the name of the chart.
*/}}
{{- define "homelab-assistant.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "homelab-assistant.fullname" -}}
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
{{- define "homelab-assistant.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "homelab-assistant.labels" -}}
helm.sh/chart: {{ include "homelab-assistant.chart" . }}
{{ include "homelab-assistant.selectorLabels" . }}
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
{{- define "homelab-assistant.selectorLabels" -}}
app.kubernetes.io/name: {{ include "homelab-assistant.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "homelab-assistant.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "homelab-assistant.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the name of the namespace to use
*/}}
{{- define "homelab-assistant.namespace" -}}
{{- if .Values.namespace.name }}
{{- .Values.namespace.name }}
{{- else }}
{{- .Release.Namespace }}
{{- end }}
{{- end }}

{{/*
Controller image
*/}}
{{- define "homelab-assistant.controller.image" -}}
{{- printf "%s:%s" .Values.controller.image.repository (.Values.controller.image.tag | default .Chart.AppVersion) }}
{{- end }}

{{/*
VolSync unlock job image
*/}}
{{- define "homelab-assistant.volsyncMonitor.unlockJob.image" -}}
{{- printf "%s:%s" .Values.volsyncMonitor.unlockJob.image.repository .Values.volsyncMonitor.unlockJob.image.tag }}
{{- end }}

{{/*
Common annotations
*/}}
{{- define "homelab-assistant.annotations" -}}
{{- with .Values.commonAnnotations }}
{{ toYaml . }}
{{- end }}
{{- end }}
