package main

import (
	"fmt"
	"github.com/araddon/dateparse"
	"github.com/photoprism/photoprism"
	"github.com/photoprism/photoprism/server"
	"github.com/urfave/cli"
	"os"
)

func main() {
	conf := photoprism.NewConfig()

	app := cli.NewApp()
	app.Name = "PhotoPrism"
	app.Usage = "Digital Photo Archive"
	app.Version = "0.0.0"
	app.Flags = globalCliFlags
	app.Commands = []cli.Command{
		{
			Name:  "config",
			Usage: "Displays global configuration values",
			Action: func(context *cli.Context) error {
				conf.SetValuesFromFile(photoprism.GetExpandedFilename(context.GlobalString("config-file")))

				conf.SetValuesFromCliContext(context)

				fmt.Printf("NAME                  VALUE\n")
				fmt.Printf("debug                 %t\n", conf.Debug)
				fmt.Printf("config-file           %s\n", conf.ConfigFile)
				fmt.Printf("server-ip             %s\n", conf.ServerIP)
				fmt.Printf("server-port           %d\n", conf.ServerPort)
				fmt.Printf("server-mode           %s\n", conf.ServerMode)
				fmt.Printf("server-assets-path    %s\n", conf.ServerAssetsPath)
				fmt.Printf("darktable-cli         %s\n", conf.DarktableCli)
				fmt.Printf("originals-path        %s\n", conf.OriginalsPath)
				fmt.Printf("thumbnails-path       %s\n", conf.ThumbnailsPath)
				fmt.Printf("import-path           %s\n", conf.ImportPath)
				fmt.Printf("export-path           %s\n", conf.ExportPath)

				return nil
			},
		},
		{
			Name:  "start",
			Usage: "Starts web server",
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "server-port, p",
					Usage: "HTTP server port",
					Value: 80,
				},
				cli.StringFlag{
					Name:  "server-ip, i",
					Usage: "HTTP server IP address (optional)",
					Value: "",
				},
				cli.StringFlag{
					Name:  "server-mode, m",
					Usage: "debug, release or test",
					Value: "",
				},
			},
			Action: func(context *cli.Context) error {
				conf.SetValuesFromFile(photoprism.GetExpandedFilename(context.GlobalString("config-file")))

				conf.SetValuesFromCliContext(context)

				if context.IsSet("server-ip") {
					conf.ServerIP = context.String("server-ip")
				}

				if context.IsSet("server-port") {
					conf.ServerPort = context.Int("server-port")
				}

				if context.IsSet("server-mode") {
					conf.ServerMode = context.String("server-mode")
				}

				conf.CreateDirectories()

				conf.MigrateDb()

				fmt.Printf("Starting web server at port %d...\n", context.Int("port"))

				server.Start(conf)

				fmt.Println("Done.")

				return nil
			},
		},
		{
			Name:  "migrate-db",
			Usage: "Automatically migrates / initializes database",
			Action: func(context *cli.Context) error {
				conf.SetValuesFromFile(photoprism.GetExpandedFilename(context.GlobalString("config-file")))

				conf.SetValuesFromCliContext(context)

				fmt.Println("Migrating database...")

				conf.MigrateDb()

				fmt.Println("Done.")

				return nil
			},
		},
		{
			Name:  "import",
			Usage: "Imports photos",
			Action: func(context *cli.Context) error {
				conf.SetValuesFromFile(photoprism.GetExpandedFilename(context.GlobalString("config-file")))

				conf.SetValuesFromCliContext(context)

				conf.CreateDirectories()

				conf.MigrateDb()

				fmt.Printf("Importing photos from %s...\n", conf.ImportPath)

				indexer := photoprism.NewIndexer(conf.OriginalsPath, conf.GetDb())

				importer := photoprism.NewImporter(conf.OriginalsPath, indexer)

				importer.ImportPhotosFromDirectory(conf.ImportPath)

				fmt.Println("Done.")

				return nil
			},
		},
		{
			Name:  "index",
			Usage: "Re-indexes all originals",
			Action: func(context *cli.Context) error {
				conf.SetValuesFromFile(photoprism.GetExpandedFilename(context.GlobalString("config-file")))

				conf.SetValuesFromCliContext(context)

				conf.CreateDirectories()

				conf.MigrateDb()

				fmt.Printf("Indexing photos in %s...\n", conf.OriginalsPath)

				indexer := photoprism.NewIndexer(conf.OriginalsPath, conf.GetDb())

				indexer.IndexAll()

				fmt.Println("Done.")

				return nil
			},
		},
		{
			Name:  "convert",
			Usage: "Converts RAW originals to JPEG",
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
			Usage: "Creates thumbnails",
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
			Usage: "Exports photos as JPEG",
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

				exportPath := fmt.Sprintf("%s/%s", conf.ExportPath, name)
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
	cli.BoolFlag{
		Name:  "debug",
		Usage: "run in debug mode",
	},
	cli.StringFlag{
		Name:  "config-file, c",
		Usage: "config filename",
		Value: "/etc/photoprism/config.yml",
	},
	cli.StringFlag{
		Name:  "darktable-cli",
		Usage: "darktable CLI",
		Value: "/usr/bin/darktable-cli",
	},
	cli.StringFlag{
		Name:  "originals-path",
		Usage: "originals path",
		Value: "/var/photoprism/originals",
	},
	cli.StringFlag{
		Name:  "thumbnails-path",
		Usage: "thumbnails path",
		Value: "/var/photoprism/thumbnails",
	},
	cli.StringFlag{
		Name:  "import-path",
		Usage: "import path",
		Value: "/var/photoprism/import",
	},
	cli.StringFlag{
		Name:  "export-path",
		Usage: "export path",
		Value: "/var/photoprism/export",
	},
	cli.StringFlag{
		Name:  "server-assets-path",
		Usage: "server assets path for templates, js and css",
		Value: "/var/photoprism/server",
	},
	cli.StringFlag{
		Name:  "database-driver",
		Usage: "database driver (mysql, mssql, postgres or sqlite)",
		Value: "mysql",
	},
	cli.StringFlag{
		Name:  "database-dsn",
		Usage: "database data source name (DSN)",
		Value: "photoprism:photoprism@tcp(localhost:3306)/photoprism",
	},
}
