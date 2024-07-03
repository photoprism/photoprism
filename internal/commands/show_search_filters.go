package commands

import (
	"fmt"
	"sort"

	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/txt/report"
)

// ShowSearchFiltersCommand configures the command name, flags, and action.
var ShowSearchFiltersCommand = cli.Command{
	Name:   "search-filters",
	Usage:  "Displays supported search filters with examples",
	Flags:  report.CliFlags,
	Action: showSearchFiltersAction,
}

// showSearchFiltersAction displays a search filter overview with examples.
func showSearchFiltersAction(ctx *cli.Context) error {
	rows, cols := form.Report(&form.SearchPhotos{})

	sort.Slice(rows, func(i, j int) bool {
		if rows[i][1] == rows[j][1] {
			return rows[i][0] < rows[j][0]
		} else {
			return rows[i][1] < rows[j][1]
		}
	})

	result, err := report.RenderFormat(rows, cols, report.CliFormat(ctx))

	fmt.Println(result)

	return err
}
