package commands

import (
	"time"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/urfave/cli"
)

// Pre-renders thumbnails
var ThumbnailsCommand = cli.Command{
	Name:  "thumbnails",
	Usage: "Pre-renders thumbnails to boost performance",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "force, f",
			Usage: "re-create existing thumbnails",
		},
	},
	Action: thumbnailsAction,
}

func thumbnailsAction(ctx *cli.Context) error {
	start := time.Now()

	conf := config.NewConfig(ctx)

	if err := conf.CreateDirectories(); err != nil {
		return err
	}

	log.Infof("creating thumbnails in \"%s\"", conf.ThumbnailsPath())

	thumbnails := photoprism.NewThumbnails(conf)

	if err := thumbnails.Start(ctx.Bool("force")); err != nil {
		log.Error(err)
		return err
	}

	elapsed := time.Since(start)

	log.Infof("thumbnails created in %s", elapsed)

	return nil
}
