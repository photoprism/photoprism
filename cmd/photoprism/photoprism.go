/*

PhotoPrism is a server-based application for browsing, organizing and sharing
your personal photo collection. It makes use of the latest technologies to
automatically tag and find pictures without getting in your way.

For more information see https://photoprism.org/

Copyright © 2018 The PhotoPrism contributors

Licensed under the Apache License, Version 2.0 (the "License"); you may not use
this application except in compliance with the License.

Unless required by applicable law or agreed to in writing, software distributed
under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR
CONDITIONS OF ANY KIND, either express or implied. See the License for the
specific language governing permissions and limitations under the License.

*/
package main

import (
	"os"

	"github.com/photoprism/photoprism/internal/commands"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "PhotoPrism"
	app.Usage = "Browse your life in pictures"
	app.Version = "0.0.0"
	app.Copyright = "Copyright (c) 2018 The PhotoPrism contributors"
	app.EnableBashCompletion = true
	app.Flags = commands.GlobalFlags

	app.Commands = []cli.Command{
		commands.ConfigCommand,
		commands.StartCommand,
		commands.MigrateCommand,
		commands.ImportCommand,
		commands.IndexCommand,
		commands.ConvertCommand,
		commands.ThumbnailsCommand,
		commands.ExportCommand,
	}

	app.Run(os.Args)
}
