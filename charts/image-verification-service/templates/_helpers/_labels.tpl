{{/* vim: set filetype=mustache: */}}

{{- define "kyverno-image-verification-service.labels.merge" -}}
{{- $labels := dict -}}
{{- range . -}}
  {{- $labels = merge $labels (fromYaml .) -}}
{{- end -}}
{{- with $labels -}}
  {{- toYaml $labels -}}
{{- end -}}
{{- end -}}

{{- define "kyverno-image-verification-service.labels" -}}
{{- template "kyverno-image-verification-service.labels.merge" (list
  (include "kyverno-image-verification-service.labels.common" .)
  (include "kyverno-image-verification-service.matchLabels.common" .)
) -}}
{{- end -}}

{{- define "kyverno-image-verification-service.labels.helm" -}}
helm.sh/chart: {{ template "kyverno-image-verification-service.chart" . }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{- define "kyverno-image-verification-service.labels.version" -}}
app.kubernetes.io/version: {{ template "kyverno-image-verification-service.chartVersion" . }}
{{- end -}}

{{- define "kyverno-image-verification-service.labels.common" -}}
{{- template "kyverno-image-verification-service.labels.merge" (list
  (include "kyverno-image-verification-service.labels.helm" .)
  (include "kyverno-image-verification-service.labels.version" .)
) -}}
{{- end -}}

{{- define "kyverno-image-verification-service.matchLabels.common" -}}
app.kubernetes.io/part-of: {{ template "kyverno-image-verification-service.fullname" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}
