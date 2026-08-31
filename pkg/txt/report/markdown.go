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

// rowAlignment returns the per-column alignment for the data rows, or nil where none was asked for.
//
// Padded to the column count, since the table writer indexes it positionally and a short slice would
// otherwise leave the trailing columns unset rather than defaulted.
func rowAlignment(cols []string, align []Align) []tw.Align {
	if len(align) == 0 {
		return nil
	}

	out := make([]tw.Align, len(cols))

	for i := range out {
		out[i] = tw.AlignLeft
	}

	for i, a := range align {
		if i >= len(out) {
			break
		}

		switch a {
		case AlignRight:
			out[i] = tw.AlignRight
		case AlignCenter:
			out[i] = tw.AlignCenter
		case AlignLeft, AlignDefault:
			out[i] = tw.AlignLeft
		}
	}

	return out
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

	rowAlign := rowAlignment(cols, opt.Align)

	if opt.Valid {
		// Set on both because the renderer resolves alignment once, from the header, and Markdown
		// carries it in the delimiter row for the whole column rather than for the body alone.
		tableRenderer = renderer.NewMarkdown()
		tableConfig = tablewriter.Config{
			Header: tw.CellConfig{
				Alignment:  tw.CellAlignment{Global: tw.AlignLeft, PerColumn: rowAlign},
				Formatting: tw.CellFormatting{AutoFormat: -1},
			},
			Row: tw.CellConfig{
				Alignment: tw.CellAlignment{Global: tw.AlignLeft, PerColumn: rowAlign},
			},
		}
	} else {
		tableRenderer = renderer.NewBlueprint()
		tableConfig = tablewriter.Config{
			Header: tw.CellConfig{Alignment: tw.CellAlignment{Global: tw.AlignCenter}, Formatting: tw.CellFormatting{AutoFormat: -1}},
			Row: tw.CellConfig{
				Alignment: tw.CellAlignment{Global: tw.AlignLeft, PerColumn: rowAlign},
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
