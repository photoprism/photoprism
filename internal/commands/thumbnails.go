package commands

import (
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/photoprism"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

// Pre-renders thumbnails
var ThumbnailsCommand = cli.Command{
	Name:  "thumbnails",
	Usage: "Pre-renders thumbnails",
	Flags: []cli.Flag{
		cli.IntSliceFlag{
			Name:  "size, s",
			Usage: "Thumbnail size in pixels",
		},
		cli.BoolFlag{
			Name:  "default, d",
			Usage: "Render default sizes: 720, 1280, 1920, 2560 and 3840px",
		},
		cli.BoolFlag{
			Name:  "square, q",
			Usage: "Square aspect ratio",
		},
	},
	Action: thumbnailsAction,
}

func thumbnailsAction(ctx *cli.Context) error {
	conf := config.NewConfig(ctx)

	if err := conf.CreateDirectories(); err != nil {
		return err
	}

	log.Infof("creating thumbnails in \"%s\"", conf.ThumbnailsPath())

	sizes := ctx.IntSlice("size")

	if ctx.Bool("default") {
		sizes = []int{720, 1280, 1920, 2560, 3840}
	}

	if len(sizes) == 0 {
		log.Warn("no thumbnail size selected")
		return nil
	}

	for _, size := range sizes {
		photoprism.CreateThumbnailsFromOriginals(conf.OriginalsPath(), conf.ThumbnailsPath(), size, ctx.Bool("square"))
	}

	log.Info("thumbnails created")

	return nil
}
