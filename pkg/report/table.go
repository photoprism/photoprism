package report

import (
	"bytes"

	"github.com/olekukonko/tablewriter"
)

// Table returns a text-formatted table, optionally as valid Markdown,
// so the output can be pasted into the docs.
func Table(rows [][]string, cols []string, markDown bool) string {
	return TableWithCaption(rows, cols, "", markDown)
}

// TableWithCaption returns a text-formatted table with caption, optionally as valid Markdown,
// so the output can be pasted into the docs.
func TableWithCaption(rows [][]string, cols []string, caption string, markDown bool) string {
	buf := &bytes.Buffer{}

	// Set Borders.
	borders := tablewriter.Border{
		Left:   true,
		Right:  true,
		Top:    !markDown,
		Bottom: !markDown,
	}

	// Render.
	table := tablewriter.NewWriter(buf)

	// Set Caption.
	if caption != "" {
		table.SetCaption(true, caption)
	}

	table.SetAutoWrapText(!markDown)
	table.SetAutoFormatHeaders(false)
	table.SetHeader(cols)
	table.SetBorders(borders)
	table.SetCenterSeparator("|")
	table.AppendBulk(rows)
	table.Render()

	return buf.String()
}
