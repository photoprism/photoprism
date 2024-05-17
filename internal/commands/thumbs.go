package commands

import (
	"context"
	"strings"
	"time"

	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/get"
	"github.com/photoprism/photoprism/pkg/clean"
)

// ThumbsCommand configures the command name, flags, and action.
var ThumbsCommand = cli.Command{
	Name:      "thumbs",
	Usage:     "(Re-)generates thumbnails based on the current configuration",
	ArgsUsage: "[subfolder]",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "force, f",
			Usage: "replace existing thumbnail images",
		},
		cli.BoolFlag{
			Name:  "originals, o",
			Usage: "scan originals only, skip sidecar folder",
		},
	},
	Action: thumbsAction,
}

// thumbsAction regenerates thumbnails based on the current configuration.
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

	dir := strings.TrimSpace(ctx.Args().First())
	force := ctx.Bool("force")
	originals := ctx.Bool("originals")

	var action, done string
	if force {
		action = "replacing"
		done = "replaced"
	} else {
		action = "generating"
		done = "generated"
	}

	// Display info.
	if dir == "" {
		if originals {
			log.Infof("%s thumbnails for originals only", action)
		} else {
			log.Infof("%s thumbnails for originals and sidecar files", action)
		}
	} else {
		if originals {
			log.Infof("%s thumbnails for originals in %s", action, clean.LogQuote(dir))
		} else {
			log.Infof("%s thumbnails for originals and sidecar files in %s", action, clean.LogQuote(dir))
		}
	}

	w := get.Thumbs()

	if err = w.Start(dir, ctx.Bool("force"), ctx.Bool("originals")); err != nil {
		return err
	}

	log.Infof("thumbnails %s in %s", done, time.Since(start))

	return nil
}
