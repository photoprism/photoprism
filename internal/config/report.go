package config

import (
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/dsn"
)

// Report returns global config values as a table for reporting.
func (c *Config) Report() (rows [][]string, cols []string) {
	cols = []string{"Name", "Value"}

	reportDatabaseDSN := c.ReportDatabaseDSN()

	rows = [][]string{
		// Authentication.
		{"auth-mode", c.AuthMode()},
		{"admin-user", c.AdminUser()},
		{"admin-password", strings.Repeat("*", utf8.RuneCountInString(c.AdminPassword()))},
		{"admin-scope", c.AdminScope()},
		{"password-length", fmt.Sprintf("%d", c.PasswordLength())},
		{"password-reset-uri", c.PasswordResetUri()},
		{"register-uri", c.RegisterUri()},
		{"login-uri", c.LoginUri()},
		{"login-info", c.LoginInfo()},
		{"session-maxage", fmt.Sprintf("%d", c.SessionMaxAge())},
		{"session-timeout", fmt.Sprintf("%d", c.SessionTimeout())},
		{"session-cache", fmt.Sprintf("%d", c.SessionCache())},
		{"download-token", c.DownloadToken()},
		{"download-token-maxage", fmt.Sprintf("%d", int64(c.DownloadTokenMaxAge().Seconds()))},
		{"preview-token", c.PreviewToken()},

		// Logging.
		{"log-level", c.LogLevel().String()},
		{"debug", fmt.Sprintf("%t", c.Debug())},
		{"trace", fmt.Sprintf("%t", c.Trace())},

		// Storage.
		{"storage-path", c.StoragePath()},
		{"storage-free", fmt.Sprintf("%.0f", c.StorageFree())},

		// Config.
		{"config-path", c.ConfigPath()},
		{"certificates-path", c.CertificatesPath()},
		{"options-yaml", c.OptionsYaml()},
		{"defaults-yaml", c.DefaultsYaml()},
	}

	// Settings.
	if settingsDefaults := c.SettingsYamlDefaults(""); settingsDefaults != "" && settingsDefaults != c.SettingsYaml() {
		rows = append(rows, []string{"settings-yaml", fmt.Sprintf("%s (defaults)", settingsDefaults)})
	}

	rows = append(rows, [][]string{
		{"settings-yaml", c.SettingsYaml()},

		// Originals.
		{"originals-path", c.OriginalsPath()},
		{"originals-limit", fmt.Sprintf("%d", c.OriginalsLimit())},
		{"resolution-limit", fmt.Sprintf("%d", c.ResolutionLimit())},

		// Other Paths.
		{"users-path", c.UsersPath()},
		{"users-storage-path", c.UsersStoragePath()},
		{"users-originals-path", c.UsersOriginalsPath()},
		{"import-path", c.ImportPath()},
		{"import-dest", c.ImportDest()},
		{"import-allow", c.ImportAllow().String()},
		{"upload-nsfw", fmt.Sprintf("%t", c.UploadNSFW())},
		{"upload-allow", c.UploadAllow().String()},
		{"upload-archives", fmt.Sprintf("%t", c.UploadArchives())},
		{"upload-limit", fmt.Sprintf("%d", c.UploadLimit())},
		{"cache-path", c.CachePath()},
		{"cmd-cache-path", c.CmdCachePath()},
		{"media-cache-path", c.MediaCachePath()},
		{"thumb-cache-path", c.ThumbCachePath()},
		{"temp-path", c.TempPath()},
		{"assets-path", c.AssetsPath()},
		{"models-path", c.ModelsPath()},
		{"static-path", c.StaticPath()},
		{"static-build-path", c.StaticBuildPath()},
		{"static-img-path", c.StaticImgPath()},
		{"templates-path", c.TemplatesPath()},

		// Sidecar Files.
		{"sidecar-path", c.SidecarPath()},
		{"sidecar-yaml", fmt.Sprintf("%t", c.SidecarYaml())},

		// Usage.
		{"usage-info", fmt.Sprintf("%t", c.UsageInfo())},
		{"files-quota", fmt.Sprintf("%d", c.FilesQuota())},
		{"users-quota", fmt.Sprintf("%d", c.UsersQuota())},

		// Backups.
		{"backup-path", c.BackupBasePath()},
		{"backup-schedule", c.BackupSchedule()},
		{"backup-retain", fmt.Sprintf("%d", c.BackupRetain())},
		{"backup-database", fmt.Sprintf("%t", c.BackupDatabase())},
		{"backup-database-path", c.BackupDatabasePath()},
		{"backup-albums", fmt.Sprintf("%t", c.BackupAlbums())},
		{"backup-albums-path", c.BackupAlbumsPath()},

		// Indexing.
		{"index-workers", fmt.Sprintf("%d (%s)", c.IndexWorkers(), c.IndexWorkersReason())},
		{"index-schedule", c.IndexSchedule()},
		{"wakeup-interval", c.WakeupInterval().String()},
		{"auto-index", fmt.Sprintf("%d", c.AutoIndex()/time.Second)},
		{"auto-import", fmt.Sprintf("%d", c.AutoImport()/time.Second)},

		// Feature Flags.
		{"read-only", fmt.Sprintf("%t", c.ReadOnly())},
		{"develop", fmt.Sprintf("%t", c.Develop())},
		{"experimental", fmt.Sprintf("%t", c.Experimental())},
		{"disable-frontend", fmt.Sprintf("%t", c.DisableFrontend())},
		{"disable-settings", fmt.Sprintf("%t", c.DisableSettings())},
		{"disable-backups", fmt.Sprintf("%t", c.DisableBackups())},
		{"disable-restart", fmt.Sprintf("%t", c.DisableRestart())},
		{"disable-webdav", fmt.Sprintf("%t", c.DisableWebDAV())},
		{"disable-mcp", fmt.Sprintf("%t", c.DisableMCP())},
		{"disable-places", fmt.Sprintf("%t", c.DisablePlaces())},
		{"disable-tensorflow", fmt.Sprintf("%t", c.DisableTensorFlow())},
		{"disable-faces", fmt.Sprintf("%t", c.DisableFaces())},
		{"disable-classification", fmt.Sprintf("%t", c.DisableClassification())},
		{"disable-ffmpeg", fmt.Sprintf("%t", c.DisableFFmpeg())},
		{"disable-exiftool", fmt.Sprintf("%t", c.DisableExifTool())},
		{"disable-sips", fmt.Sprintf("%t", c.DisableSips())},
		{"disable-darktable", fmt.Sprintf("%t", c.DisableDarktable())},
		{"disable-rawtherapee", fmt.Sprintf("%t", c.DisableRawTherapee())},
		{"disable-imagemagick", fmt.Sprintf("%t", c.DisableImageMagick())},
		{"disable-heifconvert", fmt.Sprintf("%t", c.DisableHeifConvert())},
		{"disable-rsvgconvert", fmt.Sprintf("%t", c.DisableRsvgConvert())},
		{"disable-vectors", fmt.Sprintf("%t", c.DisableVectors())},
		{"disable-jpegxl", fmt.Sprintf("%t", c.DisableJpegXL())},
		{"disable-raw", fmt.Sprintf("%t", c.DisableRaw())},

		// Format Flags.
		{"raw-presets", fmt.Sprintf("%t", c.RawPresets())},
		{"exif-bruteforce", fmt.Sprintf("%t", c.ExifBruteForce())},

		// Customization.
		{"default-locale", c.DefaultLocale()},
		{"default-timezone", c.DefaultTimezone().String()},
		{"default-theme", c.DefaultTheme()},
	}...)

	if Features == Pro || c.Portal() {
		rows = append(rows, [][]string{
			{"theme-url", c.ThemeUrlRedacted()},
		}...)
	}

	rows = append(rows, [][]string{
		{"places-locale", c.PlacesLocale()},
		{"app-name", c.AppName()},
		{"app-mode", c.AppMode()},
		{"app-icon", c.AppIcon()},
		{"app-color", c.AppColor()},
		{"legal-info", c.LegalInfo()},
		{"legal-url", c.LegalUrl()},
		{"wallpaper-uri", c.WallpaperUri()},

		// Site Infos.
		{"site-url", c.SiteUrl()},
		{"site-https", fmt.Sprintf("%t", c.SiteHttps())},
		{"site-domain", c.SiteDomain()},
		{"site-author", c.SiteAuthor()},
		{"site-name", c.SiteName()},
		{"site-title", c.SiteTitle()},
		{"site-caption", c.SiteCaption()},
		{"site-description", c.SiteDescription()},
		{"site-favicon", c.SiteFavicon()},
		{"site-preview", c.SitePreview()},

		// CDN and Cross-Origin Resource Sharing (CORS).
		{"cdn-url", c.CdnUrl("/")},
		{"cdn-video", fmt.Sprintf("%t", c.CdnVideo())},
		{"cors-origin", c.CORSOrigin()},
		{"cors-headers", c.CORSHeaders()},
		{"cors-methods", c.CORSMethods()},

		// URIs.
		{"base-uri", c.BaseUri("/")},
		{"frontend-uri", c.FrontendUri("")},
		{"api-uri", c.ApiUri()},
		{"static-uri", c.StaticUri()},
		{"content-uri", c.ContentUri()},
		{"video-uri", c.VideoUri()},

		// Cluster Configuration.
		{"cluster-domain", c.ClusterDomain()},
		{"cluster-cidr", c.ClusterCIDR()},
		{"cluster-uuid", c.ClusterUUID()},
		{"cluster-oidc", fmt.Sprintf("%t", c.ClusterOIDC())},
		{"cluster-allow-groups", strings.Join(c.ClusterAllowGroups(), ", ")},
		{"cluster-allow-group-roles", reportGroupRoles(c.ClusterAllowGroupRoles())},
		{"cluster-groups-full-view", fmt.Sprintf("%t", c.ClusterGroupsFullView())},
		{"portal-url", clean.UriRedacted(c.PortalUrl())},
		{"portal-login-url", clean.UriRedacted(c.PortalLoginUrl())},
	}...)

	if c.Portal() {
		rows = append(rows, [][]string{
			{"portal-proxy", fmt.Sprintf("%t", c.PortalProxy())},
			{"portal-proxy-uri", c.PortalProxyUri()},
			{"portal-config-path", c.PortalConfigPath()},
			{"portal-theme-path", c.PortalThemePath()},
		}...)
	}

	rows = append(rows, [][]string{
		{"join-token", strings.Repeat("*", utf8.RuneCountInString(c.JoinToken()))},
		{"node-name", c.NodeName()},
		{"node-role", c.NodeRole()},
		{"node-uuid", c.NodeUUID()},
		{"node-client-id", c.NodeClientID()},
		{"node-client-secret", strings.Repeat("*", utf8.RuneCountInString(c.NodeClientSecret()))},
		{"jwks-url", clean.UriRedacted(c.JWKSUrl())},
		{"jwks-cache-ttl", fmt.Sprintf("%d", c.JWKSCacheTTL())},
		{"jwt-scope", c.JWTAllowedScopes().String()},
		{"jwt-leeway", fmt.Sprintf("%d", c.JWTLeeway())},
		{"advertise-url", clean.UriRedacted(c.AdvertiseUrl())},

		// Networking.
		{"https-proxy", clean.UriRedacted(c.HttpsProxy())},
		{"https-proxy-insecure", fmt.Sprintf("%t", c.HttpsProxyInsecure())},
		{"trusted-platform", c.TrustedPlatform()},
		{"trusted-proxy", c.TrustedProxy()},
		{"proxy-client-header", c.ProxyClientHeader()},
		{"proxy-proto-header", strings.Join(c.ProxyProtoHeader(), ", ")},
		{"proxy-proto-https", strings.Join(c.ProxyProtoHttps(), ", ")},
		{"services-cidr", c.ServicesCIDR()},

		// Web Server.
		{"disable-tls", fmt.Sprintf("%t", c.DisableTLS())},
		{"default-tls", fmt.Sprintf("%t", c.DefaultTLS())},
		{"tls-email", c.TLSEmail()},
		{"tls-cert", c.TLSCert()},
		{"tls-key", c.TLSKey()},
		{"http-mode", c.HttpMode()},
		{"http-compression", c.HttpCompression()},
		{"http-header-timeout", c.HttpHeaderTimeout().String()},
		{"http-header-bytes", fmt.Sprintf("%d", c.HttpHeaderBytes())},
		{"http-idle-timeout", c.HttpIdleTimeout().String()},
		{"http-cache-public", fmt.Sprintf("%t", c.HttpCachePublic())},
		{"http-cache-maxage", fmt.Sprintf("%d", c.HttpCacheMaxAge())},
		{"http-video-maxage", fmt.Sprintf("%d", c.HttpVideoMaxAge())},
		{"http-host", c.HttpHost()},
		{"http-port", fmt.Sprintf("%d", c.HttpPort())},
	}...)

	// Primary Database, Cluster Provision, and ProxySQL Credentials.
	if reportDatabaseDSN {
		rows = append(rows, [][]string{
			{"database-driver", c.DatabaseDriver()},
			{"database-dsn", dsn.Mask(c.DatabaseDSN())},
		}...)
	} else {
		rows = append(rows, [][]string{
			{"database-driver", c.DatabaseDriver()},
			{"database-name", c.DatabaseName()},
			{"database-server", c.DatabaseServer()},
			{"database-host", c.DatabaseHost()},
			{"database-port", c.DatabasePortString()},
			{"database-user", c.DatabaseUser()},
			{"database-password", strings.Repeat("*", utf8.RuneCountInString(c.DatabasePassword()))},
		}...)
	}

	if c.Portal() {
		rows = append(rows, [][]string{
			{"database-provision-driver", c.options.DatabaseProvisionDriver},
			{"database-provision-prefix", c.DatabaseProvisionPrefix()},
			{"database-provision-dsn", dsn.Mask(c.options.DatabaseProvisionDSN)},
			{"database-provision-proxy-dsn", dsn.Mask(c.options.DatabaseProvisionProxyDSN)},
		}...)
	}

	rows = append(rows, [][]string{
		{"database-timeout", fmt.Sprintf("%d", c.DatabaseTimeout())},
		{"database-conns", fmt.Sprintf("%d", c.DatabaseConns())},
		{"database-conns-idle", fmt.Sprintf("%d", c.DatabaseConnsIdle())},
		{"mariadb-bin", c.MariadbBin()},
		{"mariadb-dump-bin", c.MariadbDumpBin()},

		// File Converters.
		{"ffmpeg-bin", c.FFmpegBin()},
		{"ffmpeg-encoder", c.FFmpegEncoder().String()},
		{"ffmpeg-size", fmt.Sprintf("%d", c.FFmpegSize())},
		{"ffmpeg-quality", fmt.Sprintf("%d", c.FFmpegQuality())},
		{"ffmpeg-bitrate", fmt.Sprintf("%d", c.FFmpegBitrate())},
		{"ffmpeg-fisheye-fov", fmt.Sprintf("%d", c.FFmpegFisheyeFov())},
		{"ffmpeg-preset", c.FFmpegPreset()},
		{"ffmpeg-device", c.FFmpegDevice()},
		{"ffmpeg-map-video", c.FFmpegMapVideo()},
		{"ffmpeg-map-audio", c.FFmpegMapAudio()},
		{"ffmpeg-exclude", c.FFmpegExclude().String()},
		{"exiftool-bin", c.ExifToolBin()},
		{"sips-bin", c.SipsBin()},
		{"sips-exclude", c.SipsExclude()},
		{"darktable-bin", c.DarktableBin()},
		{"darktable-cache-path", c.DarktableCachePath()},
		{"darktable-config-path", c.DarktableConfigPath()},
		{"darktable-exclude", c.DarktableExclude()},
		{"rawtherapee-bin", c.RawTherapeeBin()},
		{"rawtherapee-exclude", c.RawTherapeeExclude()},
		{"imagemagick-bin", c.ImageMagickBin()},
		{"imagemagick-exclude", c.ImageMagickExclude()},
		{"heifconvert-bin", c.HeifConvertBin()},
		{"heifconvert-orientation", c.HeifConvertOrientation()},
		{"rsvgconvert-bin", c.RsvgConvertBin()},
		{"jpegxldecoder-bin", c.JpegXLDecoderBin()},

		// Thumbnails.
		{"thumb-library", c.ThumbLibrary()},
		{"thumb-color", c.ThumbColor()},
		{"thumb-size", fmt.Sprintf("%d", c.ThumbSizePrecached())},
		{"thumb-size-uncached", fmt.Sprintf("%d", c.ThumbSizeUncached())},
		{"thumb-uncached", fmt.Sprintf("%t", c.ThumbUncached())},
		{"jpeg-quality", fmt.Sprintf("%d", c.JpegQuality())},
		{"jpeg-size", fmt.Sprintf("%d", c.JpegSize())},
		{"png-size", fmt.Sprintf("%d", c.PngSize())},

		// Computer Vision & Facial Recognition.
		{"vision-yaml", c.VisionYaml()},
		{"vision-api", fmt.Sprintf("%t", c.VisionApi())},
		{"vision-uri", clean.UriRedacted(c.VisionUri())},
		{"vision-key", strings.Repeat("*", utf8.RuneCountInString(c.VisionKey()))},
		{"vision-schedule", c.VisionSchedule()},
		{"vision-filter", c.VisionFilter()},
		{"nasnet-model-path", c.NasnetModelPath()},
		{"facenet-model-path", c.FacenetModelPath()},
		{"nsfw-model-path", c.NsfwModelPath()},
		{"detect-nsfw", fmt.Sprintf("%t", c.DetectNSFW())},
		{"xmp-faces", fmt.Sprintf("%t", c.XMPFaces())},
	}...)

	for _, row := range c.faceConfigRows(false) {
		rows = append(rows, []string{row.Flag, row.Value})
	}

	rows = append(rows, [][]string{
		// Daemon Mode.
		{"pid-filename", c.PIDFilename()},
		{"log-filename", c.LogFilename()},
	}...)

	if v := c.CustomAssetsPath(); v != "" {
		rows = append(rows, []string{"custom-assets-path", v})
	}

	if v := c.CustomStaticUri(); v != "" {
		rows = append(rows, []string{"custom-static-uri", v})
	}

	return rows, cols
}

