{{/*
Expand the name of the chart.
*/}}
{{- define "mockgoserver.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "mockgoserver.namespace" -}}
{{- default .Release.Namespace .Values.namespaceOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}


{{- define "mockgoserver.fullname" -}}
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

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "mockgoserver.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "mockgoserver.labels" -}}
helm.sh/chart: {{ include "mockgoserver.chart" . }}
{{- range $name, $value := .Values.commonLabels }}
{{ $name }}: {{ tpl $value $ }}
{{- end }}
{{ include "mockgoserver.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "mockgoserver.selectorLabels" -}}
{{- if .Values.selectorLabels }}
{{ tpl (toYaml .Values.selectorLabels) . }}
{{- else -}}
app.kubernetes.io/name: {{ include "mockgoserver.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
{{- end }}

{{/*
Return the mockgoserver cluster routes.
*/}}
{{- define "mockgoserver.clusterUrls" -}}
{{- if .Values.cluster.enabled }}
{{- $name := (include "mockgoserver.fullname" . ) -}}
{{- $namespace := (include "mockgoserver.namespace" . ) -}}
{{- $clusterDomain := .Values.k8sClusterDomain -}}
{{- $port := .Values.configEndpoint.servicePort -}}

{{- range $i, $e := until (.Values.cluster.replicas | int) -}}
{{- if $.Values.cluster.useFQDN }}
{{- printf "http://%s-%d.%s.%s.svc.%s:%.0f," $name $i $name $namespace $clusterDomain $port -}}
{{- else }}
{{- printf "http://%s-%d.%s.%s.:%.0f," $name $i $name $namespace $port -}}
{{- end }}
{{- end }}
{{- end }}
{{- end }}


{{/*
Create the name of the service account to use
*/}}
{{- define "mockgoserver.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "mockgoserver.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}