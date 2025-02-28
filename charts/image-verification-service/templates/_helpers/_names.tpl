{{/* vim: set filetype=mustache: */}}

{{- define "kyverno-image-verification-service.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "kyverno-image-verification-service.fullname" -}}
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

{{- define "kyverno-image-verification-service.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "kyverno-image-verification-service.chartVersion" -}}
{{- .Chart.Version | replace "+" "_" -}}
{{- end -}}

{{- define "kyverno-image-verification-service.namespace" -}}
{{ default .Release.Namespace .Values.namespaceOverride }}
{{- end -}}

{{- define "kyverno-image-verification-service.clusterRoleName" -}}
{{ include "kyverno-image-verification-service.fullname" . }}-clusterrole
{{- end -}}

{{- define "kyverno-image-verification-service.roleName" -}}
{{ include "kyverno-image-verification-service.fullname" . }}-role
{{- end -}}

{{- define "kyverno-image-verification-service.serviceAccountName" -}}
{{ default (include "kyverno-image-verification-service.name" .) .Values.serviceAccount.name }}
{{- end -}}

{{- define "kyverno-image-verification-service.serviceName" -}}
{{- printf "%s-svc" (include "kyverno-image-verification-service.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "kyverno-image-verification-service.configMap" -}}
{{ default (include "kyverno-image-verification-service.name" .) .Values.configMap.name }}
{{- end -}}
