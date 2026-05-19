package report

import (
	"bytes"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
)

// escapeMarkdownCell escapes Markdown table-significant runes (`|`, the
// horizontal-rule sequence `* * *`) and angle brackets in a single cell.
// Angle brackets are backslash-escaped so a value that happens to look
// like an HTML tag renders as plain text in both the Markdown source
// and any HTML produced from it (e.g. on docs.photoprism.app). CommonMark
// guarantees that `\<` renders as the literal `<` (entity `&lt;`) in
// HTML, so the escaped form is safe in both pipelines.
func escapeMarkdownCell(cell string) string {
	if strings.ContainsRune(cell, '|') {
		cell = strings.ReplaceAll(cell, "|", "\\|")
	}
	if strings.Contains(cell, "* * *") {
		cell = strings.ReplaceAll(cell, "* * *", "\\* \\* \\*")
	}
	if strings.ContainsRune(cell, '<') {
		cell = strings.ReplaceAll(cell, "<", "\\<")
	}
	if strings.ContainsRune(cell, '>') {
		cell = strings.ReplaceAll(cell, ">", "\\>")
	}
	return cell
}

// MarkdownTable returns a text-formatted table with caption, optionally as valid Markdown,
// so the output can be pasted into the docs.
func MarkdownTable(rows [][]string, cols []string, opt Options) string {
	// Escape Markdown.
	if opt.Valid {
		for i := range cols {
			cols[i] = escapeMarkdownCell(cols[i])
		}
		for i := range rows {
			for j := range rows[i] {
				rows[i][j] = escapeMarkdownCell(rows[i][j])
			}
		}
	}

	result := &bytes.Buffer{}

	var tableRenderer tw.Renderer
	var tableConfig tablewriter.Config

	if opt.Valid {
		tableRenderer = renderer.NewMarkdown()
		tableConfig = tablewriter.Config{
			Header: tw.CellConfig{Alignment: tw.CellAlignment{Global: tw.AlignLeft}, Formatting: tw.CellFormatting{AutoFormat: -1}},
			Row: tw.CellConfig{
				Alignment: tw.CellAlignment{Global: tw.AlignLeft},
			},
		}
	} else {
		tableRenderer = renderer.NewBlueprint()
		tableConfig = tablewriter.Config{
			Header: tw.CellConfig{Alignment: tw.CellAlignment{Global: tw.AlignCenter}, Formatting: tw.CellFormatting{AutoFormat: -1}},
			Row: tw.CellConfig{
				Alignment: tw.CellAlignment{Global: tw.AlignLeft},
			},
		}
	}

	// RenderFormat.
	table := tablewriter.NewTable(result,
		tablewriter.WithRenderer(tableRenderer),
		tablewriter.WithConfig(tableConfig),
	)

	// Set Caption.
	if opt.Caption != "" {
		table.Caption(tw.Caption{Text: opt.Caption})
	}

	table.Header(cols)
	_ = table.Bulk(rows)
	_ = table.Render()

	return result.String()
}
