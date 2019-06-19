package main

import (
	"os"

	"github.com/photoprism/photoprism/internal/commands"
	"github.com/photoprism/photoprism/internal/config"
	log "github.com/sirupsen/logrus"
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
	app.Flags = config.GlobalFlags

	app.Commands = []cli.Command{
		commands.ConfigCommand,
		commands.StartCommand,
		commands.StopCommand,
		commands.MigrateCommand,
		commands.ImportCommand,
		commands.IndexCommand,
		commands.ConvertCommand,
		commands.ThumbnailsCommand,
		commands.VersionCommand,
	}

	if err := app.Run(os.Args); err != nil {
		log.Error(err)
	}
}
