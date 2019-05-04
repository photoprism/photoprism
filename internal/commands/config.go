package commands

import (
	"fmt"

	"github.com/photoprism/photoprism/internal/context"
	"github.com/urfave/cli"
)

// Prints current configuration
var ConfigCommand = cli.Command{
	Name:   "config",
	Usage:  "Displays global configuration values",
	Action: configAction,
}

func configAction(ctx *cli.Context) error {
	app := context.NewContext(ctx)

	fmt.Printf("NAME                  VALUE\n")
	fmt.Printf("name                  %s\n", app.Name())
	fmt.Printf("version               %s\n", app.Version())
	fmt.Printf("copyright             %s\n", app.Copyright())
	fmt.Printf("debug                 %t\n", app.Debug())
	fmt.Printf("read-only             %t\n", app.ReadOnly())
	fmt.Printf("log-level             %s\n", app.LogLevel())
	fmt.Printf("config-file           %s\n", app.ConfigFile())

	fmt.Printf("database-driver       %s\n", app.DatabaseDriver())
	fmt.Printf("database-dsn          %s\n", app.DatabaseDsn())

	fmt.Printf("http-host             %s\n", app.HttpServerHost())
	fmt.Printf("http-port             %d\n", app.HttpServerPort())
	fmt.Printf("http-mode             %s\n", app.HttpServerMode())

	fmt.Printf("sql-host              %s\n", app.SqlServerHost())
	fmt.Printf("sql-port              %d\n", app.SqlServerPort())
	fmt.Printf("sql-password          %s\n", app.SqlServerPassword())
	fmt.Printf("sql-path              %s\n", app.SqlServerPath())

	fmt.Printf("assets-path           %s\n", app.AssetsPath())
	fmt.Printf("originals-path        %s\n", app.OriginalsPath())
	fmt.Printf("import-path           %s\n", app.ImportPath())
	fmt.Printf("export-path           %s\n", app.ExportPath())
	fmt.Printf("cache-path            %s\n", app.CachePath())
	fmt.Printf("thumbnails-path       %s\n", app.ThumbnailsPath())
	fmt.Printf("tf-model-path         %s\n", app.TensorFlowModelPath())
	fmt.Printf("templates-path        %s\n", app.HttpTemplatesPath())
	fmt.Printf("favicons-path         %s\n", app.HttpFaviconsPath())
	fmt.Printf("public-path           %s\n", app.HttpPublicPath())
	fmt.Printf("public-build-path     %s\n", app.HttpPublicBuildPath())

	fmt.Printf("darktable-cli         %s\n", app.DarktableCli())

	return nil
}
