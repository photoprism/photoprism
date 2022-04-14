package commands

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/report"
)

var ShowFormatsCommand = cli.Command{
	Name:  "formats",
	Usage: "Lists supported media and sidecar file formats",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "compact, c",
			Usage: "hide format descriptions to make the output more compact",
		},
		cli.BoolFlag{
			Name:  "md, m",
			Usage: "renders valid Markdown",
		},
	},
	Action: showFormatsAction,
}

// showFormatsAction lists supported media and sidecar file formats.
func showFormatsAction(ctx *cli.Context) error {
	rows, cols := fs.Extensions.Formats(true).Report(!ctx.Bool("compact"), true, true)

	fmt.Println(report.Table(rows, cols, ctx.Bool("md")))

	return nil
}
