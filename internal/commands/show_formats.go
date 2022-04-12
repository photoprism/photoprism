package commands

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/report"
)

var ShowFormatsCommand = cli.Command{
	Name:  "formats",
	Usage: "Displays supported media and sidecar file formats",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "compact, c",
			Usage: "hide format descriptions to make the output more compact",
		},
		cli.BoolFlag{
			Name:  "no-wrap, n",
			Usage: "disable text-wrapping so the output can be pasted into Markdown files",
		},
	},
	Action: showFormatsAction,
}

// showFormatsAction lists supported media and sidecar file formats.
func showFormatsAction(ctx *cli.Context) error {
	rows, cols := fs.Extensions.Formats(true).Table(!ctx.Bool("compact"), true, true)

	fmt.Println(report.Markdown(rows, cols, !ctx.Bool("no-wrap")))

	return nil
}
