{{/*
General helpers for the PhotoPrism Plus chart
*/}}
{{- define "photoprism-plus.name" -}}
{{- if .Values.nameOverride -}}
{{- .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- "photoprism" -}}
{{- end -}}
{{- end -}}

{{- define "photoprism-plus.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-photoprism" .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}

{{- define "photoprism-plus.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version -}}
{{- end -}}

{{- define "photoprism-plus.labels" -}}
helm.sh/chart: {{ include "photoprism-plus.chart" . }}
app.kubernetes.io/name: {{ include "photoprism-plus.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{- define "photoprism-plus.clusterSecretName" -}}
{{- if .Values.cluster.integration.secretName -}}
{{- .Values.cluster.integration.secretName -}}
{{- else -}}
photoprism-cluster-secrets
{{- end -}}
{{- end -}}

{{- define "photoprism-plus.clusterDomain" -}}
{{- $domain := default "" .Values.cluster.integration.domain | trim -}}
{{- if and (eq $domain "") .Values.cluster.integration.enabled -}}
  {{- $secret := lookup "v1" "Secret" .Release.Namespace (include "photoprism-plus.clusterSecretName" .) -}}
  {{- if $secret -}}
    {{- $encoded := index $secret.data "PHOTOPRISM_CLUSTER_DOMAIN" -}}
    {{- if $encoded -}}
      {{- $domain = b64dec $encoded | trim -}}
    {{- end -}}
  {{- end -}}
{{- end -}}
{{- $domain -}}
{{- end -}}

{{- define "photoprism-plus.siteURL" -}}
{{- $site := default "" .Values.config.PHOTOPRISM_SITE_URL | trim -}}
{{- if and (eq $site "") .Values.cluster.integration.enabled -}}
  {{- $domain := include "photoprism-plus.clusterDomain" . | trim -}}
  {{- if ne $domain "" -}}
    {{- printf "https://%s.%s/" .Release.Name $domain -}}
  {{- end -}}
{{- else -}}
  {{- $site -}}
{{- end -}}
{{- end -}}
