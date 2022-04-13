package commands

import (
	"fmt"
	"sort"

	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/report"
)

var ShowFiltersCommand = cli.Command{
	Name:  "filters",
	Usage: "Displays a search filter overview with examples",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "no-wrap, n",
			Usage: "disable text-wrapping so the output can be pasted into Markdown files",
		},
	},
	Action: showFiltersAction,
}

// showFiltersAction lists supported search filters.
func showFiltersAction(ctx *cli.Context) error {
	rows, cols := form.Table(&form.SearchPhotos{})

	sort.Slice(rows, func(i, j int) bool {
		if rows[i][1] == rows[j][1] {
			return rows[i][0] < rows[j][0]
		} else {
			return rows[i][1] < rows[j][1]
		}
	})

	fmt.Println(report.Markdown(rows, cols, !ctx.Bool("no-wrap")))

	return nil
}
