package commands

import (
	"fmt"
	"sort"

	"github.com/photoprism/photoprism/internal/meta"

	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/pkg/report"
)

// ShowTagsCommand configures the command name, flags, and action.
var ShowTagsCommand = cli.Command{
	Name:    "tags",
	Aliases: []string{"metadata"},
	Usage:   "Shows an overview of the supported metadata tags",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "short, s",
			Usage: "hide links to documentation",
		},
		cli.BoolFlag{
			Name:  "md, m",
			Usage: "render Markdown without line breaks",
		},
	},
	Action: showTagsAction,
}

// showTagsAction reports supported Exif and XMP metadata tags.
func showTagsAction(ctx *cli.Context) error {
	rows, cols := meta.Report(&meta.Data{})

	// Sort rows by type data type and name.
	sort.Slice(rows, func(i, j int) bool {
		if rows[i][1] == rows[j][1] {
			return rows[i][0] < rows[j][0]
		} else {
			return rows[i][1] < rows[j][1]
		}
	})

	// Show table with the supported metadata tags.
	fmt.Println(report.Table(rows, cols, ctx.Bool("md")))

	// Show documentation links for those who want to delve deeper.
	if !ctx.Bool("short") {
		fmt.Printf("## Metadata Tags by Namespace ##\n\n")
		fmt.Println(report.Table(meta.Docs, []string{"Namespace", "Documentation"}, ctx.Bool("md")))
	}

	return nil
}
