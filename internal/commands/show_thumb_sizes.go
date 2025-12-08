package commands

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/txt/report"
)

// ShowThumbSizesCommand configures the command name, flags, and action.
var ShowThumbSizesCommand = &cli.Command{
	Name:    "thumb-sizes",
	Aliases: []string{"thumbs"},
	Usage:   "Displays supported thumbnail types and sizes",
	Flags:   report.CliFlags,
	Action:  showThumbSizesAction,
}

// showThumbSizesAction displays supported standard thumbnail sizes.
func showThumbSizesAction(ctx *cli.Context) error {
	rows, cols := thumb.Report(thumb.Sizes.All(), false)
	format, formatErr := report.CliFormatStrict(ctx)
	if formatErr != nil {
		return formatErr
	}
	result, err := report.RenderFormat(rows, cols, format)
	fmt.Println(result)
	return err
}
