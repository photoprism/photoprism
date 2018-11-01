package commands

import (
	"fmt"

	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/urfave/cli"
)

var ConfigCommand = cli.Command{
	Name:   "config",
	Usage:  "Displays global configuration values",
	Action: configAction,
}

// Prints current configuration; called by ConfigCommand
func configAction(context *cli.Context) error {
	conf := photoprism.NewConfig(context)

	fmt.Printf("NAME                  VALUE\n")
	fmt.Printf("debug                 %t\n", conf.Debug)
	fmt.Printf("config-file           %s\n", conf.ConfigFile)
	fmt.Printf("assets-path           %s\n", conf.AssetsPath)
	fmt.Printf("originals-path        %s\n", conf.OriginalsPath)
	fmt.Printf("thumbnails-path       %s\n", conf.ThumbnailsPath)
	fmt.Printf("import-path           %s\n", conf.ImportPath)
	fmt.Printf("export-path           %s\n", conf.ExportPath)
	fmt.Printf("darktable-cli         %s\n", conf.DarktableCli)
	fmt.Printf("database-driver       %s\n", conf.DatabaseDriver)
	fmt.Printf("database-dsn          %s\n", conf.DatabaseDsn)

	return nil
}
