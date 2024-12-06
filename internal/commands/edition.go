package commands

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/config"
)

// EditionCommand configures the "photoprism edition" command.
var EditionCommand = &cli.Command{
	Name:   "edition",
	Usage:  "Shows edition information",
	Hidden: true,
	Action: editionAction,
}

// editionAction displays information about the current edition.
func editionAction(ctx *cli.Context) error {
	conf := config.NewConfig(ctx)

	fmt.Println(conf.Edition())

	return nil
}
