{{- define "l2keeper.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "l2keeper.fullname" -}}
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

{{- define "l2keeper.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "l2keeper.labels" -}}
helm.sh/chart: {{ include "l2keeper.chart" . }}
{{ include "l2keeper.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{- define "l2keeper.selectorLabels" -}}
app.kubernetes.io/name: {{ include "l2keeper.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{- define "l2keeper.controllerSelectorLabels" -}}
{{ include "l2keeper.selectorLabels" . }}
app.kubernetes.io/component: controller
{{- end -}}

{{- define "l2keeper.speakerSelectorLabels" -}}
{{ include "l2keeper.selectorLabels" . }}
app.kubernetes.io/component: speaker
{{- end -}}

{{- define "l2keeper.controllerImage" -}}
{{- printf "%s:%s" .Values.controller.image.repository (default .Chart.AppVersion .Values.controller.image.tag) -}}
{{- end -}}

{{- define "l2keeper.speakerImage" -}}
{{- printf "%s:%s" .Values.speaker.image.repository (default .Chart.AppVersion .Values.speaker.image.tag) -}}
{{- end -}}
