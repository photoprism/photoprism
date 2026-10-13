package report

import (
	"bytes"

	"github.com/olekukonko/tablewriter"
)

// Credentials returns a text-formatted table with credentials, each value passed through clean.Display.
func Credentials(idName, idValue, secretName, secretValue string) string {
	result := &bytes.Buffer{}
	table := tablewriter.NewWriter(result)

	// Set values.
	rows := make([][]string, 2)
	rows[0] = displayRow([]string{idName, secretName})
	rows[1] = displayRow([]string{idValue, secretValue})

	_ = table.Bulk(rows)
	_ = table.Render()

	return result.String()
}
