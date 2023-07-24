package commands

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/report"
)

// ShowVideoSizesCommand configures the command name, flags, and action.
var ShowVideoSizesCommand = cli.Command{
	Name:   "video-sizes",
	Usage:  "Displays supported standard video sizes",
	Flags:  report.CliFlags,
	Action: showVideoSizesAction,
}

// showVideoSizesAction displays supported standard video sizes.
func showVideoSizesAction(ctx *cli.Context) error {
	rows, cols := thumb.Report(thumb.VideoSizes, true)

	result, err := report.RenderFormat(rows, cols, report.CliFormat(ctx))

	fmt.Println(result)

	return err
}
