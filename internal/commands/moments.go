package commands

import (
	"context"
	"time"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/service"
	"github.com/urfave/cli"
)

// MomentsCommand is used to register the index cli command
var MomentsCommand = cli.Command{
	Name:   "moments",
	Usage:  "Creates albums based on popular locations, dates and labels",
	Action: momentsAction,
}

// momentsAction creates albums based on popular locations, dates and labels.
func momentsAction(ctx *cli.Context) error {
	start := time.Now()

	conf := config.NewConfig(ctx)
	service.SetConfig(conf)

	cctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := conf.Init(cctx); err != nil {
		return err
	}

	conf.InitDb()

	if conf.ReadOnly() {
		log.Infof("read-only mode enabled")
	}

	moments := service.Moments()

	if err := moments.Start(); err != nil {
		return err
	} else {
		elapsed := time.Since(start)

		log.Infof("completed in %s", elapsed)
	}

	conf.Shutdown()

	return nil
}
