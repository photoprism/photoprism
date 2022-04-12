package commands

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/report"
)

var ShowConfigCommand = cli.Command{
	Name:  "config",
	Usage: "Displays global configuration values",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "no-wrap, n",
			Usage: "disable text-wrapping",
		},
	},
	Action: showConfigAction,
}

// showConfigAction lists configuration options and their values.
func showConfigAction(ctx *cli.Context) error {
	conf := config.NewConfig(ctx)
	conf.SetLogLevel(logrus.FatalLevel)

	rows, cols := conf.Table()

	fmt.Println(report.Markdown(rows, cols, !ctx.Bool("no-wrap")))

	return nil
}
