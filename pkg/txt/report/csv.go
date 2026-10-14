package report

import (
	"bytes"
	"encoding/csv"
)

// CsvExport returns the report as character separated values, with each cell passed through clean.Display.
func CsvExport(rows [][]string, cols []string, sep rune) (string, error) {
	rows, cols = displayRows(rows), displayRow(cols)

	buf := &bytes.Buffer{}
	writer := csv.NewWriter(buf)

	if sep > 0 {
		writer.Comma = sep
	}

	err := writer.Write(cols)

	if err != nil {
		return "", err
	}

	if err := writer.WriteAll(rows); err != nil {
		return "", err
	}

	return buf.String(), nil
}
