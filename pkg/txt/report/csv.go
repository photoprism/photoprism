package report

import (
	"bytes"
	"encoding/csv"
)

// CsvExport returns the report as character separated values.
func CsvExport(rows [][]string, cols []string, sep rune) (string, error) {
	buf := &bytes.Buffer{}
	writer := csv.NewWriter(buf)

	if sep > 0 {
		writer.Comma = sep
	}

	err := writer.Write(cols)

	if err != nil {
		return "", err
	}

	err = writer.WriteAll(rows)

	return buf.String(), nil
}
