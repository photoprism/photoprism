package commands

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media"
	"github.com/photoprism/photoprism/pkg/report"
)

// ShowFormatsCommand configures the command name, flags, and action.
var ShowFormatsCommand = cli.Command{
	Name:  "formats",
	Usage: "Lists supported media and sidecar file formats",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "short, s",
			Usage: "hide format descriptions",
		},
		cli.BoolFlag{
			Name:  "md, m",
			Usage: "render valid Markdown",
		},
	},
	Action: showFormatsAction,
}

// showFormatsAction lists supported media and sidecar file formats.
func showFormatsAction(ctx *cli.Context) error {
	rows, cols := media.Report(fs.Extensions.Types(true), !ctx.Bool("short"), true, true)
	fmt.Println(report.Table(rows, cols, ctx.Bool("md")))

	return nil
}
