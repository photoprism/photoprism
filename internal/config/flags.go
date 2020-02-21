package config

import (
	"github.com/urfave/cli"
)

// GlobalFlags lists all CLI flags
var GlobalFlags = []cli.Flag{
	cli.BoolFlag{
		Name:   "debug",
		Usage:  "run in debug mode",
		EnvVar: "PHOTOPRISM_DEBUG",
	},
	cli.BoolFlag{
		Name:   "read-only, r",
		Usage:  "run in read-only mode",
		EnvVar: "PHOTOPRISM_READ_ONLY",
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
		Name:   "twitter",
		Usage:  "twitter handle for sharing",
		Value:  "@browseyourlife",
		EnvVar: "PHOTOPRISM_TWITTER",
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
		Value:  "",
		EnvVar: "PHOTOPRISM_WEBDAV_PASSWORD",
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
	cli.StringFlag{
		Name:   "import-path",
		Usage:  "import `PATH`",
		Value:  "~/Pictures/Import",
		EnvVar: "PHOTOPRISM_IMPORT_PATH",
	},
	cli.StringFlag{
		Name:   "export-path",
		Usage:  "export `PATH`",
		Value:  "~/Pictures/Export",
		EnvVar: "PHOTOPRISM_EXPORT_PATH",
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
		Name:   "database-driver",
		Usage:  "database `DRIVER` (internal or mysql)",
		Value:  "internal",
		EnvVar: "PHOTOPRISM_DATABASE_DRIVER",
	},
	cli.StringFlag{
		Name:   "database-dsn",
		Usage:  "database data source name (`DSN`)",
		Value:  "root:@tcp(localhost:4000)/photoprism?parseTime=true",
		EnvVar: "PHOTOPRISM_DATABASE_DSN",
	},
	cli.StringFlag{
		Name:   "sips-bin",
		Usage:  "sips cli binary `FILENAME`",
		Value:  "sips",
		EnvVar: "PHOTOPRISM_SIPS_BIN",
	},
	cli.StringFlag{
		Name:   "darktable-bin",
		Usage:  "darktable cli binary `FILENAME`",
		Value:  "darktable-cli",
		EnvVar: "PHOTOPRISM_DARKTABLE_BIN",
	},
	cli.StringFlag{
		Name:   "exiftool-bin",
		Usage:  "exiftool cli binary `FILENAME`",
		Value:  "exiftool",
		EnvVar: "PHOTOPRISM_EXIFTOOL_BIN",
	},
	cli.StringFlag{
		Name:   "heifconvert-bin",
		Usage:  "heif conversion cli binary `FILENAME`",
		Value:  "heif-convert",
		EnvVar: "PHOTOPRISM_HEIFCONVERT_BIN",
	},
	cli.IntFlag{
		Name:   "http-port",
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
		Name:   "sql-port",
		Usage:  "built-in SQL server port",
		EnvVar: "PHOTOPRISM_SQL_PORT",
	},
	cli.StringFlag{
		Name:   "sql-host",
		Usage:  "built-in SQL server host",
		EnvVar: "PHOTOPRISM_SQL_HOST",
	},
	cli.StringFlag{
		Name:   "sql-path",
		Usage:  "built-in SQL server storage path",
		EnvVar: "PHOTOPRISM_SQL_PATH",
	},
	cli.StringFlag{
		Name:   "sql-password",
		Usage:  "built-in SQL server password",
		EnvVar: "PHOTOPRISM_SQL_PASSWORD",
	},
	cli.BoolFlag{
		Name:   "detect-nsfw",
		Usage:  "flag photos that may be offensive",
		EnvVar: "PHOTOPRISM_DETECT_NSFW",
	},
	cli.BoolFlag{
		Name:   "upload-nsfw",
		Usage:  "allow uploads that may contain offensive content",
		EnvVar: "PHOTOPRISM_UPLOAD_NSFW",
	},
	cli.BoolFlag{
		Name:   "tf-disabled, t",
		Usage:  "don't use TensorFlow for image classification",
		EnvVar: "PHOTOPRISM_TF_DISABLED",
	},
	cli.StringFlag{
		Name:   "geocoding-api, g",
		Usage:  "geocoding api (none, osm or places)",
		Value:  "places",
		EnvVar: "PHOTOPRISM_GEOCODING_API",
	},
	cli.IntFlag{
		Name:   "thumb-quality, q",
		Usage:  "jpeg quality of thumbnails (25-100)",
		Value:  90,
		EnvVar: "PHOTOPRISM_THUMB_QUALITY",
	},
	cli.IntFlag{
		Name:   "thumb-size, s",
		Usage:  "pre-render size limit in pixels (720-3840)",
		Value:  2048,
		EnvVar: "PHOTOPRISM_THUMB_SIZE",
	},
	cli.IntFlag{
		Name:   "thumb-limit, x",
		Usage:  "on-demand size limit in pixels (720-3840)",
		Value:  3840,
		EnvVar: "PHOTOPRISM_THUMB_LIMIT",
	},
	cli.StringFlag{
		Name:   "thumb-filter, f",
		Usage:  "resample filter (blackman, lanczos, cubic or linear)",
		Value:  "lanczos",
		EnvVar: "PHOTOPRISM_THUMB_FILTER",
	},
}
