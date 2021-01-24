package commands

import (
	"context"
	"time"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/service"
	"github.com/urfave/cli"
)

// CleanUpCommand registers the cleanup command.
var CleanUpCommand = cli.Command{
	Name:   "cleanup",
	Usage:  "Removes orphaned thumbnails and index entries",
	Flags:  cleanUpFlags,
	Action: cleanUpAction,
}

var cleanUpFlags = []cli.Flag{
	cli.BoolFlag{
		Name:  "dry",
		Usage: "dry run, don't actually remove anything",
	},
}

// cleanUpAction removes orphaned thumbnails and index entries.
func cleanUpAction(ctx *cli.Context) error {
	start := time.Now()

	conf := config.NewConfig(ctx)
	service.SetConfig(conf)

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := conf.Init(); err != nil {
		return err
	}

	conf.InitDb()

	if conf.ReadOnly() {
		log.Infof("cleanup: read-only mode enabled")
	}

	cleanUp := service.CleanUp()

	opt := photoprism.CleanUpOptions{
		Dry: ctx.Bool("dry"),
	}

	if thumbs, orphans, err := cleanUp.Start(opt); err != nil {
		return err
	} else {
		elapsed := time.Since(start)

		log.Infof("cleanup: removed %d orphaned thumbnails and %d photos in %s", thumbs, orphans, elapsed)
	}

	conf.Shutdown()

	return nil
}
