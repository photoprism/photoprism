package config

import (
	"fmt"

	"github.com/klauspost/cpuid/v2"
	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/face"
	"github.com/photoprism/photoprism/internal/i18n"
	"github.com/photoprism/photoprism/internal/thumb"
)

// Flags lists all global command-line parameters.
var Flags = CliFlags{
	CliFlag{
		Flag: cli.StringFlag{
			Name:   "admin-password, pw",
			Usage:  fmt.Sprintf("initial admin `PASSWORD`, must have at least %d characters", entity.PasswordLength),
			EnvVar: "PHOTOPRISM_ADMIN_PASSWORD",
		}},
	CliFlag{
		Flag: cli.BoolFlag{
			Name:   "public",
			Usage:  "disable password authentication, incl WebDAV and Advanced Settings",
			EnvVar: "PHOTOPRISM_PUBLIC",
		}},
	CliFlag{
		Flag: cli.StringFlag{
			Name:   "log-level, l",
			Usage:  "log message verbosity `LEVEL` (trace, debug, info, warning, error, fatal, panic)",
			Value:  "info",
			EnvVar: "PHOTOPRISM_LOG_LEVEL",
		}},
	CliFlag{
		Flag: cli.BoolFlag{
			Name:   "debug",
			Usage:  "enable debug mode, show non-essential log messages",
			EnvVar: "PHOTOPRISM_DEBUG",
		}},
	CliFlag{
		Flag: cli.BoolFlag{
			Name:   "trace",
			Usage:  "enable trace mode, show all log messages",
			EnvVar: "PHOTOPRISM_TRACE",
		}},
	CliFlag{
		Flag: cli.BoolFlag{
			Name:   "test",
			Hidden: true,
			Usage:  "enable test mode",
		},
	},
	CliFlag{
		Flag: cli.BoolFlag{
			Name:   "unsafe",
			Hidden: true,
			Usage:  "disable safety checks",
			EnvVar: "PHOTOPRISM_UNSAFE",
		}},
	CliFlag{
		Flag: cli.BoolFlag{
			Name:   "demo",
			Hidden: true,
			Usage:  "enable demo mode",
			EnvVar: "PHOTOPRISM_DEMO",
		}},
	CliFlag{
		Flag: cli.BoolFlag{
			Name:   "sponsor",
			Hidden: true,
			Usage:  "your continuous support helps to pay for development and operating expenses",
			EnvVar: "PHOTOPRISM_SPONSOR",
		}},
	CliFlag{
		Flag: cli.StringFlag{
			Name:   "partner-id",
			Hidden: true,
			Usage:  "hosting partner id",
			EnvVar: "PHOTOPRISM_PARTNER_ID",
		}},
	CliFlag{
		Flag: cli.StringFlag{
			Name:   "config-path, c",
			Usage:  "config storage `PATH`, values in options.yml override CLI flags and environment variables if present",
			EnvVar: "PHOTOPRISM_CONFIG_PATH",
		}},
	CliFlag{
		Flag: cli.StringFlag{
			Name:   "defaults-yaml, y",
			Usage:  "load config defaults from `FILE` if exists, does not override CLI flags and environment variables",
			Value:  "/etc/photoprism/defaults.yml",
			EnvVar: "PHOTOPRISM_DEFAULTS_YAML",
		}},
	CliFlag{
		Flag: cli.StringFlag{
			Name:   "originals-path, o",
			Usage:  "storage `PATH` of your original media files (photos and videos)",
			EnvVar: "PHOTOPRISM_ORIGINALS_PATH",
		}},
	CliFlag{
		Flag: cli.IntFlag{
			Name:   "originals-limit, mb",
			Value:  1000,
			Usage:  "maximum size of media files in `MB` (1-100000; -1 to disable)",
			EnvVar: "PHOTOPRISM_ORIGINALS_LIMIT",
		}},
	CliFlag{
		Flag: cli.IntFlag{
			Name:   "resolution-limit, mp",
			Value:  100,
			Usage:  "maximum resolution of media files in `MEGAPIXELS` (1-900; -1 to disable)",
			EnvVar: "PHOTOPRISM_RESOLUTION_LIMIT",
		},
		Tags: []string{EnvSponsor},
	},
	CliFlag{
		Flag: cli.StringFlag{
			Name:   "storage-path, s",
			Usage:  "writable storage `PATH` for sidecar, cache, and database files",
			EnvVar: "PHOTOPRISM_STORAGE_PATH",
		}},
	CliFlag{
		Flag: cli.StringFlag{
			Name:   "sidecar-path, sc",
			Usage:  "custom relative or absolute sidecar `PATH` *optional*",
			EnvVar: "PHOTOPRISM_SIDECAR_PATH",
		}},
	CliFlag{
		Flag: cli.StringFlag{
			Name:   "backup-path, ba",
			Usage:  "custom backup `PATH` for index backup files *optional*",
			EnvVar: "PHOTOPRISM_BACKUP_PATH",
		}},
	CliFlag{
		Flag: cli.StringFlag{
			Name:   "cache-path, ca",
			Usage:  "custom cache `PATH` for sessions and thumbnail files *optional*",
			EnvVar: "PHOTOPRISM_CACHE_PATH",
		}},
	CliFlag{
		Flag: cli.StringFlag{
			Name:   "import-path, im",
			Usage:  "base `PATH` from which files can be imported to originals *optional*",
			EnvVar: "PHOTOPRISM_IMPORT_PATH",
		}},
	CliFlag{
		Flag: cli.StringFlag{
			Name:   "assets-path, as",
			Usage:  "assets `PATH` containing static resources like icons, models, and translations",
			EnvVar: "PHOTOPRISM_ASSETS_PATH",
		}},
	CliFlag{
		Flag: cli.StringFlag{
			Name:   "temp-path, tmp",
			Usage:  "temporary file `PATH` *optional*",
			EnvVar: "PHOTOPRISM_TEMP_PATH",
		}},
	CliFlag{
		Flag: cli.IntFlag{
			Name:   "workers, w",
			Usage:  "maximum `NUMBER` of indexing workers, default depends on the number of physical cores",
			EnvVar: "PHOTOPRISM_WORKERS",
			Value:  cpuid.CPU.PhysicalCores / 2,
		}},
	CliFlag{
		Flag: cli.StringFlag{
			Name:   "wakeup-interval, i",
			Usage:  "`DURATION` between worker runs required for face recognition and index maintenance (1-86400s)",
			Value:  DefaultWakeupInterval.String(),
			EnvVar: "PHOTOPRISM_WAKEUP_INTERVAL",
		}},
	CliFlag{
		Flag: cli.IntFlag{
			Name:   "auto-index",
			Usage:  "WebDAV auto index safety delay in `SECONDS` (-1 to disable)",
			Value:  DefaultAutoIndexDelay,
			EnvVar: "PHOTOPRISM_AUTO_INDEX",
		}},
	CliFlag{
		Flag: cli.IntFlag{
			Name:   "auto-import",
			Usage:  "WebDAV auto import safety delay in `SECONDS` (-1 to disable)",
			Value:  DefaultAutoImportDelay,
			EnvVar: "PHOTOPRISM_AUTO_IMPORT",
		}},
	CliFlag{
		Flag: cli.BoolFlag{
			Name:   "read-only, r",
			Usage:  "disable import, upload, delete, and all other operations that require write permissions",
			EnvVar: "PHOTOPRISM_READONLY",
		}},
	CliFlag{
		Flag: cli.BoolFlag{
			Name:   "experimental, e",
			Usage:  "enable experimental features",
			EnvVar: "PHOTOPRISM_EXPERIMENTAL",
		}},
	CliFlag{
		Flag: cli.BoolFlag{
			Name:   "disable-webdav",
			Usage:  "disable built-in WebDAV server",
			EnvVar: "PHOTOPRISM_DISABLE_WEBDAV",
		}},
	CliFlag{
		Flag: cli.BoolFlag{
			Name:   "disable-settings",
			Usage:  "disable settings UI and API",
			EnvVar: "PHOTOPRISM_DISABLE_SETTINGS",
		}},
	CliFlag{
		Flag: cli.BoolFlag{
			Name:   "disable-places",
			Usage:  "disable reverse geocoding and maps",
			EnvVar: "PHOTOPRISM_DISABLE_PLACES",
		}},
	CliFlag{
		Flag: cli.BoolFlag{
			Name:   "disable-backups",
			Usage:  "disable backing up albums and photo metadata to YAML files",
			EnvVar: "PHOTOPRISM_DISABLE_BACKUPS",
		}},
	CliFlag{
		Flag: cli.BoolFlag{
			Name:   "disable-tensorflow",
			Usage:  "disable all features depending on TensorFlow",
			EnvVar: "PHOTOPRISM_DISABLE_TENSORFLOW",
		}},
	CliFlag{
		Flag: cli.BoolFlag{
			Name:   "disable-faces",
			Usage:  "disable facial recognition",
			EnvVar: "PHOTOPRISM_DISABLE_FACES",
		}},
	CliFlag{
		Flag: cli.BoolFlag{
			Name:   "disable-classification",
			Usage:  "disable image classification",
			EnvVar: "PHOTOPRISM_DISABLE_CLASSIFICATION",
		}},
	CliFlag{
		Flag: cli.BoolFlag{
			Name:   "disable-ffmpeg",
			Usage:  "disable video transcoding and thumbnail extraction with FFmpeg",
			EnvVar: "PHOTOPRISM_DISABLE_FFMPEG",
		}},
	CliFlag{
		Flag: cli.BoolFlag{
			Name:   "disable-exiftool",
			Usage:  "disable creating JSON metadata sidecar files with ExifTool",
			EnvVar: "PHOTOPRISM_DISABLE_EXIFTOOL",
		}},
	CliFlag{
		Flag: cli.BoolFlag{
			Name:   "disable-heifconvert",
			Usage:  "disable conversion of HEIC/HEIF files",
			EnvVar: "PHOTOPRISM_DISABLE_HEIFCONVERT",
		}},
	CliFlag{
		Flag: cli.BoolFlag{
			Name:   "disable-darktable",
			Usage:  "disable conversion of RAW files with Darktable",
			EnvVar: "PHOTOPRISM_DISABLE_DARKTABLE",
		}},
	CliFlag{
		Flag: cli.BoolFlag{
			Name:   "disable-rawtherapee",
			Usage:  "disable conversion of RAW files with RawTherapee",
			EnvVar: "PHOTOPRISM_DISABLE_RAWTHERAPEE",
		}},
	CliFlag{
		Flag: cli.BoolFlag{
			Name:   "disable-sips",
			Usage:  "disable conversion of RAW files with Sips *macOS only*",
			EnvVar: "PHOTOPRISM_DISABLE_SIPS",
		}},
	CliFlag{
		Flag: cli.BoolFlag{
			Name:   "disable-raw",
			Usage:  "disable indexing and conversion of RAW files",
			EnvVar: "PHOTOPRISM_DISABLE_RAW",
		}},
	CliFlag{
		Flag: cli.BoolFlag{
			Name:   "raw-presets",
			Usage:  "enables applying user presets when converting RAW files (reduces performance)",
			EnvVar: "PHOTOPRISM_RAW_PRESETS",
		}},
	CliFlag{
		Flag: cli.BoolFlag{
			Name:   "exif-bruteforce",
			Usage:  "always perform a brute-force search if no Exif headers were found",
			EnvVar: "PHOTOPRISM_EXIF_BRUTEFORCE",
		}},
	CliFlag{
		Flag: cli.BoolFlag{
			Name:   "detect-nsfw",
			Usage:  "flag photos as private that may be offensive (requires TensorFlow)",
			EnvVar: "PHOTOPRISM_DETECT_NSFW",
		}},
	CliFlag{
		Flag: cli.BoolFlag{
			Name:   "upload-nsfw, n",
			Usage:  "allow uploads that may be offensive",
			EnvVar: "PHOTOPRISM_UPLOAD_NSFW",
		}},
	CliFlag{
		Flag: cli.StringFlag{
			Name:   "default-locale, lang",
			Usage:  "standard user interface language `CODE`",
			Value:  i18n.Default.Locale(),
			EnvVar: "PHOTOPRISM_DEFAULT_LOCALE",
		},
	},
	CliFlag{
		Flag: cli.StringFlag{
			Name:   "default-theme",
			Usage:  "standard user interface theme `NAME`",
			EnvVar: "PHOTOPRISM_DEFAULT_THEME",
		},
		Tags: []string{EnvSponsor},
	},
	CliFlag{
		Flag: cli.StringFlag{
			Name:   "app-mode",
			Usage:  "progressive web app `MODE` (fullscreen, standalone, minimal-ui, browser)",
			Value:  "standalone",
			EnvVar: "PHOTOPRISM_APP_MODE",
		}},
	CliFlag{
		Flag: cli.StringFlag{
			Name:   "app-icon",
			Usage:  "progressive web app `ICON` (logo, app, crisp, mint, bold)",
			EnvVar: "PHOTOPRISM_APP_ICON",
		},
		Tags: []string{EnvSponsor},
	},
	CliFlag{
		Flag: cli.StringFlag{
			Name:   "app-name",
			Usage:  "progressive web app `NAME` when installed on a device",
			Value:  "",
			EnvVar: "PHOTOPRISM_APP_NAME",
		},
		Tags: []string{EnvSponsor},
	},
	CliFlag{
		Flag: cli.StringFlag{
			Name:   "imprint",
			Usage:  "legal `INFORMATION`, displayed in the page footer",
			Value:  "",
			EnvVar: "PHOTOPRISM_IMPRINT",
		},
		Tags: []string{EnvSponsor},
	},
	CliFlag{
		Flag: cli.StringFlag{
			Name:   "imprint-url",
			Usage:  "legal information `URL`",
			Value:  "",
			EnvVar: "PHOTOPRISM_IMPRINT_URL",
		},
		Tags: []string{EnvSponsor},
	},
	CliFlag{
		Flag: cli.StringFlag{
			Name:   "wallpaper-uri",
			Usage:  "login screen background image `URI`",
			EnvVar: "PHOTOPRISM_WALLPAPER_URI",
			Value:  "",
		},
		Tags: []string{EnvSponsor},
	},
	CliFlag{
		Flag: cli.StringFlag{
			Name:   "cdn-url",
			Usage:  "content delivery network `URL`",
			EnvVar: "PHOTOPRISM_CDN_URL",
		},
		Tags: []string{EnvSponsor},
	},
	CliFlag{
		Flag: cli.StringFlag{
			Name:   "site-url, url",
			Usage:  "public site `URL`",
			Value:  "http://localhost:2342/",
			EnvVar: "PHOTOPRISM_SITE_URL",
		}},
	CliFlag{
		Flag: cli.StringFlag{
			Name:   "site-author",
			Usage:  "site `OWNER`, copyright, or artist",
			EnvVar: "PHOTOPRISM_SITE_AUTHOR",
		}},
	CliFlag{
		Flag: cli.StringFlag{
			Name:   "site-title",
			Usage:  "site `TITLE`",
			Value:  "",
			EnvVar: "PHOTOPRISM_SITE_TITLE",
		},
		Tags: []string{EnvSponsor},
	},
	CliFlag{
		Flag: cli.StringFlag{
			Name:   "site-caption",
			Usage:  "site `CAPTION`",
			Value:  "AI-Powered Photos App",
			EnvVar: "PHOTOPRISM_SITE_CAPTION",
		},
	},
	CliFlag{
		Flag: cli.StringFlag{
			Name:   "site-description",
			Usage:  "site `DESCRIPTION` *optional*",
			EnvVar: "PHOTOPRISM_SITE_DESCRIPTION",
		}},
	CliFlag{
		Flag: cli.StringFlag{
			Name:   "site-preview",
			Usage:  "sharing preview image `URL`",
			EnvVar: "PHOTOPRISM_SITE_PREVIEW",
		},
		Tags: []string{EnvSponsor},
	},
	CliFlag{
		Flag: cli.IntFlag{
			Name:   "http-port, port",
			Value:  2342,
			Usage:  "http server port `NUMBER`",
			EnvVar: "PHOTOPRISM_HTTP_PORT",
		}},
	CliFlag{
		Flag: cli.StringFlag{
			Name:   "http-host, ip",
			Usage:  "http server `IP` address",
			EnvVar: "PHOTOPRISM_HTTP_HOST",
		}},
	CliFlag{
		Flag: cli.StringFlag{
			Name:   "http-mode, mode",
			Usage:  "http server `MODE` (debug, release, or test)",
			EnvVar: "PHOTOPRISM_HTTP_MODE",
		}},
	CliFlag{
		Flag: cli.StringFlag{
			Name:   "http-compression, z",
			Usage:  "http server compression `METHOD` (none or gzip)",
			EnvVar: "PHOTOPRISM_HTTP_COMPRESSION",
		}},
	CliFlag{
		Flag: cli.StringFlag{
			Name:   "database-driver, db",
			Usage:  "database `DRIVER` (sqlite, mysql)",
			Value:  "sqlite",
			EnvVar: "PHOTOPRISM_DATABASE_DRIVER",
		}},
	CliFlag{
		Flag: cli.StringFlag{
			Name:   "database-dsn, dsn",
			Usage:  "database connection `DSN` (sqlite file, optional for mysql)",
			EnvVar: "PHOTOPRISM_DATABASE_DSN",
		}},
	CliFlag{
		Flag: cli.StringFlag{
			Name:   "database-server, db-server",
			Usage:  "database `HOST` incl. port e.g. \"mariadb:3306\" (or socket path)",
			EnvVar: "PHOTOPRISM_DATABASE_SERVER",
		}},
	CliFlag{
		Flag: cli.StringFlag{
			Name:   "database-name, db-name",
			Value:  "photoprism",
			Usage:  "database schema `NAME`",
			EnvVar: "PHOTOPRISM_DATABASE_NAME",
		}},
	CliFlag{
		Flag: cli.StringFlag{
			Name:   "database-user, db-user",
			Value:  "photoprism",
			Usage:  "database user `NAME`",
			EnvVar: "PHOTOPRISM_DATABASE_USER",
		}},
	CliFlag{
		Flag: cli.StringFlag{
			Name:   "database-password, db-pass",
			Usage:  "database user `PASSWORD`",
			EnvVar: "PHOTOPRISM_DATABASE_PASSWORD",
		}},
	CliFlag{
		Flag: cli.IntFlag{
			Name:   "database-conns",
			Usage:  "maximum `NUMBER` of open database connections",
			EnvVar: "PHOTOPRISM_DATABASE_CONNS",
		}},
	CliFlag{
		Flag: cli.IntFlag{
			Name:   "database-conns-idle",
			Usage:  "maximum `NUMBER` of idle database connections",
			EnvVar: "PHOTOPRISM_DATABASE_CONNS_IDLE",
		}},
	CliFlag{
		Flag: cli.StringFlag{
			Name:   "darktable-bin",
			Usage:  "Darktable CLI `COMMAND` for RAW to JPEG conversion",
			Value:  "darktable-cli",
			EnvVar: "PHOTOPRISM_DARKTABLE_BIN",
		}},
	CliFlag{
		Flag: cli.StringFlag{
			Name:   "darktable-blacklist",
			Usage:  "do not use Darktable to convert files with these `EXTENSIONS`",
			Value:  "dng,cr3",
			EnvVar: "PHOTOPRISM_DARKTABLE_BLACKLIST",
		}},
	CliFlag{
		Flag: cli.StringFlag{
			Name:   "darktable-cache-path",
			Usage:  "custom Darktable cache `PATH`",
			Value:  "",
			EnvVar: "PHOTOPRISM_DARKTABLE_CACHE_PATH",
		}},
	CliFlag{
		Flag: cli.StringFlag{
			Name:   "darktable-config-path",
			Usage:  "custom Darktable config `PATH`",
			Value:  "",
			EnvVar: "PHOTOPRISM_DARKTABLE_CONFIG_PATH",
		}},
	CliFlag{
		Flag: cli.StringFlag{
			Name:   "rawtherapee-bin",
			Usage:  "RawTherapee CLI `COMMAND` for RAW to JPEG conversion",
			Value:  "rawtherapee-cli",
			EnvVar: "PHOTOPRISM_RAWTHERAPEE_BIN",
		}},
	CliFlag{
		Flag: cli.StringFlag{
			Name:   "rawtherapee-blacklist",
			Usage:  "do not use RawTherapee to convert files with these `EXTENSIONS`",
			Value:  "",
			EnvVar: "PHOTOPRISM_RAWTHERAPEE_BLACKLIST",
		}},
	CliFlag{
		Flag: cli.StringFlag{
			Name:   "sips-bin",
			Usage:  "Sips `COMMAND` for RAW to JPEG conversion *macOS only*",
			Value:  "sips",
			EnvVar: "PHOTOPRISM_SIPS_BIN",
		}},
	CliFlag{
		Flag: cli.StringFlag{
			Name:   "heifconvert-bin",
			Usage:  "HEIC/HEIF image conversion `COMMAND`",
			Value:  "heif-convert",
			EnvVar: "PHOTOPRISM_HEIFCONVERT_BIN",
		}},
	CliFlag{
		Flag: cli.StringFlag{
			Name:   "ffmpeg-bin",
			Usage:  "FFmpeg `COMMAND` for video transcoding and thumbnail extraction",
			Value:  "ffmpeg",
			EnvVar: "PHOTOPRISM_FFMPEG_BIN",
		}},
	CliFlag{
		Flag: cli.StringFlag{
			Name:   "ffmpeg-encoder, vc",
			Usage:  "FFmpeg AVC encoder `NAME`",
			Value:  "libx264",
			EnvVar: "PHOTOPRISM_FFMPEG_ENCODER",
		},
		Tags: []string{EnvSponsor},
	},
	CliFlag{
		Flag: cli.IntFlag{
			Name:   "ffmpeg-bitrate, vb",
			Usage:  "maximum FFmpeg encoding `BITRATE` (Mbit/s)",
			Value:  50,
			EnvVar: "PHOTOPRISM_FFMPEG_BITRATE",
		}},
	CliFlag{
		Flag: cli.StringFlag{
			Name:   "exiftool-bin",
			Usage:  "ExifTool `COMMAND` for extracting metadata",
			Value:  "exiftool",
			EnvVar: "PHOTOPRISM_EXIFTOOL_BIN",
		}},
	CliFlag{
		Flag: cli.StringFlag{
			Name:   "download-token",
			Usage:  "`SECRET` download URL token for originals (default: random)",
			EnvVar: "PHOTOPRISM_DOWNLOAD_TOKEN",
		}},
	CliFlag{
		Flag: cli.StringFlag{
			Name:   "preview-token",
			Usage:  "`SECRET` thumbnail and video streaming URL token (default: random)",
			EnvVar: "PHOTOPRISM_PREVIEW_TOKEN",
		}},
	CliFlag{
		Flag: cli.StringFlag{
			Name:   "thumb-color",
			Usage:  "standard color `PROFILE` for thumbnails (leave blank to disable)",
			Value:  "sRGB",
			EnvVar: "PHOTOPRISM_THUMB_COLOR",
		}},
	CliFlag{
		Flag: cli.StringFlag{
			Name:   "thumb-filter, filter",
			Usage:  "image downscaling filter `NAME` (best to worst: blackman, lanczos, cubic, linear)",
			Value:  "lanczos",
			EnvVar: "PHOTOPRISM_THUMB_FILTER",
		}},
	CliFlag{
		Flag: cli.IntFlag{
			Name:   "thumb-size",
			Usage:  "maximum size of thumbnails created during indexing in `PIXELS` (720-7680)",
			Value:  2048,
			EnvVar: "PHOTOPRISM_THUMB_SIZE",
		}},
	CliFlag{
		Flag: cli.IntFlag{
			Name:   "thumb-size-uncached",
			Usage:  "maximum size of missing thumbnails created on demand in `PIXELS` (720-7680)",
			Value:  7680,
			EnvVar: "PHOTOPRISM_THUMB_SIZE_UNCACHED",
		}},
	CliFlag{
		Flag: cli.BoolFlag{
			Name:   "thumb-uncached, u",
			Usage:  "enable on-demand creation of missing thumbnails (high memory and cpu usage)",
			EnvVar: "PHOTOPRISM_THUMB_UNCACHED",
		}},
	CliFlag{
		Flag: cli.StringFlag{
			Name:   "jpeg-quality, q",
			Usage:  "`QUALITY` of created JPEG sidecars and thumbnails (25-100, best, high, default, low, worst)",
			Value:  thumb.JpegQuality.String(),
			EnvVar: "PHOTOPRISM_JPEG_QUALITY",
		}},
	CliFlag{
		Flag: cli.IntFlag{
			Name:   "jpeg-size",
			Usage:  "maximum size of created JPEG sidecar files in `PIXELS` (720-30000)",
			Value:  7680,
			EnvVar: "PHOTOPRISM_JPEG_SIZE",
		}},
	CliFlag{
		Flag: cli.IntFlag{
			Name:   "face-size",
			Usage:  "minimum size of faces in `PIXELS` (20-10000)",
			Value:  face.SizeThreshold,
			EnvVar: "PHOTOPRISM_FACE_SIZE",
		}},
	CliFlag{
		Flag: cli.Float64Flag{
			Name:   "face-score",
			Usage:  "minimum face `QUALITY` score (1-100)",
			Value:  face.ScoreThreshold,
			EnvVar: "PHOTOPRISM_FACE_SCORE",
		}},
	CliFlag{
		Flag: cli.IntFlag{
			Name:   "face-overlap",
			Usage:  "face area overlap threshold in `PERCENT` (1-100)",
			Value:  face.OverlapThreshold,
			EnvVar: "PHOTOPRISM_FACE_OVERLAP",
		},
	},
	CliFlag{
		Flag: cli.IntFlag{
			Name:   "face-cluster-size",
			Usage:  "minimum size of automatically clustered faces in `PIXELS` (20-10000)",
			Value:  face.ClusterSizeThreshold,
			EnvVar: "PHOTOPRISM_FACE_CLUSTER_SIZE",
		},
		Tags: []string{EnvSponsor},
	},
	CliFlag{
		Flag: cli.IntFlag{
			Name:   "face-cluster-score",
			Usage:  "minimum `QUALITY` score of automatically clustered faces (1-100)",
			Value:  face.ClusterScoreThreshold,
			EnvVar: "PHOTOPRISM_FACE_CLUSTER_SCORE",
		},
		Tags: []string{EnvSponsor},
	},
	CliFlag{
		Flag: cli.IntFlag{
			Name:   "face-cluster-core",
			Usage:  "`NUMBER` of faces forming a cluster core (1-100)",
			Value:  face.ClusterCore,
			EnvVar: "PHOTOPRISM_FACE_CLUSTER_CORE",
		},
		Tags: []string{EnvSponsor},
	},
	CliFlag{
		Flag: cli.Float64Flag{
			Name:   "face-cluster-dist",
			Usage:  "similarity `DISTANCE` of faces forming a cluster core (0.1-1.5)",
			Value:  face.ClusterDist,
			EnvVar: "PHOTOPRISM_FACE_CLUSTER_DIST",
		},
		Tags: []string{EnvSponsor},
	},
	CliFlag{
		Flag: cli.Float64Flag{
			Name:   "face-match-dist",
			Usage:  "similarity `OFFSET` for matching faces with existing clusters (0.1-1.5)",
			Value:  face.MatchDist,
			EnvVar: "PHOTOPRISM_FACE_MATCH_DIST",
		},
		Tags: []string{EnvSponsor},
	},
	CliFlag{
		Flag: cli.StringFlag{
			Name:   "pid-filename",
			Usage:  "process id `FILE` *daemon-mode only*",
			EnvVar: "PHOTOPRISM_PID_FILENAME",
		}},
	CliFlag{
		Flag: cli.StringFlag{
			Name:   "log-filename",
			Usage:  "server log `FILE` *daemon-mode only*",
			EnvVar: "PHOTOPRISM_LOG_FILENAME",
			Value:  "",
		}},
}