// faceConfigRow is one row of the face detection and recognition configuration. Flag is the
// option name `photoprism show config` reports, Label the heading `photoprism faces status`
// gives the same value, which drops the prefix every row in that report would repeat.
type faceConfigRow struct {
	Flag  string
	Label string
	Value string
}

// faceConfigRows returns the face configuration in the order the Options struct and flags.go
// declare the options, so both reports read like the flag list rather than like each other.
//
// Verbose adds the qualifiers only the faces report shows. `show config` states the value alone:
// it runs without a database and can neither detect the model nor see that one is blocked.
func (c *Config) faceConfigRows(verbose bool) []faceConfigRow {
	detector := c.FaceDetector()
	model := c.EffectiveFaceModel()

	if verbose {
		detector = c.faceDetectorReport()
		model = c.faceModelReport()
	}

	return []faceConfigRow{
		{"face-detector", "Detector", detector},
		{"face-detector-path", "Detector Path", c.FaceEngineModelPath()},
		{"face-detector-threads", "Detector Threads", fmt.Sprintf("%d", c.FaceDetectorThreads())},
		{"face-model", "Model", model},
		{"face-model-path", "Model Path", c.FaceModelPath()},
		{"face-model-threads", "Model Threads", fmt.Sprintf("%d", c.FaceModelThreads())},
		{"face-run", "Schedule", vision.ReportRunType(c.FaceEngineRunType())},
		{"face-engine", "Engine", c.faceEngineReport()},
		{"face-size", "Min Size", fmt.Sprintf("%d", c.FaceSize())},
		{"face-size-retry", "Retry Size", fmt.Sprintf("%d", c.FaceSizeRetry())},
		{"face-score", "Min Score", c.faceScoreReport(verbose)},
		{"face-migrate-size", "Migration Size", fmt.Sprintf("%d", c.FaceMigrateSize())},
		{"face-migrate-score", "Migration Score", fmt.Sprintf("%g", c.FaceMigrateScore())},
		{"face-overlap", "Overlap", fmt.Sprintf("%d", c.FaceOverlap())},
		{"face-cluster-size", "Cluster Size", fmt.Sprintf("%d", c.FaceClusterSize())},
		{"face-cluster-score", "Cluster Score", c.faceClusterScoreReport(verbose)},
		{"face-cluster-core", "Cluster Core", fmt.Sprintf("%d", c.FaceClusterCore())},
		{"face-cluster-dist", "Cluster Distance", c.faceDistReport(c.FaceClusterDist)},
		{"face-cluster-radius", "Cluster Radius", c.faceDistReport(c.FaceClusterRadius)},
		{"face-collision-dist", "Collision Distance", c.faceDistReport(c.FaceCollisionDist)},
		{"face-epsilon-dist", "Epsilon Distance", c.faceDistReport(c.FaceEpsilonDist)},
		{"face-match-dist", "Match Distance", c.faceDistReport(c.FaceMatchDist)},
	}
}

