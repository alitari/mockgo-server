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
Return the proper mockgoserver image name
*/}}
{{- define "mockgoserver.clusterAdvertise" -}}
{{- if $.Values.useFQDN }}
{{- printf "$(POD_NAME).%s.$(POD_NAMESPACE).svc.%s" (include "mockgoserver.fullname" . ) $.Values.k8sClusterDomain }}
{{- else }}
{{- printf "$(POD_NAME).%s.$(POD_NAMESPACE)" (include "mockgoserver.fullname" . ) }}
{{- end }}
{{- end }}

{{/*
Return the mockgoserver cluster auth.
*/}}
{{- define "mockgoserver.clusterAuth" -}}
{{- if $.Values.cluster.authorization }}
{{- printf "%s:%s@" (urlquery $.Values.cluster.authorization.user) (urlquery $.Values.cluster.authorization.password) -}}
{{- else }}
{{- end }}
{{- end }}

{{/*
Return the mockgoserver cluster routes.
*/}}
{{- define "mockgoserver.clusterUrls" -}}
{{- if .Values.cluster.enabled }}
{{- $name := (include "mockgoserver.fullname" . ) -}}
{{- $namespace := (include "mockgoserver.namespace" . ) -}}
{{- range $i, $e := until (.Values.cluster.replicas | int) -}}
{{- printf "http://%s-%d.%s.%s:8080," $name $i $name $namespace -}}
{{- end }}
{{- end }}
{{- end }}




{{/*
Return the appropriate apiVersion for networkpolicy.
*/}}
{{- define "networkPolicy.apiVersion" -}}
{{- if semverCompare ">=1.4-0, <1.7-0" .Capabilities.KubeVersion.GitVersion -}}
{{- print "extensions/v1beta1" -}}
{{- else -}}
{{- print "networking.k8s.io/v1" -}}
{{- end -}}
{{- end -}}

{{/*
Renders a value that contains template.
Usage:
{{ include "tplvalues.render" ( dict "value" .Values.path.to.the.Value "context" $) }}
*/}}
{{- define "tplvalues.render" -}}
  {{- if typeIs "string" .value }}
    {{- tpl .value .context }}
  {{- else }}
    {{- tpl (toYaml .value) .context }}
  {{- end }}
{{- end -}}


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