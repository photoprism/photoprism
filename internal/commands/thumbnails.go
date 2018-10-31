package commands

import (
	"fmt"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/urfave/cli"
	"log"
)

var ThumbnailsCommand = cli.Command{
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
	Action: thumbnailsAction,
}

// Creates thumbnail; called by ThumbnailsCommand
func thumbnailsAction(context *cli.Context) error {
	conf := photoprism.NewConfig(context)

	if err := conf.CreateDirectories(); err != nil {
		log.Fatal(err)
	}

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
}
