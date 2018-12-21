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
	conf := context.NewConfig(ctx)

	fmt.Printf("NAME                  VALUE\n")
	fmt.Printf("debug                 %t\n", conf.Debug())
	fmt.Printf("config-file           %s\n", conf.ConfigFile())
	fmt.Printf("darktable-cli         %s\n", conf.GetDarktableCli())
	fmt.Printf("originals-path        %s\n", conf.GetOriginalsPath())
	fmt.Printf("import-path           %s\n", conf.GetImportPath())
	fmt.Printf("export-path           %s\n", conf.GetExportPath())
	fmt.Printf("cache-path            %s\n", conf.GetCachePath())
	fmt.Printf("assets-path           %s\n", conf.GetAssetsPath())
	fmt.Printf("database-driver       %s\n", conf.DatabaseDriver())
	fmt.Printf("database-dsn          %s\n", conf.DatabaseDsn())

	return nil
}
