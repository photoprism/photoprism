package report

import (
	"bytes"

	"github.com/olekukonko/tablewriter"
)

// Table returns a text-formatted table, optionally as valid Markdown,
// so the output can be pasted into the docs.
func Table(rows [][]string, cols []string, markDown bool) string {
	buf := &bytes.Buffer{}

	// Configure.
	borders := tablewriter.Border{
		Left:   true,
		Right:  true,
		Top:    !markDown,
		Bottom: !markDown,
	}

	// Render.
	table := tablewriter.NewWriter(buf)
	table.SetAutoWrapText(!markDown)
	table.SetAutoFormatHeaders(false)
	table.SetHeader(cols)
	table.SetBorders(borders)
	table.SetCenterSeparator("|")
	table.AppendBulk(rows)
	table.Render()

	return buf.String()
}
