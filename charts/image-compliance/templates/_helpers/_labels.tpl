{{/* vim: set filetype=mustache: */}}

{{- define "nirmata-image-compliance.labels.merge" -}}
{{- $labels := dict -}}
{{- range . -}}
  {{- $labels = merge $labels (fromYaml .) -}}
{{- end -}}
{{- with $labels -}}
  {{- toYaml $labels -}}
{{- end -}}
{{- end -}}

{{- define "nirmata-image-compliance.labels" -}}
{{- template "nirmata-image-compliance.labels.merge" (list
  (include "nirmata-image-compliance.labels.common" .)
  (include "nirmata-image-compliance.matchLabels.common" .)
) -}}
{{- end -}}

{{- define "nirmata-image-compliance.labels.helm" -}}
helm.sh/chart: {{ template "nirmata-image-compliance.chart" . }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{- define "nirmata-image-compliance.labels.version" -}}
app.kubernetes.io/version: {{ template "nirmata-image-compliance.chartVersion" . }}
{{- end -}}

{{- define "nirmata-image-compliance.labels.common" -}}
{{- template "nirmata-image-compliance.labels.merge" (list
  (include "nirmata-image-compliance.labels.helm" .)
  (include "nirmata-image-compliance.labels.version" .)
) -}}
{{- end -}}

{{- define "nirmata-image-compliance.matchLabels.common" -}}
app.kubernetes.io/part-of: {{ template "nirmata-image-compliance.fullname" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}
