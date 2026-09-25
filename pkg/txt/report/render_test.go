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

// TestTableColumnWidth verifies that a column is sized by the display width of
// its cells rather than by their byte or rune count, so that a title holding
// CJK characters, emoji, or combining marks still renders a rectangular table.
// Counting the horizontal rules in the top border measures the column exactly:
// it is the cell width plus the two padding spaces.
func TestTableColumnWidth(t *testing.T) {
	// width renders a one-cell table under a one-character heading and returns
	// the number of rules in its top border.
	width := func(cell string) int {
		t.Helper()
		result, err := RenderFormat([][]string{{cell}}, []string{"T"}, Default)
		if err != nil {
			t.Fatal(err)
		}
		return strings.Count(strings.SplitN(result, "\n", 2)[0], "─")
	}

	t.Run("Ascii", func(t *testing.T) {
		assert.Equal(t, 6, width("abcd"))
		assert.Equal(t, 4, width("ab"))
	})
	t.Run("DoubleWidth", func(t *testing.T) {
		// Two CJK characters occupy four columns, as much as four ASCII ones.
		assert.Equal(t, width("abcd"), width("東京"))
	})
	t.Run("Emoji", func(t *testing.T) {
		assert.Equal(t, width("ab"), width("📷"))
	})
	t.Run("CombiningMarks", func(t *testing.T) {
		// A decomposed accent adds a rune but no column. Written as an escape
		// because the precomposed U+00E9 is East Asian ambiguous, which would
		// tie the expected width to the locale the test runs under.
		assert.Equal(t, width("e"), width("e\u0301"))
	})
	t.Run("Marks", func(t *testing.T) {
		// Hebrew niqqud and Devanagari matras are marks on a base letter, so
		// both of these are four columns wide despite their longer rune counts.
		assert.Equal(t, 6, width("שָׁלוֹם"))
		assert.Equal(t, 6, width("नमस्ते"))
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
