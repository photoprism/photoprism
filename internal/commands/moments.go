package commands

import (
	"context"
	"time"

	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/service"
)

// MomentsCommand registers the moments cli command.
var MomentsCommand = cli.Command{
	Name:   "moments",
	Usage:  "Creates albums based on popular locations, dates, and labels",
	Action: momentsAction,
}

// momentsAction creates albums based on popular locations, dates and labels.
func momentsAction(ctx *cli.Context) error {
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
		log.Infof("moments: read-only mode enabled")
	}

	w := service.Moments()

	if err := w.Start(); err != nil {
		return err
	} else {
		elapsed := time.Since(start)

		log.Infof("completed in %s", elapsed)
	}

	conf.Shutdown()

	return nil
}
