package config

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/ai/vision"
)

// Report returns global config values as a table for reporting.
func (c *Config) Report() (rows [][]string, cols []string) {
	cols = []string{"Name", "Value"}

	reportDatabaseDSN := c.ReportDatabaseDSN()
	faceEngine := c.FaceEngine()

	rows = [][]string{
		// Authentication.
		{"auth-mode", fmt.Sprintf("%s", c.AuthMode())},
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

		// Logging.
		{"log-level", c.LogLevel().String()},
		{"debug", fmt.Sprintf("%t", c.Debug())},
		{"trace", fmt.Sprintf("%t", c.Trace())},

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
		{"users-path", c.UsersPath()},
		{"users-originals-path", c.UsersOriginalsPath()},

		// Storage.
		{"storage-path", c.StoragePath()},
		{"users-storage-path", c.UsersStoragePath()},
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
		{"index-workers", fmt.Sprintf("%d", c.IndexWorkers())},
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
		{"disable-places", fmt.Sprintf("%t", c.DisablePlaces())},
		{"disable-tensorflow", fmt.Sprintf("%t", c.DisableTensorFlow())},
		{"disable-faces", fmt.Sprintf("%t", c.DisableFaces())},
		{"disable-classification", fmt.Sprintf("%t", c.DisableClassification())},
		{"disable-ffmpeg", fmt.Sprintf("%t", c.DisableFFmpeg())},
		{"disable-exiftool", fmt.Sprintf("%t", c.DisableExifTool())},
		{"disable-vips", fmt.Sprintf("%t", c.DisableVips())},
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
		{"api-uri", c.ApiUri()},
		{"static-uri", c.StaticUri()},
		{"content-uri", c.ContentUri()},
		{"video-uri", c.VideoUri()},

		// Cluster Configuration.
		{"cluster-domain", c.ClusterDomain()},
		{"cluster-cidr", c.ClusterCIDR()},
		{"cluster-uuid", c.ClusterUUID()},
		{"portal-url", c.PortalUrl()},
		{"portal-config-path", c.PortalConfigPath()},
		{"portal-theme-path", c.PortalThemePath()},
		{"join-token", fmt.Sprintf("%s", strings.Repeat("*", utf8.RuneCountInString(c.JoinToken())))},
		{"node-name", c.NodeName()},
		{"node-role", c.NodeRole()},
		{"node-uuid", c.NodeUUID()},
		{"node-client-id", c.NodeClientID()},
		{"node-client-secret", fmt.Sprintf("%s", strings.Repeat("*", utf8.RuneCountInString(c.NodeClientSecret())))},
		{"jwks-url", c.JWKSUrl()},
		{"jwks-cache-ttl", fmt.Sprintf("%d", c.JWKSCacheTTL())},
		{"jwt-scope", c.JWTAllowedScopes().String()},
		{"jwt-leeway", fmt.Sprintf("%d", c.JWTLeeway())},
		{"advertise-url", c.AdvertiseUrl()},
	}...)

	if c.Portal() {
		rows = append(rows, [][]string{
			{"database-provision-driver", c.options.DatabaseProvisionDriver},
			{"database-provision-prefix", c.DatabaseProvisionPrefix()},
			{"database-provision-dsn", maskDatabaseProvisionDSN(c.options.DatabaseProvisionDSN)},
		}...)
	}

	rows = append(rows, [][]string{
		// Proxy Servers.
		{"https-proxy", c.HttpsProxy()},
		{"https-proxy-insecure", fmt.Sprintf("%t", c.HttpsProxyInsecure())},
		{"trusted-platform", c.TrustedPlatform()},
		{"trusted-proxy", c.TrustedProxy()},
		{"proxy-client-header", c.ProxyClientHeader()},
		{"proxy-proto-header", strings.Join(c.ProxyProtoHeader(), ", ")},
		{"proxy-proto-https", strings.Join(c.ProxyProtoHttps(), ", ")},

		// Web Server.
		{"disable-tls", fmt.Sprintf("%t", c.DisableTLS())},
		{"default-tls", fmt.Sprintf("%t", c.DefaultTLS())},
		{"tls-email", c.TLSEmail()},
		{"tls-cert", c.TLSCert()},
		{"tls-key", c.TLSKey()},
		{"http-mode", c.HttpMode()},
		{"http-compression", c.HttpCompression()},
		{"http-cache-public", fmt.Sprintf("%t", c.HttpCachePublic())},
		{"http-cache-maxage", fmt.Sprintf("%d", c.HttpCacheMaxAge())},
		{"http-video-maxage", fmt.Sprintf("%d", c.HttpVideoMaxAge())},
		{"http-host", c.HttpHost()},
		{"http-port", fmt.Sprintf("%d", c.HttpPort())},
	}...)

	// Database.
	if reportDatabaseDSN {
		rows = append(rows, [][]string{
			{"database-driver", c.DatabaseDriver()},
			{"database-dsn", c.DatabaseDSN()},
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
		{"ffmpeg-preset", c.FFmpegPreset()},
		{"ffmpeg-device", c.FFmpegDevice()},
		{"ffmpeg-map-video", c.FFmpegMapVideo()},
		{"ffmpeg-map-audio", c.FFmpegMapAudio()},
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
		{"download-token", c.DownloadToken()},
		{"preview-token", c.PreviewToken()},
		{"thumb-library", c.ThumbLibrary()},
		{"thumb-color", c.ThumbColor()},
		{"thumb-filter", c.ThumbFilter().String()},
		{"thumb-size", fmt.Sprintf("%d", c.ThumbSizePrecached())},
		{"thumb-size-uncached", fmt.Sprintf("%d", c.ThumbSizeUncached())},
		{"thumb-uncached", fmt.Sprintf("%t", c.ThumbUncached())},
		{"jpeg-quality", fmt.Sprintf("%d", c.JpegQuality())},
		{"jpeg-size", fmt.Sprintf("%d", c.JpegSize())},
		{"png-size", fmt.Sprintf("%d", c.PngSize())},

		// Computer Vision & Facial Recognition.
		{"vision-yaml", c.VisionYaml()},
		{"vision-api", fmt.Sprintf("%t", c.VisionApi())},
		{"vision-uri", c.VisionUri()},
		{"vision-key", strings.Repeat("*", utf8.RuneCountInString(c.VisionKey()))},
		{"vision-schedule", c.VisionSchedule()},
		{"vision-filter", c.VisionFilter()},
		{"nasnet-model-path", c.NasnetModelPath()},
		{"facenet-model-path", c.FacenetModelPath()},
		{"nsfw-model-path", c.NsfwModelPath()},
		{"detect-nsfw", fmt.Sprintf("%t", c.DetectNSFW())},
		{"face-engine", faceEngine},
		{"face-engine-run", vision.ReportRunType(c.FaceEngineRunType())},
	}...)

	if faceEngine == face.EngineONNX {
		rows = append(rows, [][]string{
			{"face-engine-threads", fmt.Sprintf("%d", c.FaceEngineThreads())},
			{"face-size", fmt.Sprintf("%d", c.FaceSize())},
			{"face-score", fmt.Sprintf("%f", c.FaceScore())},
		}...)
	} else if faceEngine == face.EnginePigo {
		rows = append(rows, [][]string{
			{"face-size", fmt.Sprintf("%d", c.FaceSize())},
			{"face-score", fmt.Sprintf("%f", c.FaceScore())},
			{"face-angle", fmt.Sprintf("%v", c.FaceAngles())},
		}...)
	} else {
		rows = append(rows, [][]string{
			{"face-engine-threads", fmt.Sprintf("%d", c.FaceEngineThreads())},
			{"face-size", fmt.Sprintf("%d", c.FaceSize())},
			{"face-score", fmt.Sprintf("%f", c.FaceScore())},
			{"face-angle", fmt.Sprintf("%v", c.FaceAngles())},
		}...)
	}

	rows = append(rows, [][]string{
		{"face-overlap", fmt.Sprintf("%d", c.FaceOverlap())},
		{"face-cluster-size", fmt.Sprintf("%d", c.FaceClusterSize())},
		{"face-cluster-score", fmt.Sprintf("%d", c.FaceClusterScore())},
		{"face-cluster-core", fmt.Sprintf("%d", c.FaceClusterCore())},
		{"face-cluster-dist", fmt.Sprintf("%f", c.FaceClusterDist())},
		{"face-cluster-radius", fmt.Sprintf("%f", c.FaceClusterRadius())},
		{"face-collision-dist", fmt.Sprintf("%f", c.FaceCollisionDist())},
		{"face-epsilon-dist", fmt.Sprintf("%f", c.FaceEpsilonDist())},
		{"face-match-dist", fmt.Sprintf("%f", c.FaceMatchDist())},
		{"face-skip-children", fmt.Sprintf("%t", c.FaceSkipChildren())},
		{"face-allow-background", fmt.Sprintf("%t", c.FaceAllowBackground())},

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

func maskDatabaseProvisionDSN(dsn string) string {
	if dsn == "" {
		return ""
	}

	ds := NewDSN(dsn)
	if ds.Password == "" {
		return dsn
	}

	needle := ":" + ds.Password + "@"
	if strings.Contains(dsn, needle) {
		return strings.Replace(dsn, needle, ":***@", 1)
	}

	return dsn
}
