package config

import (
	"github.com/urfave/cli"
)

// PhotoPrism command-line parameters and flags.
var GlobalFlags = []cli.Flag{
	cli.BoolFlag{
		Name:   "debug",
		Usage:  "run in debug mode (shows additional log messages)",
		EnvVar: "PHOTOPRISM_DEBUG",
	},
	cli.BoolFlag{
		Name:   "public, p",
		Usage:  "no authentication required (disables password protection)",
		EnvVar: "PHOTOPRISM_PUBLIC",
	},
	cli.BoolFlag{
		Name:   "read-only, r",
		Usage:  "don't modify originals directory (import and upload disabled)",
		EnvVar: "PHOTOPRISM_READONLY",
	},
	cli.BoolFlag{
		Name:   "tf-off",
		Usage:  "don't use TensorFlow for image classification (or anything else)",
		EnvVar: "PHOTOPRISM_TENSORFLOW_OFF",
	},
	cli.BoolFlag{
		Name:   "experimental, e",
		Usage:  "enable experimental features",
		EnvVar: "PHOTOPRISM_EXPERIMENTAL",
	},
	cli.StringFlag{
		Name:   "admin-password",
		Usage:  "admin password",
		Value:  "photoprism",
		EnvVar: "PHOTOPRISM_ADMIN_PASSWORD",
	},
	cli.StringFlag{
		Name:   "webdav-password",
		Usage:  "WebDAV password (none to disable)",
		EnvVar: "PHOTOPRISM_WEBDAV_PASSWORD",
	},
	cli.IntFlag{
		Name:   "workers, w",
		Usage:  "number of workers for indexing",
		EnvVar: "PHOTOPRISM_WORKERS",
	},
	cli.IntFlag{
		Name:   "wakeup-interval",
		Usage:  "background worker wakeup interval in seconds",
		EnvVar: "PHOTOPRISM_WAKEUP_INTERVAL",
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
		Usage:  "http server port `NUMBER``",
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
		Name:   "database-driver",
		Usage:  "database driver `NAME` (sqlite or mysql)",
		Value:  "sqlite",
		EnvVar: "PHOTOPRISM_DATABASE_DRIVER",
	},
	cli.StringFlag{
		Name:   "database-dsn",
		Usage:  "data source or file name (`DSN`)",
		EnvVar: "PHOTOPRISM_DATABASE_DSN",
	},
	cli.IntFlag{
		Name:   "database-conns",
		Usage:  "maximum `NUMBER` of open connections to the database",
		Value:  256,
		EnvVar: "PHOTOPRISM_DATABASE_CONNS",
	},
	cli.StringFlag{
		Name:   "assets-path",
		Usage:  "assets `PATH` for static files like templates and TensorFlow models",
		EnvVar: "PHOTOPRISM_ASSETS_PATH",
	},
	cli.StringFlag{
		Name:   "storage-path",
		Usage:  "storage `PATH` for generated files like cache and index",
		EnvVar: "PHOTOPRISM_STORAGE_PATH",
	},
	cli.StringFlag{
		Name:   "import-path",
		Usage:  "optional import `PATH` for copying files to originals",
		EnvVar: "PHOTOPRISM_IMPORT_PATH",
	},
	cli.StringFlag{
		Name:   "originals-path",
		Usage:  "originals `PATH` for photo, video and sidecar files",
		EnvVar: "PHOTOPRISM_ORIGINALS_PATH",
	},
	cli.IntFlag{
		Name:   "originals-limit",
		Value:  1000,
		Usage:  "file size limit for originals in `MEGABYTE`",
		EnvVar: "PHOTOPRISM_ORIGINALS_LIMIT",
	},
	cli.StringFlag{
		Name:   "log-level, l",
		Usage:  "trace, debug, info, warning, error, fatal or panic",
		Value:  "info",
		EnvVar: "PHOTOPRISM_LOG_LEVEL",
	},
	cli.StringFlag{
		Name:   "log-filename",
		Usage:  "filename for storing server logs",
		EnvVar: "PHOTOPRISM_LOG_FILENAME",
		Value:  "",
	},
	cli.StringFlag{
		Name:   "pid-filename",
		Usage:  "filename for the server process id (pid)",
		EnvVar: "PHOTOPRISM_PID_FILENAME",
	},
	cli.StringFlag{
		Name:   "cache-path",
		Usage:  "cache `PATH`",
		EnvVar: "PHOTOPRISM_CACHE_PATH",
	},
	cli.StringFlag{
		Name:   "temp-path",
		Usage:  "temporary `PATH` for uploads and downloads",
		EnvVar: "PHOTOPRISM_TEMP_PATH",
	},
	cli.StringFlag{
		Name:   "config-file, c",
		Usage:  "load configuration from `FILENAME`",
		EnvVar: "PHOTOPRISM_CONFIG_FILE",
	},
	cli.StringFlag{
		Name:   "settings-path",
		Usage:  "settings `PATH`",
		EnvVar: "PHOTOPRISM_SETTINGS_PATH",
	},
	cli.BoolFlag{
		Name:   "settings-hidden",
		Usage:  "users can not view or change settings",
		EnvVar: "PHOTOPRISM_SETTINGS_HIDDEN",
	},
	cli.StringFlag{
		Name:   "sips-bin",
		Usage:  "sips executable `FILENAME`",
		Value:  "sips",
		EnvVar: "PHOTOPRISM_SIPS_BIN",
	},
	cli.StringFlag{
		Name:   "darktable-bin",
		Usage:  "darktable-cli executable `FILENAME`",
		Value:  "darktable-cli",
		EnvVar: "PHOTOPRISM_DARKTABLE_BIN",
	},
	cli.StringFlag{
		Name:   "heifconvert-bin",
		Usage:  "heif-convert executable `FILENAME`",
		Value:  "heif-convert",
		EnvVar: "PHOTOPRISM_HEIFCONVERT_BIN",
	},
	cli.StringFlag{
		Name:   "ffmpeg-bin",
		Usage:  "ffmpeg executable `FILENAME`",
		Value:  "ffmpeg",
		EnvVar: "PHOTOPRISM_FFMPEG_BIN",
	},
	cli.StringFlag{
		Name:   "exiftool-bin",
		Usage:  "exiftool executable `FILENAME`",
		Value:  "exiftool",
		EnvVar: "PHOTOPRISM_EXIFTOOL_BIN",
	},
	cli.BoolFlag{
		Name:   "sidecar-json, j",
		Usage:  "read metadata from JSON sidecar files created by exiftool",
		EnvVar: "PHOTOPRISM_SIDECAR_JSON",
	},
	cli.BoolFlag{
		Name:   "sidecar-yaml, y",
		Usage:  "backup photo metadata to YAML sidecar files",
		EnvVar: "PHOTOPRISM_SIDECAR_YAML",
	},
	cli.BoolFlag{
		Name:   "sidecar-hidden",
		Usage:  "create JSON and YAML sidecar files in .photoprism if enabled",
		EnvVar: "PHOTOPRISM_SIDECAR_HIDDEN",
	},
	cli.StringFlag{
		Name:   "sidecar-path",
		Usage:  "storage `PATH` for automatically created sidecar files (relative or absolute)",
		EnvVar: "PHOTOPRISM_SIDECAR_PATH",
	},
	cli.BoolFlag{
		Name:   "detect-nsfw",
		Usage:  "flag photos as private that may be offensive",
		EnvVar: "PHOTOPRISM_DETECT_NSFW",
	},
	cli.BoolFlag{
		Name:   "upload-nsfw",
		Usage:  "allow uploads that may be offensive",
		EnvVar: "PHOTOPRISM_UPLOAD_NSFW",
	},
	cli.StringFlag{
		Name:   "geocoding-api, g",
		Usage:  "geocoding api (none, osm or places)",
		Value:  "places",
		EnvVar: "PHOTOPRISM_GEOCODING_API",
	},
	cli.StringFlag{
		Name:   "download-token",
		Usage:  "`SECRET` url token for file downloads",
		EnvVar: "PHOTOPRISM_DOWNLOAD_TOKEN",
	},
	cli.StringFlag{
		Name:   "preview-token",
		Usage:  "`SECRET` url token for thumbnails and video streaming",
		Value:  "static",
		EnvVar: "PHOTOPRISM_PREVIEW_TOKEN",
	},
	cli.StringFlag{
		Name:   "thumb-filter, f",
		Usage:  "resample filter `NAME` (best to worst: blackman, lanczos, cubic, linear)",
		Value:  "lanczos",
		EnvVar: "PHOTOPRISM_THUMB_FILTER",
	},
	cli.BoolFlag{
		Name:   "thumb-uncached, u",
		Usage:  "on-demand rendering of default thumbnails (high memory and cpu usage)",
		EnvVar: "PHOTOPRISM_THUMB_UNCACHED",
	},
	cli.IntFlag{
		Name:   "thumb-size, s",
		Usage:  "default thumbnail size limit in `PIXELS` (720-3840)",
		Value:  2048,
		EnvVar: "PHOTOPRISM_THUMB_SIZE",
	},
	cli.IntFlag{
		Name:   "thumb-limit, x",
		Usage:  "on-demand thumbnail size limit in `PIXELS` (720-3840)",
		Value:  3840,
		EnvVar: "PHOTOPRISM_THUMB_LIMIT",
	},
	cli.IntFlag{
		Name:   "jpeg-quality, q",
		Usage:  "choose 95 for high-quality thumbnails (25-100)",
		Value:  90,
		EnvVar: "PHOTOPRISM_JPEG_QUALITY",
	},
}
