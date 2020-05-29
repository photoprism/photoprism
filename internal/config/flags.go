package config

import (
	"github.com/urfave/cli"
)

// GlobalFlags lists all CLI flags
var GlobalFlags = []cli.Flag{
	cli.StringFlag{
		Name:   "admin-password",
		Usage:  "admin password",
		Value:  "photoprism",
		EnvVar: "PHOTOPRISM_ADMIN_PASSWORD",
	},
	cli.StringFlag{
		Name:   "webdav-password",
		Usage:  "WebDAV password (none to disable)",
		Value:  "",
		EnvVar: "PHOTOPRISM_WEBDAV_PASSWORD",
	},
	cli.BoolFlag{
		Name:   "debug",
		Usage:  "run in debug mode",
		EnvVar: "PHOTOPRISM_DEBUG",
	},
	cli.BoolFlag{
		Name:   "read-only, r",
		Usage:  "run in read-only mode",
		EnvVar: "PHOTOPRISM_READONLY",
	},
	cli.BoolFlag{
		Name:   "public, p",
		Usage:  "no authentication required",
		EnvVar: "PHOTOPRISM_PUBLIC",
	},
	cli.BoolFlag{
		Name:   "experimental, e",
		Usage:  "enable experimental features",
		EnvVar: "PHOTOPRISM_EXPERIMENTAL",
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
		Name:   "url",
		Usage:  "canonical site URL",
		Value:  "http://localhost:2342/",
		EnvVar: "PHOTOPRISM_URL",
	},
	cli.StringFlag{
		Name:   "title",
		Usage:  "site title",
		Value:  "PhotoPrism",
		EnvVar: "PHOTOPRISM_TITLE",
	},
	cli.StringFlag{
		Name:   "subtitle",
		Usage:  "site subtitle",
		Value:  "Browse your life",
		EnvVar: "PHOTOPRISM_SUBTITLE",
	},
	cli.StringFlag{
		Name:   "description",
		Usage:  "site description",
		Value:  "Personal Photo Management",
		EnvVar: "PHOTOPRISM_DESCRIPTION",
	},
	cli.StringFlag{
		Name:   "author",
		Usage:  "site owner / copyright",
		Value:  "Anonymous",
		EnvVar: "PHOTOPRISM_AUTHOR",
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
		Value:  "~/.local/share/photoprism/photoprism.log",
	},
	cli.StringFlag{
		Name:   "pid-filename",
		Usage:  "filename for the server process id (pid)",
		EnvVar: "PHOTOPRISM_PID_FILENAME",
		Value:  "~/.local/share/photoprism/photoprism.pid",
	},
	cli.StringFlag{
		Name:   "config-file, c",
		Usage:  "load configuration from `FILENAME`",
		Value:  "~/.config/photoprism/photoprism.yml",
		EnvVar: "PHOTOPRISM_CONFIG_FILE",
	},
	cli.StringFlag{
		Name:   "config-path",
		Usage:  "config `PATH`",
		Value:  "~/.config/photoprism",
		EnvVar: "PHOTOPRISM_CONFIG_PATH",
	},
	cli.StringFlag{
		Name:   "resources-path",
		Usage:  "resources `PATH`",
		EnvVar: "PHOTOPRISM_RESOURCES_PATH",
	},
	cli.StringFlag{
		Name:   "originals-path",
		Usage:  "originals `PATH`",
		Value:  "~/Pictures/Originals",
		EnvVar: "PHOTOPRISM_ORIGINALS_PATH",
	},
	cli.IntFlag{
		Name:   "originals-limit",
		Value:  1000,
		Usage:  "file `SIZE` limit for originals in MB",
		EnvVar: "PHOTOPRISM_ORIGINALS_LIMIT",
	},
	cli.StringFlag{
		Name:   "import-path",
		Usage:  "import `PATH`",
		Value:  "~/Pictures/Import",
		EnvVar: "PHOTOPRISM_IMPORT_PATH",
	},
	cli.StringFlag{
		Name:   "temp-path",
		Usage:  "temporary `PATH` for uploads and downloads",
		Value:  "",
		EnvVar: "PHOTOPRISM_TEMP_PATH",
	},
	cli.StringFlag{
		Name:   "cache-path",
		Usage:  "cache `PATH`",
		Value:  "~/.cache/photoprism",
		EnvVar: "PHOTOPRISM_CACHE_PATH",
	},
	cli.StringFlag{
		Name:   "assets-path",
		Usage:  "assets `PATH`",
		Value:  "~/.local/share/photoprism",
		EnvVar: "PHOTOPRISM_ASSETS_PATH",
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
	cli.IntFlag{
		Name:   "http-port",
		Value:  2342,
		Usage:  "HTTP server port",
		EnvVar: "PHOTOPRISM_HTTP_PORT",
	},
	cli.StringFlag{
		Name:   "http-host",
		Usage:  "HTTP server host",
		EnvVar: "PHOTOPRISM_HTTP_HOST",
	},
	cli.StringFlag{
		Name:   "http-mode, m",
		Usage:  "debug, release or test",
		EnvVar: "PHOTOPRISM_HTTP_MODE",
	},
	cli.IntFlag{
		Name:   "tidb-port",
		Value:  2343,
		Usage:  "built-in TiDB server port",
		EnvVar: "PHOTOPRISM_TIDB_PORT",
	},
	cli.StringFlag{
		Name:   "tidb-host",
		Usage:  "built-in TiDB server host",
		EnvVar: "PHOTOPRISM_TIDB_HOST",
	},
	cli.StringFlag{
		Name:   "tidb-password",
		Usage:  "built-in TiDB server password",
		EnvVar: "PHOTOPRISM_TIDB_PASSWORD",
	},
	cli.StringFlag{
		Name:   "tidb-path",
		Usage:  "built-in TiDB server storage `PATH`",
		EnvVar: "PHOTOPRISM_TIDB_PATH",
	},
	cli.StringFlag{
		Name:   "database-driver",
		Usage:  "database `DRIVER` (tidb or mysql)",
		Value:  "tidb",
		EnvVar: "PHOTOPRISM_DATABASE_DRIVER",
	},
	cli.StringFlag{
		Name:   "database-dsn",
		Usage:  "database data source name (`DSN`)",
		Value:  "root:@tcp(localhost:2343)/photoprism?parseTime=true",
		EnvVar: "PHOTOPRISM_DATABASE_DSN",
	},
	cli.IntFlag{
		Name:   "database-conns",
		Usage:  "maximum `NUMBER` of open connections to the database",
		Value:  256,
		EnvVar: "PHOTOPRISM_DATABASE_CONNS",
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
		Usage:  "url `TOKEN` for file downloads",
		Value:  "",
		EnvVar: "PHOTOPRISM_DOWNLOAD_TOKEN",
	},
	cli.StringFlag{
		Name:   "preview-token",
		Usage:  "url `TOKEN` for thumbnails and video streaming",
		Value:  "static",
		EnvVar: "PHOTOPRISM_PREVIEW_TOKEN",
	},
	cli.StringFlag{
		Name:   "thumb-filter, f",
		Usage:  "resample filter (best to worst: blackman, lanczos, cubic, linear)",
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
		Usage:  "default thumbnail size limit in pixels (720-3840)",
		Value:  2048,
		EnvVar: "PHOTOPRISM_THUMB_SIZE",
	},
	cli.IntFlag{
		Name:   "thumb-limit, x",
		Usage:  "on-demand thumbnail size limit in pixels (720-3840)",
		Value:  3840,
		EnvVar: "PHOTOPRISM_THUMB_LIMIT",
	},
	cli.IntFlag{
		Name:   "jpeg-quality, q",
		Usage:  "set to 95 for high-quality thumbnails (25-100)",
		Value:  90,
		EnvVar: "PHOTOPRISM_JPEG_QUALITY",
	},
	cli.BoolFlag{
		Name:   "jpeg-hidden",
		Usage:  "create JPEG files in .photoprism when converting other file types",
		EnvVar: "PHOTOPRISM_JPEG_HIDDEN",
	},
	cli.BoolFlag{
		Name:   "disable-tf",
		Usage:  "don't use TensorFlow for image classification",
		EnvVar: "PHOTOPRISM_DISABLE_TF",
	},
	cli.BoolFlag{
		Name:   "disable-settings",
		Usage:  "user can not change settings",
		EnvVar: "PHOTOPRISM_DISABLE_SETTINGS",
	},
}