// faceModelStatus reports why a model that is otherwise in force is generating no embeddings, or
// "" when it is. One that failed to load reports ModelNone everywhere else, so a report would
// name a model while nothing is being embedded. Commands that only read the configuration never
// load the model, so the error is reported when one is known rather than assumed.
func (c *Config) faceModelStatus() string {
	if err := face.EmbedderError(); err != nil {
		return fmt.Sprintf("failed to load: %s", err)
	} else if reason := face.EmbeddingsBlockedReason(); reason != "" {
		return fmt.Sprintf("paused: %s", reason)
	}

	return ""
}

// faceScoreReport formats the detection cutoff in force, which is the detector's own whenever
// FACE_SCORE is unset. Verbose names where it came from, because the same number means a
// calibration in one case and an operator's decision in the other.
func (c *Config) faceScoreReport(verbose bool) string {
	score := fmt.Sprintf("%g", c.FaceScoreEffective())

	if !verbose || c.FaceScore() != face.ScoreThresholdDefault {
		return score
	}

	return fmt.Sprintf("%s (%s)", score, c.FaceDetector())
}

// faceClusterScoreReport formats the clustering bar in force. Verbose names the detector it was
// calibrated for, because the bar is looked up per marker: a library holding markers from more
// than one detector applies more than one bar, and the row can only state the current one.
func (c *Config) faceClusterScoreReport(verbose bool) string {
	score := fmt.Sprintf("%d", c.FaceClusterScoreEffective())

	if !verbose || c.FaceClusterScore() != 0 {
		return score
	}

	return fmt.Sprintf("%s (%s)", score, c.FaceDetector())
}

