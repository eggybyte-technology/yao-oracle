{{/*
Expand the name of the chart.
*/}}
{{- define "yao-oracle.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "yao-oracle.fullname" -}}
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
{{- define "yao-oracle.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "yao-oracle.labels" -}}
helm.sh/chart: {{ include "yao-oracle.chart" . }}
{{ include "yao-oracle.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- with .Values.global.labels }}
{{ toYaml . }}
{{- end }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "yao-oracle.selectorLabels" -}}
app.kubernetes.io/name: {{ include "yao-oracle.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "yao-oracle.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "yao-oracle.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Proxy component labels
*/}}
{{- define "yao-oracle.proxy.labels" -}}
{{ include "yao-oracle.labels" . }}
app.kubernetes.io/component: proxy
{{- end }}

{{/*
Proxy selector labels
*/}}
{{- define "yao-oracle.proxy.selectorLabels" -}}
{{ include "yao-oracle.selectorLabels" . }}
app.kubernetes.io/component: proxy
{{- end }}

{{/*
Node component labels
*/}}
{{- define "yao-oracle.node.labels" -}}
{{ include "yao-oracle.labels" . }}
app.kubernetes.io/component: node
{{- end }}

{{/*
Node selector labels
*/}}
{{- define "yao-oracle.node.selectorLabels" -}}
{{ include "yao-oracle.selectorLabels" . }}
app.kubernetes.io/component: node
{{- end }}

{{/*
Dashboard component labels
*/}}
{{- define "yao-oracle.dashboard.labels" -}}
{{ include "yao-oracle.labels" . }}
app.kubernetes.io/component: dashboard
{{- end }}

{{/*
Dashboard selector labels
*/}}
{{- define "yao-oracle.dashboard.selectorLabels" -}}
{{ include "yao-oracle.selectorLabels" . }}
app.kubernetes.io/component: dashboard
{{- end }}

{{/*
Image pull secrets
*/}}
{{- define "yao-oracle.imagePullSecrets" -}}
{{- if .Values.global.imagePullSecrets }}
imagePullSecrets:
{{- range .Values.global.imagePullSecrets }}
  - name: {{ . }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Return the proper image name for proxy
*/}}
{{- define "yao-oracle.proxy.image" -}}
{{- $registry := .Values.global.imageRegistry -}}
{{- $repository := .Values.proxy.image.repository -}}
{{- $tag := .Values.proxy.image.tag | default .Chart.AppVersion -}}
{{- printf "%s/%s:%s" $registry $repository $tag -}}
{{- end }}

{{/*
Return the proper image name for node
*/}}
{{- define "yao-oracle.node.image" -}}
{{- $registry := .Values.global.imageRegistry -}}
{{- $repository := .Values.node.image.repository -}}
{{- $tag := .Values.node.image.tag | default .Chart.AppVersion -}}
{{- printf "%s/%s:%s" $registry $repository $tag -}}
{{- end }}

{{/*
Return the proper image name for dashboard
*/}}
{{- define "yao-oracle.dashboard.image" -}}
{{- $registry := .Values.global.imageRegistry -}}
{{- $repository := .Values.dashboard.image.repository -}}
{{- $tag := .Values.dashboard.image.tag | default .Chart.AppVersion -}}
{{- printf "%s/%s:%s" $registry $repository $tag -}}
{{- end }}

{{/*
Generate node addresses for consistent hashing
*/}}
{{- define "yao-oracle.nodeAddresses" -}}
{{- $fullname := include "yao-oracle.fullname" . -}}
{{- $nodeCount := int .Values.node.replicaCount -}}
{{- $nodePort := int .Values.node.service.grpcPort -}}
{{- range $i := until $nodeCount }}
{{- printf "%s-node-%d.%s-node:%d\n" $fullname $i $fullname $nodePort -}}
{{- end }}
{{- end }}

