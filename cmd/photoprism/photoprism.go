/*

Copyright (c) 2018 - 2022 Michael Mayer <hello@photoprism.app>

    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU Affero General Public License as published
    by the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    You should have received a copy of the GNU Affero General Public License
    along with this program.  If not, see <https://www.gnu.org/licenses/>.

    PhotoPrism® is a registered trademark of Michael Mayer.  You may use it as required
    to describe our software, run your own server, for educational purposes, but not for
    offering commercial goods, products, or services without prior written permission.
    In other words, please ask.

Feel free to send an e-mail to hello@photoprism.app if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
https://docs.photoprism.app/developer-guide/

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
	app := cli.NewApp()
	app.Name = "PhotoPrism"
	app.HelpName = filepath.Base(os.Args[0])
	app.Usage = "Browse Your Life in Pictures"
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
