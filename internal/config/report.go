package config

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

// Report returns global config values as a table for reporting.
func (c *Config) Report() (rows [][]string, cols []string) {
	cols = []string{"Name", "Value"}

	var dbKey string

	if c.DatabaseDriver() == SQLite3 {
		dbKey = "database-dsn"
	} else {
		dbKey = "database-name"
	}

	rows = [][]string{
		// Authentication.
		{"auth-mode", fmt.Sprintf("%s", c.AuthMode())},
		{"admin-user", c.AdminUser()},
		{"admin-password", strings.Repeat("*", utf8.RuneCountInString(c.AdminPassword()))},
		{"public", fmt.Sprintf("%t", c.Public())},
		{"sess-maxage", fmt.Sprintf("%d", c.SessMaxAge())},
		{"sess-timeout", fmt.Sprintf("%d", c.SessTimeout())},

		// Logging.
		{"log-level", c.LogLevel().String()},
		{"debug", fmt.Sprintf("%t", c.Debug())},
		{"trace", fmt.Sprintf("%t", c.Trace())},

		// Config.
		{"config-path", c.ConfigPath()},
		{"certificates-path", c.CertificatesPath()},
		{"options-yaml", c.OptionsYaml()},
		{"defaults-yaml", c.DefaultsYaml()},
		{"settings-yaml", c.SettingsYaml()},

		// Originals.
		{"originals-path", c.OriginalsPath()},
		{"originals-limit", fmt.Sprintf("%d", c.OriginalsLimit())},
		{"resolution-limit", fmt.Sprintf("%d", c.ResolutionLimit())},

		// Other paths.
		{"storage-path", c.StoragePath()},
		{"sidecar-path", c.SidecarPath()},
		{"users-path", c.UsersPath()},
		{"albums-path", c.AlbumsPath()},
		{"backup-path", c.BackupPath()},
		{"cache-path", c.CachePath()},
		{"cmd-cache-path", c.CmdCachePath()},
		{"thumb-cache-path", c.ThumbCachePath()},
		{"import-path", c.ImportPath()},
		{"import-dest", c.ImportDest()},
		{"assets-path", c.AssetsPath()},
		{"static-path", c.StaticPath()},
		{"build-path", c.BuildPath()},
		{"img-path", c.ImgPath()},
		{"templates-path", c.TemplatesPath()},
		{"temp-path", c.TempPath()},

		// Workers.
		{"workers", fmt.Sprintf("%d", c.Workers())},
		{"wakeup-interval", c.WakeupInterval().String()},
		{"auto-index", fmt.Sprintf("%d", c.AutoIndex()/time.Second)},
		{"auto-import", fmt.Sprintf("%d", c.AutoImport()/time.Second)},

		// Feature Flags.
		{"read-only", fmt.Sprintf("%t", c.ReadOnly())},
		{"experimental", fmt.Sprintf("%t", c.Experimental())},
		{"disable-webdav", fmt.Sprintf("%t", c.DisableWebDAV())},
		{"disable-settings", fmt.Sprintf("%t", c.DisableSettings())},
		{"disable-places", fmt.Sprintf("%t", c.DisablePlaces())},
		{"disable-backups", fmt.Sprintf("%t", c.DisableBackups())},
		{"disable-tensorflow", fmt.Sprintf("%t", c.DisableTensorFlow())},
		{"disable-faces", fmt.Sprintf("%t", c.DisableFaces())},
		{"disable-classification", fmt.Sprintf("%t", c.DisableClassification())},
		{"disable-ffmpeg", fmt.Sprintf("%t", c.DisableFFmpeg())},
		{"disable-exiftool", fmt.Sprintf("%t", c.DisableExifTool())},
		{"disable-heifconvert", fmt.Sprintf("%t", c.DisableHeifConvert())},
		{"disable-darktable", fmt.Sprintf("%t", c.DisableDarktable())},
		{"disable-rawtherapee", fmt.Sprintf("%t", c.DisableRawtherapee())},
		{"disable-sips", fmt.Sprintf("%t", c.DisableSips())},
		{"disable-raw", fmt.Sprintf("%t", c.DisableRaw())},

		// Format Flags.
		{"raw-presets", fmt.Sprintf("%t", c.RawPresets())},
		{"exif-bruteforce", fmt.Sprintf("%t", c.ExifBruteForce())},

		// TensorFlow.
		{"detect-nsfw", fmt.Sprintf("%t", c.DetectNSFW())},
		{"upload-nsfw", fmt.Sprintf("%t", c.UploadNSFW())},
		{"tensorflow-version", c.TensorFlowVersion()},
		{"tensorflow-model-path", c.TensorFlowModelPath()},

		// Customization.
		{"default-locale", c.DefaultLocale()},
		{"default-theme", c.DefaultTheme()},
		{"app-icon", c.AppIcon()},
		{"app-name", c.AppName()},
		{"app-mode", c.AppMode()},
		{"wallpaper-uri", c.WallpaperUri()},

		// Site Infos.
		{"cdn-url", c.CdnUrl("/")},
		{"site-url", c.SiteUrl()},
		{"site-https", fmt.Sprintf("%t", c.SiteHttps())},
		{"site-domain", c.SiteDomain()},
		{"site-author", c.SiteAuthor()},
		{"site-title", c.SiteTitle()},
		{"site-caption", c.SiteCaption()},
		{"site-description", c.SiteDescription()},
		{"site-preview", c.SitePreview()},

		// Legal info.
		{"imprint", c.Imprint()},
		{"imprint-url", c.ImprintUrl()},

		// URIs.
		{"content-uri", c.ContentUri()},
		{"static-uri", c.StaticUri()},
		{"api-uri", c.ApiUri()},
		{"base-uri", c.BaseUri("/")},

		// HTTP(S) Proxy.
		{"trusted-proxy", c.TrustedProxy()},
		{"https-proxy-header", strings.Join(c.HttpsProxyHeader(), ", ")},
		{"https-proxy-proto", strings.Join(c.HttpsProxyProto(), ", ")},

		// Web Server.
		{"http-mode", c.HttpMode()},
		{"http-compression", c.HttpCompression()},
		{"http-host", c.HttpHost()},
		{"http-port", fmt.Sprintf("%d", c.HttpPort())},
		{"tls-email", c.TLSEmail()},
		{"tls-cert", c.TLSCert()},
		{"tls-key", c.TLSKey()},
		{"disable-tls", fmt.Sprintf("%t", c.DisableTLS())},

		// Database.
		{"database-driver", c.DatabaseDriver()},
		{dbKey, c.DatabaseName()},
		{"database-server", c.DatabaseServer()},
		{"database-host", c.DatabaseHost()},
		{"database-port", c.DatabasePortString()},
		{"database-user", c.DatabaseUser()},
		{"database-password", strings.Repeat("*", utf8.RuneCountInString(c.DatabasePassword()))},
		{"database-conns", fmt.Sprintf("%d", c.DatabaseConns())},
		{"database-conns-idle", fmt.Sprintf("%d", c.DatabaseConnsIdle())},

		// External Tools.
		{"darktable-bin", c.DarktableBin()},
		{"darktable-cache-path", c.DarktableCachePath()},
		{"darktable-config-path", c.DarktableConfigPath()},
		{"darktable-blacklist", c.DarktableBlacklist()},
		{"rawtherapee-bin", c.RawtherapeeBin()},
		{"rawtherapee-blacklist", c.RawtherapeeBlacklist()},
		{"sips-bin", c.SipsBin()},
		{"sips-blacklist", c.SipsBlacklist()},
		{"heifconvert-bin", c.HeifConvertBin()},
		{"ffmpeg-bin", c.FFmpegBin()},
		{"ffmpeg-encoder", c.FFmpegEncoder().String()},
		{"ffmpeg-bitrate", fmt.Sprintf("%d", c.FFmpegBitrate())},
		{"exiftool-bin", c.ExifToolBin()},

		// Thumbnails.
		{"download-token", c.DownloadToken()},
		{"preview-token", c.PreviewToken()},
		{"thumb-color", c.ThumbColor()},
		{"thumb-filter", string(c.ThumbFilter())},
		{"thumb-size", fmt.Sprintf("%d", c.ThumbSizePrecached())},
		{"thumb-size-uncached", fmt.Sprintf("%d", c.ThumbSizeUncached())},
		{"thumb-uncached", fmt.Sprintf("%t", c.ThumbUncached())},
		{"jpeg-quality", fmt.Sprintf("%d", c.JpegQuality())},
		{"jpeg-size", fmt.Sprintf("%d", c.JpegSize())},

		// Facial Recognition.
		{"face-size", fmt.Sprintf("%d", c.FaceSize())},
		{"face-score", fmt.Sprintf("%f", c.FaceScore())},
		{"face-overlap", fmt.Sprintf("%d", c.FaceOverlap())},
		{"face-cluster-size", fmt.Sprintf("%d", c.FaceClusterSize())},
		{"face-cluster-score", fmt.Sprintf("%d", c.FaceClusterScore())},
		{"face-cluster-core", fmt.Sprintf("%d", c.FaceClusterCore())},
		{"face-cluster-dist", fmt.Sprintf("%f", c.FaceClusterDist())},
		{"face-match-dist", fmt.Sprintf("%f", c.FaceMatchDist())},

		// Daemon Mode.
		{"pid-filename", c.PIDFilename()},
		{"log-filename", c.LogFilename()},
	}

	if v := c.CustomAssetsPath(); v != "" {
		rows = append(rows, []string{"custom-assets-path", v})
	}

	if v := c.CustomStaticUri(); v != "" {
		rows = append(rows, []string{"custom-static-uri", v})
	}

	return rows, cols
}
