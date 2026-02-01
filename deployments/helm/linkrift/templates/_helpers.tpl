{{/*
Common labels
*/}}
{{- define "linkrift.labels" -}}
helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version | replace "+" "_" }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/part-of: linkrift
{{- end }}

{{/*
Selector labels for a component
*/}}
{{- define "linkrift.selectorLabels" -}}
app.kubernetes.io/name: {{ .name }}
app.kubernetes.io/instance: {{ .release }}
{{- end }}

{{/*
Full image reference
*/}}
{{- define "linkrift.image" -}}
{{- $registry := .global.imageRegistry -}}
{{- $repository := .image.repository -}}
{{- $tag := .image.tag | default .global.imageTag | default .chartVersion -}}
{{ printf "%s/%s:%s" $registry $repository $tag }}
{{- end }}

{{/*
Service account name
*/}}
{{- define "linkrift.serviceAccountName" -}}
{{- if .Values.serviceAccount.name -}}
{{ .Values.serviceAccount.name }}
{{- else -}}
{{ .Release.Name }}-linkrift
{{- end -}}
{{- end }}
