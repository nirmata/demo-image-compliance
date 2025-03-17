{{/* vim: set filetype=mustache: */}}

{{- define "nirmata-image-compliance.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "nirmata-image-compliance.fullname" -}}
{{- if .Values.fullnameOverride -}}
  {{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
  {{- $name := default .Chart.Name .Values.nameOverride -}}
  {{- if contains $name .Release.Name -}}
    {{- .Release.Name | trunc 63 | trimSuffix "-" -}}
  {{- else -}}
    {{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
  {{- end -}}
{{- end -}}
{{- end -}}

{{- define "nirmata-image-compliance.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "nirmata-image-compliance.chartVersion" -}}
{{- .Chart.Version | replace "+" "_" -}}
{{- end -}}

{{- define "nirmata-image-compliance.namespace" -}}
{{ default .Release.Namespace .Values.namespaceOverride }}
{{- end -}}

{{- define "nirmata-image-compliance.clusterRoleName" -}}
{{ include "nirmata-image-compliance.fullname" . }}-clusterrole
{{- end -}}

{{- define "nirmata-image-compliance.roleName" -}}
{{ include "nirmata-image-compliance.fullname" . }}-role
{{- end -}}

{{- define "nirmata-image-compliance.serviceAccountName" -}}
{{ default (include "nirmata-image-compliance.name" .) .Values.serviceAccount.name }}
{{- end -}}

{{- define "nirmata-image-compliance.serviceName" -}}
{{- printf "%s-svc" (include "nirmata-image-compliance.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "nirmata-image-compliance.configMap" -}}
{{ default (include "nirmata-image-compliance.name" .) .Values.configMap.name }}
{{- end -}}
