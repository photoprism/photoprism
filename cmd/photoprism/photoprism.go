package main

import (
	"fmt"
	"github.com/photoprism/photoprism"
	"github.com/urfave/cli"
	"os"
)

func main() {
	conf := photoprism.NewConfig()

	app := cli.NewApp()
	app.Name = "PhotoPrism"
	app.Usage = "Digital Photo Archive"
	app.Version = "0.0.3"
	app.Flags = globalCliFlags
	app.Commands = []cli.Command{
		{
			Name:  "config",
			Usage: "Displays global configuration values",
			Action: func(context *cli.Context) error {
				conf.SetValuesFromFile(photoprism.GetExpandedFilename(context.GlobalString("config-file")))

				conf.SetValuesFromCliContext(context)

				fmt.Printf("<name>                <value>\n")
				fmt.Printf("config-file           %s\n", conf.ConfigFile)
				fmt.Printf("darktable-cli         %s\n", conf.DarktableCli)
				fmt.Printf("originals-path        %s\n", conf.OriginalsPath)
				fmt.Printf("thumbnails-path       %s\n", conf.ThumbnailsPath)
				fmt.Printf("import-path           %s\n", conf.ImportPath)
				fmt.Printf("export-path           %s\n", conf.ExportPath)

				return nil
			},
		},
		{
			Name:  "import",
			Usage: "Import photos from directory",
			Action: func(context *cli.Context) error {
				conf.SetValuesFromFile(photoprism.GetExpandedFilename(context.GlobalString("config-file")))

				conf.SetValuesFromCliContext(context)

				fmt.Printf("Importing photos from %s...\n", conf.ImportPath)

				importer := photoprism.NewImporter(conf.OriginalsPath)

				importer.ImportPhotosFromDirectory(conf.ImportPath)

				fmt.Println("Done.")

				return nil
			},
		},
		{
			Name:  "convert",
			Usage: "Convert RAW originals to JPEG",
			Action: func(context *cli.Context) error {
				conf.SetValuesFromFile(photoprism.GetExpandedFilename(context.GlobalString("config-file")))

				conf.SetValuesFromCliContext(context)

				fmt.Printf("Converting RAW images in %s to JPEG...\n", conf.OriginalsPath)

				converter := photoprism.NewConverter(conf.DarktableCli)

				converter.ConvertAll(conf.OriginalsPath)

				fmt.Println("Done.")

				return nil
			},
		},
		{
			Name:  "thumbnails",
			Usage: "Create thumbnails",
			Action: func(context *cli.Context) error {
				conf.SetValuesFromFile(photoprism.GetExpandedFilename(context.GlobalString("config-file")))

				conf.SetValuesFromCliContext(context)

				fmt.Printf("Creating thumbnails in %s...\n", conf.ThumbnailsPath)

				fmt.Println("[TODO]")

				fmt.Println("Done.")

				return nil
			},
		},
		{
			Name:  "export",
			Usage: "Export photos as JPEG",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "from, f",
					Usage: "Start date & time",
					Value: "yesterday",
				},
				cli.StringFlag{
					Name:  "to, t",
					Usage: "End date & time",
					Value: "today",
				},
				cli.StringFlag{
					Name:  "size, s",
					Usage: "Max image size in pixels",
					Value: "4096",
				},
			},
			Action: func(context *cli.Context) error {
				conf.SetValuesFromFile(photoprism.GetExpandedFilename(context.GlobalString("config-file")))

				conf.SetValuesFromCliContext(context)

				fmt.Printf("Exporting photos to %s...\n", conf.ExportPath)

				fmt.Println("[TODO]")

				fmt.Println("Done.")

				return nil
			},
		},
	}

	app.Run(os.Args)
}

var globalCliFlags = []cli.Flag{
	cli.StringFlag{
		Name:  "config-file, c",
		Usage: "Config filename",
		Value: "~/.photoprism",
	},
	cli.StringFlag{
		Name:  "darktable-cli",
		Usage: "Darktable CLI",
		Value: "/Applications/darktable.app/Contents/MacOS/darktable-cli",
	},
	cli.StringFlag{
		Name:  "originals-path",
		Usage: "Originals path",
		Value: "~/Photos/Originals",
	},
	cli.StringFlag{
		Name:  "import-path",
		Usage: "Import path",
		Value: "~/Photos/Import",
	},
	cli.StringFlag{
		Name:  "export-path",
		Usage: "Export path",
		Value: "~/Photos/Export",
	},
	cli.StringFlag{
		Name:  "thumbnails-path",
		Usage: "Thumbnails path",
		Value: "~/Photos/Thumbnails",
	},
}
