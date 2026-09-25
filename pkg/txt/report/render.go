package report

import (
	"fmt"

	"github.com/photoprism/photoprism/pkg/clean"
)

// RenderFormat returns a text-formatted table, optionally as valid Markdown,
// so the output can be pasted into the docs.
func RenderFormat(rows [][]string, cols []string, format Format) (string, error) {
	return RenderFormatOptions(rows, cols, format, Options{})
}

// RenderFormatOptions is RenderFormat with render options the caller supplies, so a report can ask
// for per-column alignment without restating which formats are valid Markdown.
func RenderFormatOptions(rows [][]string, cols []string, format Format, opt Options) (string, error) {
	opt.Format = format

	switch format {
	case JSON:
		return JSONExportKeys(rows, cols, opt.Keys)
	case CSV, TSV:
		return Render(rows, cols, opt)
	case Markdown:
		opt.Valid = true
		return Render(rows, cols, opt)
	case Default:
		opt.Valid = false
		return Render(rows, cols, opt)
	default:
		return "", fmt.Errorf("invalid format %s", clean.Log(string(format)))
	}
}

// Render returns a text-formatted table, optionally as valid Markdown,
// so the output can be pasted into the docs.
func Render(rows [][]string, cols []string, opt Options) (string, error) {
	switch opt.Format {
	case JSON:
		return JSONExport(rows, cols)
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
