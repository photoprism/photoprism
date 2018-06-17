package main

import (
	"fmt"
	"github.com/photoprism/photoprism"
	"github.com/urfave/cli"
	"os"
	"github.com/araddon/dateparse"
)

func main() {
	conf := photoprism.NewConfig()

	app := cli.NewApp()
	app.Name = "PhotoPrism"
	app.Usage = "Digital Photo Archive"
	app.Version = "0.1.0"
	app.Flags = globalCliFlags
	app.Commands = []cli.Command{
		{
			Name:  "config",
			Usage: "Displays global configuration values",
			Action: func(context *cli.Context) error {
				conf.SetValuesFromFile(photoprism.GetExpandedFilename(context.GlobalString("config-file")))

				conf.SetValuesFromCliContext(context)

				fmt.Printf("NAME                  VALUE\n")
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

				conf.CreateDirectories()

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

				conf.CreateDirectories()

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
			Flags: []cli.Flag{
				cli.IntSliceFlag{
					Name:  "size, s",
					Usage: "Thumbnail size in pixels",
				},
				cli.BoolFlag{
					Name:  "default, d",
					Usage: "Render default sizes: 320, 500, 640, 1280, 1920 and 2560px",
				},
				cli.BoolFlag{
					Name:  "square, q",
					Usage: "Square aspect ratio",
				},
			},
			Action: func(context *cli.Context) error {
				conf.SetValuesFromFile(photoprism.GetExpandedFilename(context.GlobalString("config-file")))

				conf.SetValuesFromCliContext(context)

				conf.CreateDirectories()

				fmt.Printf("Creating thumbnails in %s...\n", conf.ThumbnailsPath)

				sizes := context.IntSlice("size")

				if context.Bool("default") {
					sizes = []int{320, 500, 640, 1280, 1920, 2560}
				}

				if len(sizes) == 0 {
					fmt.Println("No sizes selected. Nothing to do.")
					return nil
				}

				for _, size := range sizes {
					photoprism.CreateThumbnailsFromOriginals(conf.OriginalsPath, conf.ThumbnailsPath, size, context.Bool("square"))
				}

				fmt.Println("Done.")

				return nil
			},
		},
		{
			Name:  "export",
			Usage: "Export photos as JPEG",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "name, n",
					Usage: "Sub-directory name",
				},
				cli.StringFlag{
					Name:  "after, a",
					Usage: "Start date e.g. 2017/04/15",
				},
				cli.StringFlag{
					Name:  "before, b",
					Usage: "End date e.g. 2018/05/02",
				},
				cli.IntFlag{
					Name:  "size, s",
					Usage: "Image size in pixels",
					Value: 2560,
				},
			},
			Action: func(context *cli.Context) error {
				conf.SetValuesFromFile(photoprism.GetExpandedFilename(context.GlobalString("config-file")))

				conf.SetValuesFromCliContext(context)

				conf.CreateDirectories()

				before := context.String("before")
				after := context.String("after")

				if before == "" || after == "" {
					fmt.Println("You need to provide before and after dates for export, e.g.\n\nphotoprism export --after 2018/04/10 --before '2018/04/15 23:00:00'")

					return nil
				}

				afterDate, _ := dateparse.ParseAny(after)
				beforeDate, _ := dateparse.ParseAny(before)
				afterDateFormatted := afterDate.Format("20060102")
				beforeDateFormatted := beforeDate.Format("20060102")

				name := context.String("name")

				if name == "" {
					if afterDateFormatted == beforeDateFormatted {
						name = beforeDateFormatted
					} else {
						name = fmt.Sprintf("%s - %s", afterDateFormatted, beforeDateFormatted)
					}
				}

				exportPath := fmt.Sprintf("%s/%s", conf.ExportPath, name);
				size := context.Int("size")
				originals := photoprism.FindOriginalsByDate(conf.OriginalsPath, afterDate, beforeDate)

				fmt.Printf("Exporting photos to %s...\n", exportPath)

				photoprism.ExportPhotosFromOriginals(originals, conf.ThumbnailsPath, exportPath, size)

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
