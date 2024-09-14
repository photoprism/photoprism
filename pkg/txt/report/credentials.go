package report

import (
	"bytes"

	"github.com/olekukonko/tablewriter"
)

// Credentials returns a text-formatted table with credentials.
func Credentials(idName, idValue, secretName, secretValue string) string {
	buf := &bytes.Buffer{}

	// Set borders.
	borders := tablewriter.Border{
		Left:   true,
		Right:  true,
		Top:    true,
		Bottom: true,
	}

	// Set values.
	rows := make([][]string, 2)
	rows[0] = []string{idName, secretName}
	rows[1] = []string{idValue, secretValue}

	// Render table.
	table := tablewriter.NewWriter(buf)

	table.SetRowLine(true)
	table.SetAutoWrapText(false)
	table.SetHeader(nil)
	table.SetBorders(borders)
	table.SetCenterSeparator("|")
	table.AppendBulk(rows)
	table.Render()

	return buf.String()
}
