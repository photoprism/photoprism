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
	Name:    "formats",
	Aliases: []string{"filetypes"},
	Usage:   "Displays supported media and sidecar file formats",
	Flags: append(report.CliFlags, cli.BoolFlag{
		Name:  "short, s",
		Usage: "hide format descriptions",
	}),
	Action: showFormatsAction,
}

// showFormatsAction lists supported media and sidecar file formats.
func showFormatsAction(ctx *cli.Context) error {
	rows, cols := media.Report(fs.Extensions.Types(true), !ctx.Bool("short"), true, true)

	result, err := report.Render(rows, cols, report.CliFormat(ctx))

	fmt.Println(result)

	return err
}
