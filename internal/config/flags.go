package config

import "github.com/urfave/cli"

// Global CLI flags
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
	cli.StringFlag{
		Name:   "log-level, l",
		Usage:  "trace, debug, info, warning, error, fatal or panic",
		Value:  "info",
		EnvVar: "PHOTOPRISM_LOG_LEVEL",
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
		Value:  "/srv/photoprism/photos/originals",
		EnvVar: "PHOTOPRISM_ORIGINALS_PATH",
	},
	cli.StringFlag{
		Name:   "import-path",
		Usage:  "import `PATH`",
		Value:  "/srv/photoprism/photos/import",
		EnvVar: "PHOTOPRISM_IMPORT_PATH",
	},
	cli.StringFlag{
		Name:   "export-path",
		Usage:  "export `PATH`",
		Value:  "/srv/photoprism/photos/export",
		EnvVar: "PHOTOPRISM_EXPORT_PATH",
	},
	cli.StringFlag{
		Name:   "cache-path",
		Usage:  "cache `PATH`",
		Value:  "/srv/photoprism/cache",
		EnvVar: "PHOTOPRISM_CACHE_PATH",
	},
	cli.StringFlag{
		Name:   "assets-path",
		Usage:  "assets `PATH`",
		Value:  "/srv/photoprism",
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
	cli.IntFlag{
		Name:   "http-port, p",
		Usage:  "HTTP server port",
		Value:  80,
		EnvVar: "PHOTOPRISM_HTTP_PORT",
	},
	cli.StringFlag{
		Name:   "http-host, i",
		Usage:  "HTTP server host",
		Value:  "",
		EnvVar: "PHOTOPRISM_HTTP_HOST",
	},
	cli.StringFlag{
		Name:   "http-mode, m",
		Usage:  "debug, release or test",
		Value:  "",
		EnvVar: "PHOTOPRISM_HTTP_MODE",
	},
	cli.StringFlag{
		Name:   "http-password",
		Usage:  "HTTP server password (optional)",
		Value:  "",
		EnvVar: "PHOTOPRISM_HTTP_PASSWORD",
	},
	cli.IntFlag{
		Name:   "sql-port, s",
		Usage:  "built-in SQL server port",
		Value:  4000,
		EnvVar: "PHOTOPRISM_SQL_PORT",
	},
	cli.StringFlag{
		Name:   "sql-host",
		Usage:  "built-in SQL server host",
		Value:  "",
		EnvVar: "PHOTOPRISM_SQL_HOST",
	},
	cli.StringFlag{
		Name:   "sql-path",
		Usage:  "built-in SQL server storage path",
		Value:  "",
		EnvVar: "PHOTOPRISM_SQL_PATH",
	},
	cli.StringFlag{
		Name:   "sql-password",
		Usage:  "built-in SQL server password",
		Value:  "",
		EnvVar: "PHOTOPRISM_SQL_PASSWORD",
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
}
