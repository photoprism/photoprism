package commands

import (
	"fmt"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/urfave/cli"
)

var MigrateCommand = cli.Command{
	Name:   "migrate",
	Usage:  "Automatically migrates / initializes database",
	Action: migrateAction,
}

func migrateAction(context *cli.Context) error {
	conf := photoprism.NewConfig(context)

	fmt.Println("Migrating database...")

	conf.MigrateDb()

	fmt.Println("Done.")

	return nil
}
