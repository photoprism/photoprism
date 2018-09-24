package commands

import "github.com/urfave/cli"

var GlobalFlags = []cli.Flag{
	cli.BoolFlag{
		Name:   "debug",
		Usage:  "run in debug mode",
		EnvVar: "PHOTOPRISM_DEBUG",
	},
	cli.StringFlag{
		Name:   "config-file, c",
		Usage:  "load configuration from `FILENAME`",
		Value:  "/etc/photoprism/photoprism.yml",
		EnvVar: "PHOTOPRISM_CONFIG_FILE",
	},
	cli.StringFlag{
		Name:   "darktable-cli",
		Usage:  "darktable command-line executable `FILENAME`",
		Value:  "/usr/bin/darktable-cli",
		EnvVar: "PHOTOPRISM_DARKTABLE_CLI",
	},
	cli.StringFlag{
		Name:   "originals-path",
		Usage:  "originals `PATH`",
		Value:  "/var/photoprism/photos/originals",
		EnvVar: "PHOTOPRISM_ORIGINALS_PATH",
	},
	cli.StringFlag{
		Name:   "thumbnails-path",
		Usage:  "thumbnails `PATH`",
		Value:  "/var/photoprism/photos/thumbnails",
		EnvVar: "PHOTOPRISM_THUMBNAILS_PATH",
	},
	cli.StringFlag{
		Name:   "import-path",
		Usage:  "import `PATH`",
		Value:  "/var/photoprism/photos/import",
		EnvVar: "PHOTOPRISM_IMPORT_PATH",
	},
	cli.StringFlag{
		Name:   "export-path",
		Usage:  "export `PATH`",
		Value:  "/var/photoprism/photos/export",
		EnvVar: "PHOTOPRISM_EXPORT_PATH",
	},
	cli.StringFlag{
		Name:   "assets-path",
		Usage:  "assets `PATH`",
		Value:  "/var/photoprism",
		EnvVar: "PHOTOPRISM_ASSETS_PATH",
	},
	cli.StringFlag{
		Name:   "database-driver",
		Usage:  "database `DRIVER` (mysql, mssql, postgres or sqlite)",
		Value:  "mysql",
		EnvVar: "PHOTOPRISM_DATABASE_DRIVER",
	},
	cli.StringFlag{
		Name:   "database-dsn",
		Usage:  "database data source name (`DSN`)",
		Value:  "photoprism:photoprism@tcp(localhost:3306)/photoprism",
		EnvVar: "PHOTOPRISM_DATABASE_DSN",
	},
}
