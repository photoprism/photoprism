package report

import (
	"fmt"

	"github.com/photoprism/photoprism/pkg/clean"
)

// RenderFormat returns a text-formatted table, optionally as valid Markdown,
// so the output can be pasted into the docs.
func RenderFormat(rows [][]string, cols []string, format Format) (string, error) {
	switch format {
	case CSV:
		return Render(rows, cols, Options{Format: CSV})
	case TSV:
		return Render(rows, cols, Options{Format: TSV})
	case Markdown:
		return Render(rows, cols, Options{Format: Markdown, Valid: true})
	case Default:
		return Render(rows, cols, Options{Format: Default, Valid: false})
	default:
		return "", fmt.Errorf("invalid format %s", clean.Log(string(format)))
	}
}

// Render returns a text-formatted table, optionally as valid Markdown,
// so the output can be pasted into the docs.
func Render(rows [][]string, cols []string, opt Options) (string, error) {
	switch opt.Format {
	case CSV:
		return CsvExport(rows, cols, ';')
	case TSV:
		return CsvExport(rows, cols, '\t')
	case Markdown:
		opt.Valid = true
		return MarkdownTable(rows, cols, opt), nil
	case Default:
		opt.Valid = false
		return MarkdownTable(rows, cols, opt), nil
	default:
		return "", fmt.Errorf("invalid format %s", clean.Log(string(opt.Format)))
	}
}
