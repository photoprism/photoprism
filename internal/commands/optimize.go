package commands

import (
	"context"
	"time"

	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/service"
	"github.com/photoprism/photoprism/internal/workers"
)

// OptimizeCommand registers the index cli command.
var OptimizeCommand = cli.Command{
	Name:   "optimize",
	Usage:  "Performs photo metadata maintenance",
	Action: optimizeAction,
}

// optimizeAction starts metadata check and optimization.
func optimizeAction(ctx *cli.Context) error {
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
		log.Infof("optimize: read-only mode enabled")
	}

	worker := workers.NewMeta(conf)

	if err := worker.Start(time.Second * 15); err != nil {
		return err
	} else {
		elapsed := time.Since(start)

		log.Infof("completed in %s", elapsed)
	}

	conf.Shutdown()

	return nil
}
