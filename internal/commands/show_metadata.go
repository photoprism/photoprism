package commands

import (
	"fmt"
	"sort"

	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/meta"
	"github.com/photoprism/photoprism/pkg/txt/report"
)

// ShowMetadataCommand configures the command name, flags, and action.
var ShowMetadataCommand = cli.Command{
	Name:    "metadata",
	Aliases: []string{"meta"},
	Usage:   "Displays supported metadata tags and standards",
	Flags: append(report.CliFlags, cli.BoolFlag{
		Name:  "short, s",
		Usage: "hide links to documentation",
	}),
	Action: showMetadataAction,
}

// showMetadataAction reports supported Exif and XMP metadata tags.
func showMetadataAction(ctx *cli.Context) error {
	rows, cols := meta.Report(&meta.Data{})

	// Sort rows by type data type and name.
	sort.Slice(rows, func(i, j int) bool {
		if rows[i][1] == rows[j][1] {
			return rows[i][0] < rows[j][0]
		} else {
			return rows[i][1] < rows[j][1]
		}
	})

	// Output overview of supported metadata tags.
	format := report.CliFormat(ctx)
	result, err := report.RenderFormat(rows, cols, format)

	fmt.Println(result)

	if err != nil || ctx.Bool("short") || format == report.TSV {
		return err
	}

	// Documentation links for those who want to delve deeper.
	result, err = report.RenderFormat(meta.Docs, []string{"Namespace", "Documentation"}, format)

	fmt.Printf("## Metadata Tags by Namespace ##\n\n")
	fmt.Println(result)

	return err
}
