package config

// ReportSection represents a group of config options.
type ReportSection struct {
	Start string
	Title string
	Info  string
}

// see https://docs.photoprism.app/getting-started/config-options/#face-recognition
var faceFlagsInfo = `!!! info ""
    The distance thresholds default to values calibrated for the configured face model, so leaving them unset is recommended: a range that suits one embedding model does not transfer to another. Higher values cluster more aggressively and produce more false positives. To cluster a smaller number of faces, reduce the core to 3 or 2 similar faces. After changing any of the clustering parameters, run "photoprism faces update --force" so existing clusters are recalculated; changing the face model instead requires "photoprism faces migrate".

We recommend that only advanced users change these parameters:`

// OptionsReportSections is used to generate config options reports in ../commands/show_config_options.go.
var OptionsReportSections = []ReportSection{
	{Start: "PHOTOPRISM_AUTH_MODE", Title: "Authentication"},
	{Start: "PHOTOPRISM_LOG_LEVEL", Title: "Logging"},
	{Start: "PHOTOPRISM_STORAGE_PATH", Title: "Storage"},
	{Start: "PHOTOPRISM_SIDECAR_PATH", Title: "Sidecar Files"},
	{Start: "PHOTOPRISM_USAGE_INFO", Title: "Usage"},
	{Start: "PHOTOPRISM_BACKUP_PATH", Title: "Backup"},
	{Start: "PHOTOPRISM_INDEX_WORKERS, PHOTOPRISM_WORKERS", Title: "Indexing"},
	{Start: "PHOTOPRISM_READONLY", Title: "Feature Flags"},
	{Start: "PHOTOPRISM_DEFAULT_LOCALE", Title: "Customization"},
	{Start: "PHOTOPRISM_SITE_URL", Title: "Site Information"},
	{Start: "PHOTOPRISM_CLUSTER_DOMAIN", Title: "Cluster Configuration"},
	{Start: "PHOTOPRISM_HTTPS_PROXY", Title: "Networking"},
	{Start: "PHOTOPRISM_DISABLE_TLS", Title: "Web Server"},
	{Start: "PHOTOPRISM_DATABASE_DRIVER", Title: "Database Connection"},
	{Start: "PHOTOPRISM_FFMPEG_BIN", Title: "File Conversion"},
	{Start: "PHOTOPRISM_THUMB_LIBRARY", Title: "Preview Images"},
	{Start: "PHOTOPRISM_JPEG_QUALITY", Title: "Image Quality"},
	{Start: "PHOTOPRISM_VISION_YAML", Title: "Computer Vision"},
	{Start: "PHOTOPRISM_FACE_DETECTOR", Title: "Face Recognition",
		Info: faceFlagsInfo},
	{Start: "PHOTOPRISM_PID_FILENAME", Title: "Daemon Mode",
		Info: "If you start the server as a *daemon* in the background, you can additionally specify a filename for the log and the process ID:"},
}

// YamlReportSections is used to generate config options reports in ../commands/show_config_yaml.go.
var YamlReportSections = []ReportSection{
	{Start: "AuthMode", Title: "Authentication"},
	{Start: "LogLevel", Title: "Logging"},
	{Start: "StoragePath", Title: "Storage"},
	{Start: "SidecarPath", Title: "Sidecar Files"},
	{Start: "UsageInfo", Title: "Usage"},
	{Start: "BackupPath", Title: "Backup"},
	{Start: "IndexWorkers", Title: "Indexing"},
	{Start: "ReadOnly", Title: "Feature Flags"},
	{Start: "DefaultLocale", Title: "Customization"},
	{Start: "SiteUrl", Title: "Site Information"},
	{Start: "ClusterDomain", Title: "Cluster Configuration"},
	{Start: "HttpsProxy", Title: "Networking"},
	{Start: "DisableTLS", Title: "Web Server"},
	{Start: "DatabaseDriver", Title: "Database Connection"},
	{Start: "FFmpegBin", Title: "File Conversion"},
	{Start: "ThumbLibrary", Title: "Preview Images"},
	{Start: "JpegQuality", Title: "Image Quality"},
	{Start: "VisionYaml", Title: "Computer Vision"},
	{Start: "FaceDetector", Title: "Face Recognition"},
	{Start: "PIDFilename", Title: "Daemon Mode",
		Info: "If you start the server as a *daemon* in the background, you can additionally specify a filename for the log and the process ID:"},
}
