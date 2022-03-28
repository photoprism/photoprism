/*

Copyright (c) 2018 - 2022 Michael Mayer <hello@photoprism.app>

    This program is free software: you can redistribute it and/or modify
    it under Version 3 of the GNU Affero General Public License (the "AGPL"):
    <https://docs.photoprism.app/license/agpl>

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    The AGPL is supplemented by our Trademark and Brand Guidelines,
    which describe how our Brand Assets may be used:
    <https://photoprism.app/trademark>

Feel free to send an e-mail to hello@photoprism.app if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
<https://docs.photoprism.app/developer-guide/>

*/
package main

import (
	"os"
	"path/filepath"

	"github.com/photoprism/photoprism/internal/commands"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/urfave/cli"
)

var version = "development"
var log = event.Log

const appDescription = "For installation instructions, visit https://docs.photoprism.app/"
const appCopyright = "(c) 2018-2022 Michael Mayer <hello@photoprism.app>"

func main() {
	defer func() {
		if r := recover(); r != nil {
			os.Exit(1)
		}
	}()

	app := cli.NewApp()
	app.Name = "PhotoPrism"
	app.HelpName = filepath.Base(os.Args[0])
	app.Usage = "AI-Powered Photos App"
	app.Description = appDescription
	app.Version = version
	app.Copyright = appCopyright
	app.EnableBashCompletion = true
	app.Flags = config.GlobalFlags

	app.Commands = []cli.Command{
		commands.StartCommand,
		commands.StopCommand,
		commands.StatusCommand,
		commands.IndexCommand,
		commands.ImportCommand,
		commands.CopyCommand,
		commands.FacesCommand,
		commands.PlacesCommand,
		commands.PurgeCommand,
		commands.CleanUpCommand,
		commands.OptimizeCommand,
		commands.MomentsCommand,
		commands.ConvertCommand,
		commands.ThumbsCommand,
		commands.MigrateCommand,
		commands.BackupCommand,
		commands.RestoreCommand,
		commands.ResetCommand,
		commands.PasswdCommand,
		commands.UsersCommand,
		commands.ConfigCommand,
		commands.VersionCommand,
	}

	if err := app.Run(os.Args); err != nil {
		log.Error(err)
	}
}
