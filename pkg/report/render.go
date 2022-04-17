package report

import (
	"fmt"

	"github.com/photoprism/photoprism/pkg/clean"
)

// Render returns a text-formatted table, optionally as valid Markdown,
// so the output can be pasted into the docs.
func Render(rows [][]string, cols []string, format Format) (string, error) {
	switch format {
	case CSV:
		return CsvExport(rows, cols, ';')
	case TSV:
		return CsvExport(rows, cols, '\t')
	case Markdown:
		return MarkdownTable(rows, cols, "", true), nil
	case Default:
		return MarkdownTable(rows, cols, "", false), nil
	default:
		return "", fmt.Errorf("invalid format %s", clean.Log(string(format)))
	}
}
