package commands

import (
	"context"
	"time"

	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/get"
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
		cli.BoolFlag{
			Name:  "originals, o",
			Usage: "originals only, skip sidecar files",
		},
	},
	Action: thumbsAction,
}

// thumbsAction pre-renders thumbnail images.
func thumbsAction(ctx *cli.Context) error {
	start := time.Now()

	conf, err := InitConfig(ctx)

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err != nil {
		return err
	}

	conf.RegisterDb()
	defer conf.Shutdown()

	log.Infof("creating thumbs in %s", clean.Log(conf.ThumbCachePath()))

	rs := get.Thumbs()

	if err := rs.Start(ctx.Bool("force"), ctx.Bool("originals")); err != nil {
		log.Error(err)
		return err
	}

	log.Infof("thumbs created in %s", time.Since(start))

	return nil
}
