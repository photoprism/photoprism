package commands

import (
	"time"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/service"
	"github.com/photoprism/photoprism/pkg/txt"
	"github.com/urfave/cli"
)

// ResampleCommand is used to register the thumbs cli command
var ResampleCommand = cli.Command{
	Name:    "resample",
	Aliases: []string{"thumbs"},
	Usage:   "Pre-renders thumbnails (significantly reduces memory and cpu usage)",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "force, f",
			Usage: "re-create existing thumbnails",
		},
	},
	Action: resampleAction,
}

// resampleAction pre-render the thumbnails
func resampleAction(ctx *cli.Context) error {
	start := time.Now()

	conf := config.NewConfig(ctx)
	service.SetConfig(conf)

	if err := conf.CreateDirectories(); err != nil {
		return err
	}

	log.Infof("creating thumbnails in %s", txt.Quote(conf.ThumbPath()))

	rs := service.Resample()

	if err := rs.Start(ctx.Bool("force")); err != nil {
		log.Error(err)
		return err
	}

	elapsed := time.Since(start)

	log.Infof("thumbnails created in %s", elapsed)

	return nil
}
