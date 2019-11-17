package commands

import (
	"fmt"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/urfave/cli"
)

// Prints current configuration
var ConfigCommand = cli.Command{
	Name:   "config",
	Usage:  "Displays global configuration values",
	Action: configAction,
}

func configAction(ctx *cli.Context) error {
	conf := config.NewConfig(ctx)

	fmt.Printf("NAME                  VALUE\n")
	fmt.Printf("name                  %s\n", conf.Name())
	fmt.Printf("version               %s\n", conf.Version())
	fmt.Printf("copyright             %s\n", conf.Copyright())
	fmt.Printf("debug                 %t\n", conf.Debug())
	fmt.Printf("read-only             %t\n", conf.ReadOnly())
	fmt.Printf("public                %t\n", conf.Public())
	fmt.Printf("admin-password        %s\n", conf.AdminPassword())
	fmt.Printf("log-level             %s\n", conf.LogLevel())
	fmt.Printf("log-filename          %s\n", conf.LogFilename())
	fmt.Printf("pid-filename          %s\n", conf.PIDFilename())
	fmt.Printf("config-file           %s\n", conf.ConfigFile())
	fmt.Printf("config-path           %s\n", conf.ConfigPath())

	fmt.Printf("database-driver       %s\n", conf.DatabaseDriver())
	fmt.Printf("database-dsn          %s\n", conf.DatabaseDsn())

	fmt.Printf("sql-host              %s\n", conf.SqlServerHost())
	fmt.Printf("sql-port              %d\n", conf.SqlServerPort())
	fmt.Printf("sql-password          %s\n", conf.SqlServerPassword())
	fmt.Printf("sql-path              %s\n", conf.SqlServerPath())

	fmt.Printf("http-host             %s\n", conf.HttpServerHost())
	fmt.Printf("http-port             %d\n", conf.HttpServerPort())
	fmt.Printf("http-mode             %s\n", conf.HttpServerMode())

	fmt.Printf("assets-path           %s\n", conf.AssetsPath())
	fmt.Printf("originals-path        %s\n", conf.OriginalsPath())
	fmt.Printf("import-path           %s\n", conf.ImportPath())
	fmt.Printf("export-path           %s\n", conf.ExportPath())
	fmt.Printf("cache-path            %s\n", conf.CachePath())
	fmt.Printf("thumbnails-path       %s\n", conf.ThumbnailsPath())
	fmt.Printf("resources-path        %s\n", conf.ResourcesPath())
	fmt.Printf("tf-version            %s\n", conf.TensorFlowVersion())
	fmt.Printf("tf-model-path         %s\n", conf.TensorFlowModelPath())
	fmt.Printf("templates-path        %s\n", conf.HttpTemplatesPath())
	fmt.Printf("favicons-path         %s\n", conf.HttpFaviconsPath())
	fmt.Printf("static-path           %s\n", conf.HttpStaticPath())
	fmt.Printf("static-build-path     %s\n", conf.HttpStaticBuildPath())

	fmt.Printf("sips-bin              %s\n", conf.SipsBin())
	fmt.Printf("darktable-bin         %s\n", conf.DarktableBin())
	fmt.Printf("exiftool-bin          %s\n", conf.ExifToolBin())
	fmt.Printf("heifconvert-bin       %s\n", conf.HeifConvertBin())

	return nil
}
