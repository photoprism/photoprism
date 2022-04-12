package commands

import (
	"fmt"

	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/urfave/cli"
)

var ShowFormatsCommand = cli.Command{
	Name:   "formats",
	Usage:  "Displays supported media and sidecar file formats",
	Action: showFormatsAction,
}

// showFormatsAction lists supported media and sidecar file formats.
func showFormatsAction(ctx *cli.Context) error {
	formats := fs.Extensions.Formats(true).Markdown()
	fmt.Println(formats)
	return nil
}
