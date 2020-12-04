package commands

import (
	"context"
	"time"

	"github.com/photoprism/photoprism/internal/workers"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/service"
	"github.com/urfave/cli"
)

// OptimizeCommand is used to register the index cli command.
var OptimizeCommand = cli.Command{
	Name:   "optimize",
	Usage:  "Starts metadata check and optimization",
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
		log.Infof("read-only mode enabled")
	}

	worker := workers.NewMeta(conf)

	if err := worker.Start(); err != nil {
		return err
	} else {
		elapsed := time.Since(start)

		log.Infof("completed in %s", elapsed)
	}

	conf.Shutdown()

	return nil
}
