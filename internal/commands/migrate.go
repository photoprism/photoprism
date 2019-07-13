package commands

import (
	"context"
	"time"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/urfave/cli"
)

// Automatically migrates / initializes database
var MigrateCommand = cli.Command{
	Name:   "migrate",
	Usage:  "Automatically migrates / initializes database",
	Action: migrateAction,
}

func migrateAction(ctx *cli.Context) error {
	start := time.Now()

	conf := config.NewConfig(ctx)
	cctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := conf.Init(cctx); err != nil {
		return err
	}

	log.Infoln("migrating database")

	conf.MigrateDb()

	elapsed := time.Since(start)

	log.Infof("database migration completed in %s", elapsed)

	conf.Shutdown()

	return nil
}
