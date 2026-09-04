package config

import (
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/internal/entity/query"
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
		{"jwt-rotate-days", fmt.Sprintf("%d", c.JWTRotateDays())},
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
		{"thumb-size-face", fmt.Sprintf("%d", c.ThumbSizeFace())},
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
		{"label-model", string(c.EffectiveLabelModel())},
		{"label-model-path", c.LabelModelPath()},
		{"label-model-runtime", c.LabelModelRuntime()},
		{"nasnet-model-path", c.NasnetModelPath()},
		{"facenet-model-path", c.FacenetModelPath()},
		{"nsfw-model-path", c.NsfwModelPath()},
		{"detect-nsfw", fmt.Sprintf("%t", c.DetectNSFW())},
	}...)

	for _, row := range c.faceConfigRows() {
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

// faceConfigSection names the group a face option belongs to. `photoprism faces status` renders
// one table per section, and `photoprism show config` one flat list in the same order.
type faceConfigSection string

const (
	faceSectionGlobal      faceConfigSection = "Global Options"
	faceSectionDetection   faceConfigSection = "Face Detection"
	faceSectionRecognition faceConfigSection = "Face Recognition"
)

// faceConfigRow is one row of the face detection and recognition configuration, named by the
// option `photoprism show config` reports so that both reports name a value the same way.
type faceConfigRow struct {
	Section faceConfigSection
	Flag    string
	Value   string
}

// faceConfigRows returns the face configuration in the order the Options struct and flags.go
// declare the options, so both reports read like the flag list rather than like each other.
//
// Values only: where a number came from, and why nothing is being processed, belong in the notes
// beside the tables, because `show config` runs without a database and can check neither.
func (c *Config) faceConfigRows() []faceConfigRow {
	rows := []faceConfigRow{
		{faceSectionGlobal, "xmp-faces", fmt.Sprintf("%t", c.XMPFaces())},
		{faceSectionGlobal, "face-run", vision.ReportRunType(c.FaceEngineRunType())},
	}

	// Reported only while it still decides something: the option no longer selects a runtime, but
	// a "none" left in "options.yml" is the one thing that explains why detection is off.
	if face.ParseEngine(c.options.FaceEngine) != face.EngineAuto {
		rows = append(rows, faceConfigRow{faceSectionGlobal, "face-engine", c.faceEngineReport()})
	}

	return append(rows, []faceConfigRow{
		{faceSectionDetection, "face-detector", c.FaceDetector()},
		{faceSectionDetection, "face-detector-path", c.FaceEngineModelPath()},
		{faceSectionDetection, "face-detector-threads", fmt.Sprintf("%d", c.FaceDetectorThreads())},
		{faceSectionDetection, "face-size", fmt.Sprintf("%d", c.FaceSize())},
		{faceSectionDetection, "face-size-retry", fmt.Sprintf("%d", c.FaceSizeRetry())},
		{faceSectionDetection, "face-score", fmt.Sprintf("%g", c.FaceScoreEffective())},
		{faceSectionDetection, "face-migrate-size", fmt.Sprintf("%d", c.FaceMigrateSize())},
		{faceSectionDetection, "face-migrate-score", fmt.Sprintf("%g", c.FaceMigrateScore())},
		{faceSectionDetection, "face-overlap", fmt.Sprintf("%d", c.FaceOverlap())},
		{faceSectionRecognition, "face-model", c.EffectiveFaceModel()},
		{faceSectionRecognition, "face-model-path", c.FaceModelPath()},
		{faceSectionRecognition, "face-model-threads", fmt.Sprintf("%d", c.FaceModelThreads())},
		{faceSectionRecognition, "face-cluster-size", fmt.Sprintf("%d", c.FaceClusterSize())},
		{faceSectionRecognition, "face-cluster-score", fmt.Sprintf("%d", c.FaceClusterScoreEffective())},
		{faceSectionRecognition, "face-cluster-core", fmt.Sprintf("%d", c.FaceClusterCore())},
		{faceSectionRecognition, "face-cluster-split-rounds", fmt.Sprintf("%d", c.FaceClusterSplitRounds())},
		{faceSectionRecognition, "face-cluster-split-shrink", fmt.Sprintf("%g", c.FaceClusterSplitShrink())},
		{faceSectionRecognition, "face-cluster-dist", c.faceDistReport(c.FaceClusterDist)},
		{faceSectionRecognition, "face-cluster-radius", c.faceDistReport(c.FaceClusterRadius)},
		{faceSectionRecognition, "face-cluster-percentile", fmt.Sprintf("%d", c.FaceClusterPercentile())},
		{faceSectionRecognition, "face-match-dist", c.faceDistReport(c.FaceMatchDist)},
		{faceSectionRecognition, "face-match-margin", c.faceDistReport(c.FaceMatchMargin)},
		{faceSectionRecognition, "face-collision-dist", c.faceDistReport(c.FaceCollisionDist)},
		{faceSectionRecognition, "face-epsilon-dist", c.faceDistReport(c.FaceEpsilonDist)},
		{faceSectionRecognition, "face-recompute-stats", fmt.Sprintf("%t", c.FaceRecomputeStats())},
	}...)
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

// faceDetectionNote states what the detection table cannot: which detector "auto" resolved to,
// and that a cutoff no option holds was calibrated for it rather than chosen.
func (c *Config) faceDetectionNote() string {
	notes := []string{fmt.Sprintf("Detector: %s.", c.faceDetectorReport())}

	if c.FaceDetector() == face.DetectorNone {
		return notes[0]
	}

	if c.FaceScore() == face.ScoreThresholdDefault {
		notes = append(notes, fmt.Sprintf("The minimum score of %g is calibrated for %s.",
			c.FaceScoreEffective(), c.FaceDetector()))
	}

	return strings.Join(notes, " ")
}

// faceRecognitionNote states which model is in force, why it is generating nothing when it is
// not, and that the clustering bar follows the detector rather than the model beside it.
func (c *Config) faceRecognitionNote() string {
	notes := []string{fmt.Sprintf("Model: %s.", c.faceModelReport())}

	if c.EffectiveFaceModel() == face.ModelNone {
		// The distances are calibrated per model and are not comparable between them, so their
		// rows are blank rather than showing five numbers the model finally used will not apply.
		return notes[0] + " The calibrated distances are reported once a model is in force."
	}

	if c.FaceClusterScore() == 0 {
		// Named as the bar for markers this detector scored, not as the bar in force: a library
		// whose markers predate the provenance column is filtered at ClusterScoreThresholdDefault
		// throughout, and stating the detector's own number there is the wrong one to act on.
		if detector := c.FaceDetector(); detector == face.DetectorNone {
			notes = append(notes, fmt.Sprintf("Each marker is filtered by the cluster score of the detector that scored it, "+
				"or %d where none is recorded.", face.ClusterScoreThresholdDefault))
		} else {
			notes = append(notes, fmt.Sprintf("Markers %s scored cluster at %d; every other marker is filtered by the bar of "+
				"the detector that scored it, or %d where none is recorded.",
				detector, c.FaceClusterScoreEffective(), face.ClusterScoreThresholdDefault))
		}
	}

	// Zero reads as the loosest of the three and is the strictest, so it is spelled out rather
	// than left to a number in the table.
	switch c.FaceClusterSplitRounds() {
	case face.ClusterSplitOff:
		notes = append(notes, "The cluster width guard is off, so a group holding several people is kept whole.")
	case 0:
		notes = append(notes, "A group wider than its own accept distance is discarded rather than split.")
	}

	notes = append(notes, fmt.Sprintf("A cluster's radius is the %dth percentile of the distances to its members.", c.FaceClusterPercentile()))

	return strings.Join(notes, " ")
}

// faceDistReport formats a face distance threshold, or reports nothing when no embedding model
// is in force. The calibrated distances are per model and are not comparable between them, so a
// report with no model would otherwise print numbers that will not apply to the model the
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

// faceDetectorReport names the detector in force, marks it when it is the one a build runs unless
// something selects otherwise, and says so when it is not the one that was asked for.
func (c *Config) faceDetectorReport() string {
	inForce := c.FaceDetector()
	resolved := face.DetectorDisplayName(inForce)

	if setting := c.FaceDetectorSetting(); setting != face.DetectorAuto && setting != inForce {
		return faceReportValue(resolved, fmt.Sprintf("%s is not available", clean.Log(setting)))
	}

	if inForce != face.DetectorNone && inForce == face.DefaultDetectorName() {
		return faceReportValue(resolved, "default")
	}

	return resolved
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
	case setting == face.ModelAuto:
		return faceReportValue(resolved, "no embedding model is installed")
	default:
		return faceReportValue(resolved, fmt.Sprintf("%s is not available", clean.Log(setting)))
	}

	var notes []string

	// Whether it is the shipped default, not whether it was derived: an operator reads the word as
	// a property of the model, and a detected name is written back, so the two would soon disagree
	// about a model nobody chose either way.
	if inForce == face.DefaultModelName() {
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

// FaceReportSection is one titled table of the `photoprism faces status` report, with the note
// that states what its values cannot.
type FaceReportSection struct {
	Title string
	Cols  []string
	Rows  [][]string
	Note  string
}

// FaceReportSections returns the face configuration grouped for `photoprism faces status`. It
// covers the same options as Report() in the same order, so the two can be read against each
// other, and its notes add what only a database connection reveals.
func (c *Config) FaceReportSections() []FaceReportSection {
	cols := []string{"Name", "Value"}
	notes := map[faceConfigSection]string{
		faceSectionDetection:   c.faceDetectionNote(),
		faceSectionRecognition: c.faceRecognitionNote(),
	}

	var sections []FaceReportSection

	for _, row := range c.faceConfigRows() {
		if n := len(sections); n == 0 || sections[n-1].Title != string(row.Section) {
			sections = append(sections, FaceReportSection{
				Title: string(row.Section),
				Cols:  cols,
				Note:  notes[row.Section],
			})
		}

		last := &sections[len(sections)-1]
		last.Rows = append(last.Rows, []string{row.Flag, row.Value})
	}

	return sections
}

// FaceStatus returns the lines `photoprism faces status` prints above the tables, stating whether
// faces are being processed and what is stopping them when they are not. A value cannot say that:
// every threshold reads the same while a lock or a model mismatch keeps the workers idle.
func (c *Config) FaceStatus() []string {
	if c == nil {
		return []string{"Face detection and recognition are unavailable."}
	}

	if c.DisableFaces() {
		return []string{"Face detection and recognition are disabled."}
	}

	detector := c.FaceDetector()
	model := c.EffectiveFaceModel()

	var lines []string

	switch {
	case detector == face.DetectorNone && model == face.ModelNone:
		lines = append(lines, "Face detection and recognition are disabled, because neither a detector nor an embedding model is in force.")
	case detector == face.DetectorNone:
		lines = append(lines, "Face detection is disabled, so no new faces are found.")
	case model == face.ModelNone:
		lines = append(lines, "Face embeddings are disabled, so the faces that are found are not recognized.")
	case c.FaceEngineRunType() == vision.RunNever:
		lines = append(lines, "Face detection and recognition are configured, but never scheduled to run.")
	default:
		lines = append(lines, "Face detection and recognition are enabled.")
	}

	if status := c.faceEmbedderStatus(); status != "" {
		lines = append(lines, status)
	}

	if status := c.faceClusterStatus(); status != "" {
		lines = append(lines, status)
	}

	// Only reported while one is held. A lock nothing states is the difference between an instance
	// that explains why it indexes nothing and one that has to be diagnosed by finding a file.
	if held := c.FacesLocked(); held != "" {
		lines = append(lines, fmt.Sprintf("A face embedding migration is in progress (%s), so the workers leave markers and clusters alone until it finishes.", held))
	}

	return lines
}

// faceClusterStatus names the bar holding automatic clustering back, or "" when nothing is.
//
// A library that never forms a cluster is indistinguishable from one that clustered everything:
// both simply show no new people. Answering it took hand-written SQL twice, though the thresholds
// that exclude a marker are the ones this instance already knows.
func (c *Config) faceClusterStatus() string {
	if c == nil || c.db == nil || c.EffectiveFaceModel() == face.ModelNone {
		return ""
	}

	size, floor := c.FaceClusterSize(), c.faceClusterScoreFloor()
	gates := query.CountFaceClusterGates(c.EffectiveFaceModel(), size, floor)

	// The same getter Propagate assigns to face.SampleThreshold, not the global: this command runs
	// on InitCore, which never propagates, so the global would still hold the shipped default and
	// the report would name a shortfall that is not the one holding.
	return faceClusterStatusFor(gates, c.FaceSampleThreshold(), size, c.faceClusterScorePhrase(floor), c.FaceClusterCore(), c.FaceClusterDist())
}

// faceClusterScorePhrase names the score bar the gate counts were taken at. Unset it is per marker,
// so a single number would state the bar of the detector in force for markers a different one
// scored - which on a library predating the provenance column is the wrong number entirely.
func (c *Config) faceClusterScorePhrase(floor int) string {
	switch floor {
	case face.ClusterScoreAuto:
		return fmt.Sprintf("the face-cluster-score of the detector that scored each one (%d where none is recorded)",
			face.ClusterScoreThresholdDefault)
	case 0:
		return "the face-cluster-score, which is switched off"
	default:
		return fmt.Sprintf("the face-cluster-score of %d", floor)
	}
}

// faceClusterStatusFor renders the clustering status for a set of gate counts, or "" when nothing
// is holding. Separate from the queries so every branch is reachable without a library shaped to
// produce it.
func faceClusterStatusFor(gates query.FaceClusterGates, required, size int, scorePhrase string, core int, dist float64) string {
	// Enough to run and nothing formed: no cluster advances the recency cut, so the pass repeats on
	// every wake. No threshold explains it, so name what decides whether a group forms.
	if gates.Eligible >= required && !gates.Clustered && gates.Unclustered > 0 {
		return fmt.Sprintf("Automatic clustering has %d eligible markers and has formed no clusters: "+
			"face-cluster-core %d requires that many faces of one person within a face-cluster-dist of %g, "+
			"counting the face itself.", gates.Eligible, core, dist)
	}

	// Enough to run: not a state an operator has to act on, and a line every healthy instance
	// prints is one nobody reads on the instance that is not. Eligible rather than Recent, because
	// the worker counts markers that already clear both bars.
	if gates.Eligible >= required {
		return ""
	}

	// The one shortfall no threshold explains: a marker older than the newest cluster never counts
	// toward the trigger again, so an instance can sit on hundreds of them and look idle. Naming
	// how many a forced run would actually take is what makes the remedy worth weighing.
	if gates.Recent == 0 && gates.Unclustered > 0 {
		return fmt.Sprintf("Automatic clustering will not start on its own: %d markers are unclustered, but none was "+
			"added since the last cluster, and %d of them clear both thresholds. "+
			`Run "photoprism faces update --force" to cluster them.`, gates.Unclustered, gates.Clusterable)
	}

	if gates.Unclustered == 0 {
		return ""
	}

	// No bar excludes anything, so the shortfall is volume alone. Naming thresholds here would
	// send an operator to tune bars that are already passing every marker they see.
	if gates.Eligible == gates.Recent {
		return fmt.Sprintf("Automatic clustering needs %d new markers (2 x face-cluster-core %d) and has %d, "+
			"which no threshold is excluding.", required, core, gates.Eligible)
	}

	// The derivation is spelled out because the two numbers otherwise look contradictory, and the
	// eligible count is named as the intersection: the size and score counts overlap, so reporting
	// them alone reads as two independent facts rather than as the arithmetic that produced it.
	return fmt.Sprintf("Automatic clustering needs %d new markers (2 x face-cluster-core %d) and has %d clearing both: "+
		"of the %d added since the last cluster, %d clear the face-cluster-size of %d px and %d clear %s.",
		required, core, gates.Eligible, gates.Recent, gates.SizeOK, size, gates.ScoreOK, scorePhrase)
}

// faceClusterScoreFloor maps FACE_CLUSTER_SCORE onto the convention the marker queries use, where
// zero removes the filter and negative asks for the bar of the detector that scored each marker.
// The two are the inverse of the option's, where zero is what defers to the detector.
func (c *Config) faceClusterScoreFloor() int {
	switch score := c.FaceClusterScore(); {
	case score > 0:
		return score
	case score < 0:
		return 0
	default:
		return face.ClusterScoreAuto
	}
}

// faceEmbedderStatus states why a model that is otherwise in force is generating no embeddings,
// as a sentence for the status report, or "" when it is generating them.
func (c *Config) faceEmbedderStatus() string {
	if err := face.EmbedderError(); err != nil {
		return fmt.Sprintf("The face embedding model failed to load: %s.", err)
	} else if reason := face.EmbeddingsBlockedReason(); reason != "" {
		return fmt.Sprintf("Face embeddings are paused, because %s.", reason)
	}

	return ""
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
