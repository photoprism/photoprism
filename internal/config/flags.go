package config

import (
	"github.com/klauspost/cpuid/v2"
	"github.com/urfave/cli"
)

// PhotoPrism command-line parameters and flags.
var GlobalFlags = []cli.Flag{
	cli.BoolFlag{
		Name:   "debug",
		Usage:  "run in debug mode, shows additional log messages",
		EnvVar: "PHOTOPRISM_DEBUG",
	},
	cli.BoolFlag{
		Name:   "test",
		Hidden: true,
		Usage:  "run in test mode",
	},
	cli.BoolFlag{
		Name:   "demo",
		Hidden: true,
		Usage:  "run in demo mode",
		EnvVar: "PHOTOPRISM_DEMO",
	},
	cli.BoolFlag{
		Name:   "sponsor",
		Hidden: true,
		Usage:  "your continuous support helps to pay for development and operating expenses",
		EnvVar: "PHOTOPRISM_SPONSOR",
	},
	cli.BoolFlag{
		Name:   "public, p",
		Usage:  "no authentication required, disables password protection",
		EnvVar: "PHOTOPRISM_PUBLIC",
	},
	cli.BoolFlag{
		Name:   "read-only, r",
		Usage:  "don't modify originals folder; disables import, upload, and delete",
		EnvVar: "PHOTOPRISM_READONLY",
	},
	cli.BoolFlag{
		Name:   "experimental, e",
		Usage:  "enable experimental features",
		EnvVar: "PHOTOPRISM_EXPERIMENTAL",
	},
	cli.StringFlag{
		Name:   "admin-password",
		Usage:  "initial admin `PASSWORD`, min 4 characters",
		EnvVar: "PHOTOPRISM_ADMIN_PASSWORD",
	},
	cli.StringFlag{
		Name:   "config-file, c",
		Usage:  "load initial config options from `FILENAME`",
		EnvVar: "PHOTOPRISM_CONFIG_FILE",
	},
	cli.StringFlag{
		Name:   "config-path",
		Usage:  "config `PATH` containing application settings",
		EnvVar: "PHOTOPRISM_CONFIG_PATH",
	},
	cli.StringFlag{
		Name:   "originals-path",
		Usage:  "originals `PATH` containing your photo and video collection",
		EnvVar: "PHOTOPRISM_ORIGINALS_PATH",
	},
	cli.IntFlag{
		Name:   "originals-limit",
		Value:  1000,
		Usage:  "file size limit for originals in `MB`",
		EnvVar: "PHOTOPRISM_ORIGINALS_LIMIT",
	},
	cli.StringFlag{
		Name:   "import-path",
		Usage:  "optional `PATH` for importing files to originals",
		EnvVar: "PHOTOPRISM_IMPORT_PATH",
	},
	cli.StringFlag{
		Name:   "storage-path",
		Usage:  "storage `PATH` for cache, database and sidecar files",
		EnvVar: "PHOTOPRISM_STORAGE_PATH",
	},
	cli.StringFlag{
		Name:   "sidecar-path",
		Usage:  "relative or absolute storage `PATH` for sidecar files",
		EnvVar: "PHOTOPRISM_SIDECAR_PATH",
	},
	cli.StringFlag{
		Name:   "cache-path",
		Usage:  "cache storage `PATH` for sessions and thumbnails",
		EnvVar: "PHOTOPRISM_CACHE_PATH",
	},
	cli.StringFlag{
		Name:   "temp-path",
		Usage:  "temporary `PATH` for storing uploads and downloads",
		EnvVar: "PHOTOPRISM_TEMP_PATH",
	},
	cli.StringFlag{
		Name:   "backup-path",
		Usage:  "backup storage `PATH`",
		EnvVar: "PHOTOPRISM_BACKUP_PATH",
	},
	cli.StringFlag{
		Name:   "assets-path",
		Usage:  "assets `PATH` for static resources like models and templates",
		EnvVar: "PHOTOPRISM_ASSETS_PATH",
	},
	cli.IntFlag{
		Name:   "workers, w",
		Usage:  "adjusts `MAX` number of indexing workers",
		EnvVar: "PHOTOPRISM_WORKERS",
		Value:  cpuid.CPU.PhysicalCores / 2,
	},
	cli.IntFlag{
		Name:   "wakeup-interval",
		Usage:  "background worker wakeup interval in `SECONDS`",
		EnvVar: "PHOTOPRISM_WAKEUP_INTERVAL",
	},
	cli.IntFlag{
		Name:   "auto-index",
		Usage:  "auto indexing safety delay in `SECONDS` (WebDAV)",
		EnvVar: "PHOTOPRISM_AUTO_INDEX",
	},
	cli.IntFlag{
		Name:   "auto-import",
		Usage:  "auto importing safety delay in `SECONDS` (WebDAV)",
		EnvVar: "PHOTOPRISM_AUTO_IMPORT",
	},
	cli.BoolFlag{
		Name:   "disable-backups",
		Usage:  "don't backup photo and album metadata to YAML files",
		EnvVar: "PHOTOPRISM_DISABLE_BACKUPS",
	},
	cli.BoolFlag{
		Name:   "disable-webdav",
		Usage:  "disable built-in WebDAV server",
		EnvVar: "PHOTOPRISM_DISABLE_WEBDAV",
	},
	cli.BoolFlag{
		Name:   "disable-settings",
		Usage:  "users can not view or change settings",
		EnvVar: "PHOTOPRISM_DISABLE_SETTINGS",
	},
	cli.BoolFlag{
		Name:   "disable-places",
		Usage:  "disables reverse geocoding and maps",
		EnvVar: "PHOTOPRISM_DISABLE_PLACES",
	},
	cli.BoolFlag{
		Name:   "disable-exiftool",
		Usage:  "don't create ExifTool JSON files for enhanced metadata extraction",
		EnvVar: "PHOTOPRISM_DISABLE_EXIFTOOL",
	},
	cli.BoolFlag{
		Name:   "disable-tensorflow",
		Usage:  "don't use TensorFlow for image classification",
		EnvVar: "PHOTOPRISM_DISABLE_TENSORFLOW",
	},
	cli.BoolFlag{
		Name:   "detect-nsfw",
		Usage:  "flag photos as private that may be offensive (requires TensorFlow)",
		EnvVar: "PHOTOPRISM_DETECT_NSFW",
	},
	cli.BoolFlag{
		Name:   "upload-nsfw",
		Usage:  "allow uploads that may be offensive",
		EnvVar: "PHOTOPRISM_UPLOAD_NSFW",
	},
	cli.StringFlag{
		Name:   "log-level, l",
		Usage:  "trace, debug, info, warning, error, fatal or panic",
		Value:  "info",
		EnvVar: "PHOTOPRISM_LOG_LEVEL",
	},
	cli.StringFlag{
		Name:   "log-filename",
		Usage:  "server log `FILENAME`",
		EnvVar: "PHOTOPRISM_LOG_FILENAME",
		Value:  "",
	},
	cli.StringFlag{
		Name:   "pid-filename",
		Usage:  "server process id `FILENAME`",
		EnvVar: "PHOTOPRISM_PID_FILENAME",
	},
	cli.StringFlag{
		Name:   "site-url",
		Usage:  "public site `URL`",
		Value:  "http://localhost:2342/",
		EnvVar: "PHOTOPRISM_SITE_URL",
	},
	cli.StringFlag{
		Name:   "site-preview",
		Usage:  "public preview image `URL`",
		EnvVar: "PHOTOPRISM_SITE_PREVIEW",
	},
	cli.StringFlag{
		Name:   "site-title",
		Usage:  "site title",
		Value:  "PhotoPrism",
		EnvVar: "PHOTOPRISM_SITE_TITLE",
	},
	cli.StringFlag{
		Name:   "site-caption",
		Usage:  "short site caption",
		Value:  "Browse Your Life",
		EnvVar: "PHOTOPRISM_SITE_CAPTION",
	},
	cli.StringFlag{
		Name:   "site-description",
		Usage:  "long site description",
		EnvVar: "PHOTOPRISM_SITE_DESCRIPTION",
	},
	cli.StringFlag{
		Name:   "site-author",
		Usage:  "site artist or copyright",
		EnvVar: "PHOTOPRISM_SITE_AUTHOR",
	},
	cli.IntFlag{
		Name:   "http-port",
		Value:  2342,
		Usage:  "http server port `NUMBER`",
		EnvVar: "PHOTOPRISM_HTTP_PORT",
	},
	cli.StringFlag{
		Name:   "http-host",
		Usage:  "http server `IP` address",
		EnvVar: "PHOTOPRISM_HTTP_HOST",
	},
	cli.StringFlag{
		Name:   "http-mode, m",
		Usage:  "debug, release or test",
		EnvVar: "PHOTOPRISM_HTTP_MODE",
	},
	cli.StringFlag{
		Name:   "http-compression, z",
		Usage:  "improves transfer speed and bandwidth utilization (none or gzip)",
		EnvVar: "PHOTOPRISM_HTTP_COMPRESSION",
	},
	cli.StringFlag{
		Name:   "database-driver",
		Usage:  "database driver `NAME` (sqlite or mysql)",
		Value:  "sqlite",
		EnvVar: "PHOTOPRISM_DATABASE_DRIVER",
	},
	cli.StringFlag{
		Name:   "database-dsn",
		Usage:  "sqlite file name, specifying a `DSN` is optional for mariadb and mysql",
		EnvVar: "PHOTOPRISM_DATABASE_DSN",
	},
	cli.StringFlag{
		Name:   "database-server",
		Usage:  "database server `HOST`, specifying a :port is optional",
		EnvVar: "PHOTOPRISM_DATABASE_SERVER",
	},
	cli.StringFlag{
		Name:   "database-name",
		Value:  "photoprism",
		Usage:  "database schema `NAME`",
		EnvVar: "PHOTOPRISM_DATABASE_NAME",
	},
	cli.StringFlag{
		Name:   "database-user",
		Value:  "photoprism",
		Usage:  "database user `NAME`",
		EnvVar: "PHOTOPRISM_DATABASE_USER",
	},
	cli.StringFlag{
		Name:   "database-password",
		Usage:  "database user `PASSWORD`",
		EnvVar: "PHOTOPRISM_DATABASE_PASSWORD",
	},
	cli.IntFlag{
		Name:   "database-conns",
		Usage:  "`LIMIT` the number of open database connections",
		EnvVar: "PHOTOPRISM_DATABASE_CONNS",
	},
	cli.IntFlag{
		Name:   "database-conns-idle",
		Usage:  "`LIMIT` the number of idle database connections",
		EnvVar: "PHOTOPRISM_DATABASE_CONNS_IDLE",
	},
	cli.StringFlag{
		Name:   "rawtherapee-bin",
		Usage:  "RawTherapee CLI `COMMAND` for raw image conversion",
		Value:  "rawtherapee-cli",
		EnvVar: "PHOTOPRISM_RAWTHERAPEE_BIN",
	},
	cli.StringFlag{
		Name:   "darktable-bin",
		Usage:  "Darktable CLI `COMMAND` for raw image conversion",
		Value:  "darktable-cli",
		EnvVar: "PHOTOPRISM_DARKTABLE_BIN",
	},
	cli.BoolFlag{
		Name:   "darktable-presets",
		Usage:  "enables Darktable presets and disables concurrent RAW conversion",
		EnvVar: "PHOTOPRISM_DARKTABLE_PRESETS",
	},
	cli.StringFlag{
		Name:   "sips-bin",
		Usage:  "Scriptable Image Processing System `COMMAND`",
		Value:  "sips",
		EnvVar: "PHOTOPRISM_SIPS_BIN",
	},
	cli.StringFlag{
		Name:   "heifconvert-bin",
		Usage:  "HEIC/HEIF image convert `COMMAND`",
		Value:  "heif-convert",
		EnvVar: "PHOTOPRISM_HEIFCONVERT_BIN",
	},
	cli.StringFlag{
		Name:   "ffmpeg-bin",
		Usage:  "FFmpeg `COMMAND` for video transcoding and cover images",
		Value:  "ffmpeg",
		EnvVar: "PHOTOPRISM_FFMPEG_BIN",
	},
	cli.StringFlag{
		Name:   "exiftool-bin",
		Usage:  "ExifTool `COMMAND` for enhanced metadata extraction",
		Value:  "exiftool",
		EnvVar: "PHOTOPRISM_EXIFTOOL_BIN",
	},
	cli.StringFlag{
		Name:   "download-token",
		Usage:  "optional static `SECRET` url token for file downloads",
		EnvVar: "PHOTOPRISM_DOWNLOAD_TOKEN",
	},
	cli.StringFlag{
		Name:   "preview-token",
		Usage:  "optional static `SECRET` url token for preview images and video streaming",
		EnvVar: "PHOTOPRISM_PREVIEW_TOKEN",
	},
	cli.StringFlag{
		Name:   "thumb-filter, f",
		Usage:  "downscaling filter `NAME` (best to worst: blackman, lanczos, cubic, linear)",
		Value:  "lanczos",
		EnvVar: "PHOTOPRISM_THUMB_FILTER",
	},
	cli.IntFlag{
		Name:   "thumb-size, s",
		Usage:  "static thumbnail size limit in `PIXELS` (720-7680)",
		Value:  2048,
		EnvVar: "PHOTOPRISM_THUMB_SIZE",
	},
	cli.BoolFlag{
		Name:   "thumb-uncached, u",
		Usage:  "enable dynamic thumbnail rendering (high memory and cpu usage)",
		EnvVar: "PHOTOPRISM_THUMB_UNCACHED",
	},
	cli.IntFlag{
		Name:   "thumb-size-uncached, x",
		Usage:  "dynamic rendering size limit in `PIXELS` (720-7680)",
		Value:  7680,
		EnvVar: "PHOTOPRISM_THUMB_SIZE_UNCACHED",
	},
	cli.IntFlag{
		Name:   "jpeg-size",
		Usage:  "size limit for converted image files in `PIXELS` (720-30000)",
		Value:  7680,
		EnvVar: "PHOTOPRISM_JPEG_SIZE",
	},
	cli.IntFlag{
		Name:   "jpeg-quality, q",
		Usage:  "choose 95 for high-quality thumbnails (25-100)",
		Value:  92,
		EnvVar: "PHOTOPRISM_JPEG_QUALITY",
	},
}
