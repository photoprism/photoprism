package commands

import (
	"fmt"

	"github.com/photoprism/photoprism/internal/context"
	"github.com/urfave/cli"
)

// Prints current version
var VersionCommand = cli.Command{
	Name:   "version",
	Usage:  "Shows version information",
	Action: versionAction,
}

func versionAction(ctx *cli.Context) error {
	conf := context.NewConfig(ctx)

	fmt.Println(conf.AppVersion())

	return nil
}
