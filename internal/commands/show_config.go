package commands

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/report"
)

// ShowConfigCommand configures the command name, flags, and action.
var ShowConfigCommand = cli.Command{
	Name:   "config",
	Usage:  "Displays global config options and their current values",
	Flags:  report.CliFlags,
	Action: showConfigAction,
}

// showConfigAction shows global config option names and values.
func showConfigAction(ctx *cli.Context) error {
	conf := config.NewConfig(ctx)
	conf.SetLogLevel(logrus.FatalLevel)

	rows, cols := conf.Report()

	result, err := report.Render(rows, cols, report.CliFormat(ctx))

	fmt.Println(result)

	return err
}