// faceDistReport formats a face distance threshold, or reports nothing when no embedding model
// is in force. The calibrated distances are per model and are not comparable between them, so a
// report with no model would otherwise print five numbers that will not apply to the model the
// instance ends up using.
func (c *Config) faceDistReport(value func() float64) string {
	if c.EffectiveFaceModel() == face.ModelNone {
		return ""
	}

	return fmt.Sprintf("%f", value())
}

// faceEngineReport names the deprecated runtime setting as it is configured, rather than the
// runtime in force, which now follows the detector and is already reported as one. A stale
// "none" left in "options.yml" is the only thing that explains why detection is off, so the
// row stays even though the option no longer selects anything.
func (c *Config) faceEngineReport() string {
	return fmt.Sprintf("%s (deprecated)", face.ParseEngine(c.options.FaceEngine))
}

// faceDetectorReport names the detector in force and where that name came from.
func (c *Config) faceDetectorReport() string {
	resolved := face.DetectorDisplayName(c.FaceDetector())

	switch setting := c.FaceDetectorSetting(); {
	case setting == face.DetectorAuto && c.FaceDetector() != face.DetectorNone:
		return faceReportValue(resolved, "default")
	case setting != face.DetectorAuto && setting != c.FaceDetector():
		return faceReportValue(resolved, fmt.Sprintf("%s is not available", clean.Log(setting)))
	default:
		return resolved
	}
}

