package commands

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media"
	"github.com/photoprism/photoprism/pkg/txt/report"
)

// ShowFileFormatsCommand configures the command name, flags, and action.
var ShowFileFormatsCommand = cli.Command{
	Name:    "file-formats",
	Aliases: []string{"formats"},
	Usage:   "Displays supported media and sidecar file formats",
	Flags: append(report.CliFlags, cli.BoolFlag{
		Name:  "short, s",
		Usage: "hide format descriptions",
	}),
	Action: showFileFormatsAction,
}

// showFileFormatsAction lists supported media and sidecar file formats.
func showFileFormatsAction(ctx *cli.Context) error {
	rows, cols := media.Report(fs.Extensions.Types(true), !ctx.Bool("short"), true, true)

	result, err := report.RenderFormat(rows, cols, report.CliFormat(ctx))

	fmt.Println(result)

	return err
}
