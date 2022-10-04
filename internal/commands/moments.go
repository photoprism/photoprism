package commands

import (
	"context"
	"time"

	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/service"
)

// MomentsCommand registers the moments command.
var MomentsCommand = cli.Command{
	Name:   "moments",
	Usage:  "Creates albums of special moments, trips, and places",
	Action: momentsAction,
}

// momentsAction creates albums of special moments, trips, and places.
func momentsAction(ctx *cli.Context) error {
	start := time.Now()

	conf, err := InitConfig(ctx)

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err != nil {
		return err
	}

	conf.InitDb()
	defer conf.Shutdown()

	if conf.ReadOnly() {
		log.Infof("config: read-only mode enabled")
	}

	w := service.Moments()

	if err := w.Start(); err != nil {
		return err
	} else {
		elapsed := time.Since(start)

		log.Infof("completed in %s", elapsed)
	}

	return nil
}
