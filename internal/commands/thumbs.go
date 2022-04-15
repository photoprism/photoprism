package commands

import (
	"time"

	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/service"
	"github.com/photoprism/photoprism/pkg/clean"
)

// ThumbsCommand registers the resample cli command.
var ThumbsCommand = cli.Command{
	Name:  "thumbs",
	Usage: "Generates thumbnails using the current settings",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "force, f",
			Usage: "replace existing thumbnails",
		},
	},
	Action: thumbsAction,
}

// thumbsAction pre-renders thumbnail images.
func thumbsAction(ctx *cli.Context) error {
	start := time.Now()

	conf := config.NewConfig(ctx)
	service.SetConfig(conf)

	if err := conf.Init(); err != nil {
		return err
	}

	log.Infof("creating thumbnails in %s", clean.Log(conf.ThumbPath()))

	rs := service.Resample()

	if err := rs.Start(ctx.Bool("force")); err != nil {
		log.Error(err)
		return err
	}

	log.Infof("thumbnails created in %s", time.Since(start))

	return nil
}
