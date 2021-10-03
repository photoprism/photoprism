package commands

import (
	"time"

	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/service"
	"github.com/photoprism/photoprism/pkg/txt"
)

// ResampleCommand registers the resample cli command.
var ResampleCommand = cli.Command{
	Name:    "resample",
	Aliases: []string{"thumbs"},
	Usage:   "Pre-caches thumbnail images for improved performance",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "force, f",
			Usage: "replace existing thumbnails",
		},
	},
	Action: resampleAction,
}

// resampleAction pre-caches default thumbnails.
func resampleAction(ctx *cli.Context) error {
	start := time.Now()

	conf := config.NewConfig(ctx)
	service.SetConfig(conf)

	if err := conf.Init(); err != nil {
		return err
	}

	log.Infof("creating thumbnails in %s", txt.Quote(conf.ThumbPath()))

	rs := service.Resample()

	if err := rs.Start(ctx.Bool("force")); err != nil {
		log.Error(err)
		return err
	}

	log.Infof("thumbnails created in %s", time.Since(start))

	return nil
}