// faceModelReport names the embedding model in force, where that name came from, and why
// embeddings are not being generated when they are not. `faces status` connects to the database,
// so a detected model is there the one the library holds rather than a fresh install's default.
func (c *Config) faceModelReport() string {
	inForce := c.EffectiveFaceModel()
	resolved := face.ModelDisplayName(inForce)
	setting := c.FaceModelSetting()

	// Nothing is in force, so the reason is the whole row: which of the three it is decides
	// whether an operator has a decision to revisit, an install to fix, or nothing to do.
	switch {
	case setting == face.ModelNone:
		return faceReportValue(resolved, "embeddings disabled")
	case inForce != face.ModelNone:
		break
	case setting == face.ModelDetect:
		return faceReportValue(resolved, "no embedding model is installed")
	default:
		return faceReportValue(resolved, fmt.Sprintf("%s is not available", clean.Log(setting)))
	}

	var notes []string

	if setting == face.ModelDetect {
		notes = append(notes, "default")
	}

	if status := c.faceModelStatus(); status != "" {
		notes = append(notes, status)
	}

	return faceReportValue(resolved, notes...)
}

// faceReportValue appends the qualifiers a report shows in parentheses after a resolved value.
func faceReportValue(value string, notes ...string) string {
	if len(notes) == 0 {
		return value
	}

	return fmt.Sprintf("%s (%s)", value, strings.Join(notes, ", "))
}

