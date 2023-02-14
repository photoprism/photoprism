package commands

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/config"
)

// VersionCommand configures the command name, flags, and action.
var VersionCommand = cli.Command{
	Name:   "version",
	Usage:  "Shows version information",
	Action: versionAction,
}

// versionAction prints the current version
func versionAction(ctx *cli.Context) error {
	conf := config.NewConfig(ctx)

	fmt.Println(conf.Version())

	return nil
}
