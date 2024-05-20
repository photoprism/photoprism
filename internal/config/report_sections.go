package config

// ReportSection represents a group of config options.
type ReportSection struct {
	Start string
	Title string
	Info  string
}

var faceFlagsInfo = `!!! info ""
    To [recognize faces](https://docs.photoprism.app/user-guide/organize/people/), PhotoPrism first extracts crops from your images using a [library](https://github.com/esimov/pigo) based on [pixel intensity comparisons](https://dl.photoprism.app/pdf/20140820-Pixel_Intensity_Comparisons.pdf). These are then fed into TensorFlow to compute [512-dimensional vectors](https://dl.photoprism.app/pdf/20150101-FaceNet.pdf) for characterization. In the final step, the [DBSCAN algorithm](https://en.wikipedia.org/wiki/DBSCAN) attempts to cluster these so-called face embeddings, so they can be matched to persons with just a few clicks. A reasonable range for the similarity distance between face embeddings is between 0.60 and 0.70, with a higher value being more aggressive and leading to larger clusters with more false positives. To cluster a smaller number of faces, you can reduce the core to 3 or 2 similar faces.

We recommend that only advanced users change these parameters:`

// OptionsReportSections is used to generate config options reports in ../commands/show_config_options.go.
var OptionsReportSections = []ReportSection{
	{Start: "PHOTOPRISM_ADMIN_PASSWORD", Title: "Authentication"},
	{Start: "PHOTOPRISM_LOG_LEVEL", Title: "Logging"},
	{Start: "PHOTOPRISM_CONFIG_PATH", Title: "Storage"},
	{Start: "PHOTOPRISM_BACKUP_PATH", Title: "Backup"},
	{Start: "PHOTOPRISM_INDEX_WORKERS, PHOTOPRISM_WORKERS", Title: "Index Workers"},
	{Start: "PHOTOPRISM_READONLY", Title: "Feature Flags"},
	{Start: "PHOTOPRISM_DEFAULT_LOCALE", Title: "Customization"},
	{Start: "PHOTOPRISM_SITE_URL", Title: "Site Information"},
	{Start: "PHOTOPRISM_HTTPS_PROXY", Title: "Proxy Servers"},
	{Start: "PHOTOPRISM_DISABLE_TLS", Title: "Web Server"},
	{Start: "PHOTOPRISM_DATABASE_DRIVER", Title: "Database Connection"},
	{Start: "PHOTOPRISM_SIPS_BIN", Title: "File Conversion"},
	{Start: "PHOTOPRISM_DOWNLOAD_TOKEN", Title: "Security Tokens"},
	{Start: "PHOTOPRISM_THUMB_LIBRARY", Title: "Preview Images"},
	{Start: "PHOTOPRISM_JPEG_QUALITY", Title: "Image Quality"},
	{Start: "PHOTOPRISM_FACE_SIZE", Title: "Face Recognition",
		Info: faceFlagsInfo},
	{Start: "PHOTOPRISM_PID_FILENAME", Title: "Daemon Mode",
		Info: "If you start the server as a *daemon* in the background, you can additionally specify a filename for the log and the process ID:"},
}

// YamlReportSections is used to generate config options reports in ../commands/show_config_yaml.go.
var YamlReportSections = []ReportSection{
	{Start: "AuthMode", Title: "Authentication"},
	{Start: "LogLevel", Title: "Logging"},
	{Start: "ConfigPath", Title: "Storage"},
	{Start: "BackupPath", Title: "Backup"},
	{Start: "IndexWorkers", Title: "Index Workers"},
	{Start: "ReadOnly", Title: "Feature Flags"},
	{Start: "DefaultTheme", Title: "Customization"},
	{Start: "SiteUrl", Title: "Site Information"},
	{Start: "HttpsProxy", Title: "Web Server"},
	{Start: "DatabaseDriver", Title: "Database Connection"},
	{Start: "SipsBin", Title: "File Conversion"},
	{Start: "DownloadToken", Title: "Security Tokens"},
	{Start: "ThumbLibrary", Title: "Preview Images"},
	{Start: "JpegQuality", Title: "Image Quality"},
	{Start: "PIDFilename", Title: "Daemon Mode",
		Info: "If you start the server as a *daemon* in the background, you can additionally specify a filename for the log and the process ID:"},
}
