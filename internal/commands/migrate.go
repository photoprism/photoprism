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
	conf := context.NewConfig(ctx)

	log.Infoln("migrating database")

	conf.MigrateDb()

	log.Infoln("database migration complete")

	return nil
}
