package commands

import (
	"fmt"

	"github.com/photoprism/photoprism/internal/context"
	"github.com/urfave/cli"
)

var MigrateCommand = cli.Command{
	Name:   "migrate",
	Usage:  "Automatically migrates / initializes database",
	Action: migrateAction,
}

// Automatically migrates / initializes database; called by MigrateCommand
func migrateAction(ctx *cli.Context) error {
	conf := context.NewConfig(ctx)

	fmt.Println("Migrating database...")

	conf.MigrateDb()

	fmt.Println("Done.")

	return nil
}
