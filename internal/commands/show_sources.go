package commands

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/txt/report"
)

// ShowSourcesCommand configures the command name, flags, and action.
var ShowSourcesCommand = &cli.Command{
	Name:   "sources",
	Usage:  "Displays supported metadata sources and their priorities",
	Flags:  append(report.CliFlags),
	Action: showSourcesAction,
}

// showSourcesAction displays supported metadata sources.
func showSourcesAction(ctx *cli.Context) error {
	rows, cols := entity.SrcPriority.Report()
	format, formatErr := report.CliFormatStrict(ctx)
	if formatErr != nil {
		return formatErr
	}
	result, err := report.RenderFormat(rows, cols, format)
	fmt.Println(result)
	return err
}
