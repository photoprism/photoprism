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
	Name:  "config",
	Usage: "Shows global config option names and values",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "md, m",
			Usage: "renders valid Markdown",
		},
	},
	Action: showConfigAction,
}

// showConfigAction shows global config option names and values.
func showConfigAction(ctx *cli.Context) error {
	conf := config.NewConfig(ctx)
	conf.SetLogLevel(logrus.FatalLevel)

	rows, cols := conf.Report()

	fmt.Println(report.Table(rows, cols, ctx.Bool("md")))

	return nil
}
