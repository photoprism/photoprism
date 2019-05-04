package main

import (
	"os"

	"github.com/photoprism/photoprism/internal/commands"
	"github.com/photoprism/photoprism/internal/context"
	"github.com/urfave/cli"
)

var version = "development"

func main() {
	app := cli.NewApp()
	app.Name = "PhotoPrism"
	app.Usage = "Browse your life in pictures"
	app.Version = version
	app.Copyright = "(c) 2018-2019 The PhotoPrism contributors <hello@photoprism.org>"
	app.EnableBashCompletion = true
	app.Flags = context.GlobalFlags

	app.Commands = []cli.Command{
		commands.ConfigCommand,
		commands.StartCommand,
		commands.MigrateCommand,
		commands.ImportCommand,
		commands.IndexCommand,
		commands.ConvertCommand,
		commands.ThumbnailsCommand,
		commands.ExportCommand,
		commands.VersionCommand,
	}

	app.Run(os.Args)
}
