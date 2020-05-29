package main

import (
	"os"

	"github.com/photoprism/photoprism/internal/commands"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/urfave/cli"
)

var version = "development"
var log = event.Log

func main() {
	app := cli.NewApp()
	app.Name = "PhotoPrism"
	app.Usage = "Browse your life in pictures"
	app.Version = version
	app.Copyright = "(c) 2018-2020 PhotoPrism.org <hello@photoprism.org>"
	app.EnableBashCompletion = true
	app.Flags = config.GlobalFlags

	app.Commands = []cli.Command{
		commands.StartCommand,
		commands.StopCommand,
		commands.IndexCommand,
		commands.ImportCommand,
		commands.MomentsCommand,
		commands.PurgeCommand,
		commands.CopyCommand,
		commands.ConvertCommand,
		commands.ResampleCommand,
		commands.MigrateCommand,
		commands.ConfigCommand,
		commands.VersionCommand,
		commands.StatusCommand,
	}

	if err := app.Run(os.Args); err != nil {
		log.Error(err)
	}
}
