package report

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTable(t *testing.T) {
	cols := []string{"Col1", "Col2"}
	rows := [][]string{
		{"foo", "bar" + strings.Repeat(", abc", 30)},
		{"bar", "b & a | z"}}

	t.Run("DefaultTable", func(t *testing.T) {
		result, err := RenderFormat(rows, cols, Default)
		if err != nil {
			t.Fatal(err)
		}
		// fmt.Println(result)
		assert.Contains(t, result, "│ bar  │ b & a | z")
	})
	t.Run("MarkdownTable", func(t *testing.T) {
		result, err := RenderFormat(rows, cols, Markdown)
		if err != nil {
			t.Fatal(err)
		}
		// fmt.Println(result)
		assert.Contains(t, result, "| bar  | b & a \\| z")
	})
	t.Run("CsvExport", func(t *testing.T) {
		result, err := RenderFormat(rows, cols, CSV)
		if err != nil {
			t.Fatal(err)
		}

		expected := "Col1;Col2\nfoo;bar, abc, abc, abc, abc, abc, abc," +
			" abc, abc, abc, abc, abc, abc, abc, abc, abc," +
			" abc, abc, abc, abc, abc, abc, abc, abc, abc," +
			" abc, abc, abc, abc, abc, abc\nbar;b & a \\| z\n"

		assert.Equal(t, expected, result)
	})
	t.Run("TsvExport", func(t *testing.T) {
		result, err := RenderFormat(rows, cols, TSV)
		if err != nil {
			t.Fatal(err)
		}

		assert.Contains(t, result, "Col1\tCol2\nfoo\tbar, abc, abc")
	})
	t.Run("Invalid", func(t *testing.T) {
		_, err := RenderFormat(rows, cols, Format("invalid"))

		if err == nil {
			t.Fatal("error expected")
		}
	})
}

// TestMarkdownTableEscapesAngleBrackets verifies that the Markdown
// renderer backslash-escapes '<' and '>' in row and header cells so a
// flag default, env-var name, or description that happens to contain
// HTML-looking text cannot be interpreted as raw inline HTML by any
// CommonMark renderer downstream (e.g. docs.photoprism.app).
func TestMarkdownTableEscapesAngleBrackets(t *testing.T) {
	cols := []string{"Default", "Description <Notes>"}
	rows := [][]string{
		{"<auto>", "fallback <script>alert(1)</script>"},
		{"plain", "no markup"},
	}

	result, err := RenderFormat(rows, cols, Markdown)
	if err != nil {
		t.Fatal(err)
	}

	assert.Contains(t, result, "\\<auto\\>", "row angle brackets must be escaped")
	assert.Contains(t, result, "\\<script\\>alert(1)\\</script\\>", "inline-HTML payloads must be escaped")
	assert.Contains(t, result, "Description \\<Notes\\>", "header angle brackets must be escaped")
	assert.NotContains(t, result, "<auto>", "raw '<auto>' must not survive in the Markdown output")
	assert.NotContains(t, result, "<script>", "raw '<script>' must not survive in the Markdown output")
}
