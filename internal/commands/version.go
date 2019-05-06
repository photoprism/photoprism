package commands

import (
	"fmt"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/urfave/cli"
)

// Prints current version
var VersionCommand = cli.Command{
	Name:   "version",
	Usage:  "Shows version information",
	Action: versionAction,
}

func versionAction(ctx *cli.Context) error {
	conf := config.NewConfig(ctx)

	fmt.Println(conf.Version())

	return nil
}
