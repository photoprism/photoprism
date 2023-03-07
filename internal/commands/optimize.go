package commands

import (
	"context"
	"time"

	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/workers"
)

// OptimizeCommand configures the command name, flags, and action.
var OptimizeCommand = cli.Command{
	Name:  "optimize",
	Usage: "Maintains titles, estimates, and other metadata",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "force, f",
			Usage: "update all, including recently optimized",
		},
	},
	Action: optimizeAction,
}

// optimizeAction updates titles, estimates, and other metadata.
func optimizeAction(ctx *cli.Context) error {
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
		log.Infof("config: enabled read-only mode")
	}

	force := ctx.Bool("force")
	worker := workers.NewMeta(conf)

	delay := 15 * time.Second
	interval := entity.MetadataUpdateInterval

	if force {
		delay = 0
		interval = 0
	}

	if err := worker.Start(delay, interval, force); err != nil {
		return err
	} else {
		elapsed := time.Since(start)

		log.Infof("completed in %s", elapsed)
	}

	return nil
}
