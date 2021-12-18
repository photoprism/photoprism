package commands

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/config"
)

// MigrateCommand registers the "migrate" CLI command.
var MigrateCommand = cli.Command{
	Name:  "migrate",
	Usage: "Updates the index database schema",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "failed, f",
			Usage: "run previously failed migrations",
		},
		cli.BoolFlag{
			Name:  "trace, t",
			Usage: "show trace logs for debugging",
		},
	},
	Action: migrateAction,
}

// migrateAction initializes and migrates the database.
func migrateAction(ctx *cli.Context) error {
	start := time.Now()

	conf := config.NewConfig(ctx)

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := conf.Init(); err != nil {
		return err
	}

	defer conf.Shutdown()

	if ctx.Bool("trace") {
		log.SetLevel(logrus.TraceLevel)
		log.Infoln("migrate: enabled trace mode")
	}

	runFailed := ctx.Bool("failed")

	if runFailed {
		log.Infoln("migrate: running previously failed migrations")
	}

	log.Infoln("migrating database schema...")

	conf.MigrateDb(runFailed)

	elapsed := time.Since(start)

	log.Infof("migration completed in %s", elapsed)

	return nil
}
