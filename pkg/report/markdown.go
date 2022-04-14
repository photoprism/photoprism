package report

import (
	"bytes"

	"github.com/olekukonko/tablewriter"
)

// Markdown returns markdown formatted table.
func Markdown(rows [][]string, cols []string, autoWrap bool) string {
	buf := &bytes.Buffer{}

	table := tablewriter.NewWriter(buf)

	table.SetAutoWrapText(autoWrap)
	table.SetAutoFormatHeaders(false)
	table.SetHeader(cols)
	table.SetBorders(tablewriter.Border{Left: true, Top: true, Right: true, Bottom: true})
	table.SetCenterSeparator("|")
	table.AppendBulk(rows)
	table.Render()

	return buf.String()
}
