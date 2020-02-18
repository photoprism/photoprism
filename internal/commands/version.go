package commands

import (
	"fmt"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/urfave/cli"
)

// VersionCommand is used to register the version cli command
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
