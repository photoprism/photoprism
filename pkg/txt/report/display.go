package report

import "github.com/photoprism/photoprism/pkg/clean"

// displayRow returns a copy of a row with each cell passed through clean.Display.
func displayRow(row []string) []string {
	out := make([]string, len(row))

	for i, cell := range row {
		out[i] = clean.Display(cell)
	}

	return out
}

// displayRows returns a copy of the rows with each cell passed through clean.Display.
func displayRows(rows [][]string) [][]string {
	out := make([][]string, len(rows))

	for i, row := range rows {
		out[i] = displayRow(row)
	}

	return out
}
