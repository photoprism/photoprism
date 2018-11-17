package commands

import (
	"fmt"

	"github.com/photoprism/photoprism/internal/context"
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

	fmt.Println("Migrating database...")

	conf.MigrateDb()

	fmt.Println("Done.")

	return nil
}
