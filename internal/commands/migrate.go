package commands

import (
	"github.com/photoprism/photoprism/internal/context"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

// Automatically migrates / initializes database
var MigrateCommand = cli.Command{
	Name:   "migrate",
	Usage:  "Automatically migrates / initializes database",
	Action: migrateAction,
}

func migrateAction(ctx *cli.Context) error {
	app := context.NewContext(ctx)

	log.Infoln("migrating database")

	app.MigrateDb()

	log.Infoln("database migration complete")

	app.Shutdown()

	return nil
}
