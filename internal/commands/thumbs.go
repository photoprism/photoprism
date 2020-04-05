package commands

import (
	"time"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/service"
	"github.com/urfave/cli"
)

// ThumbsCommand is used to register the thumbs cli command
var ThumbsCommand = cli.Command{
	Name:  "thumbs",
	Usage: "Pre-renders thumbnails to boost performance",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "force, f",
			Usage: "re-create existing thumbnails",
		},
	},
	Action: thumbsAction,
}

// thumbsAction pre-render the thumbnails
func thumbsAction(ctx *cli.Context) error {
	start := time.Now()

	conf := config.NewConfig(ctx)
	service.SetConfig(conf)

	if err := conf.CreateDirectories(); err != nil {
		return err
	}

	log.Infof("creating thumbnails in \"%s\"", conf.ThumbnailsPath())

	rs := service.Resample()

	if err := rs.Start(ctx.Bool("force")); err != nil {
		log.Error(err)
		return err
	}

	elapsed := time.Since(start)

	log.Infof("thumbnails created in %s", elapsed)

	return nil
}
