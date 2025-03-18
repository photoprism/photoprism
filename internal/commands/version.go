package commands

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/config"
)

// VersionCommand configures the "photoprism version" command.
var VersionCommand = &cli.Command{
	Name:   "version",
	Usage:  "Shows version information",
	Action: versionAction,
}

// versionAction displays information about the current version.
func versionAction(ctx *cli.Context) error {
	conf := config.NewConfig(ctx)

	fmt.Println(conf.Version())

	return nil
}