// FaceReport returns the face detection and recognition configuration as a table for
// `photoprism faces status`. It covers the same options as Report() in the same order, but
// names them without the prefix the subcommand already carries and states what a database
// connection adds: the model the library holds, and whether embeddings are paused.
func (c *Config) FaceReport() (rows [][]string, cols []string) {
	cols = []string{"Name", "Value"}

	rows = [][]string{
		{"Enabled", fmt.Sprintf("%t", !c.DisableFaces())},
		{"XMP Faces", fmt.Sprintf("%t", c.XMPFaces())},
		{"Vision Config", c.VisionYaml()},
	}

	for _, row := range c.faceConfigRows(true) {
		rows = append(rows, []string{row.Label, row.Value})
	}

	// Only shown while one is held. A lock nothing reports is the difference between an instance
	// that explains why it indexes nothing and one that has to be diagnosed by finding a file.
	if held := c.FacesLocked(); held != "" {
		rows = append(rows, []string{"Locked", held})
	}

	return rows, cols
}

// reportGroupRoles renders a group → role mapping as sorted "group=role" pairs
// for the config report.
func reportGroupRoles(roles map[string]string) string {
	if len(roles) == 0 {
		return ""
	}

	pairs := make([]string, 0, len(roles))

	for group, role := range roles {
		pairs = append(pairs, group+"="+role)
	}

	sort.Strings(pairs)

	return strings.Join(pairs, ", ")
}
