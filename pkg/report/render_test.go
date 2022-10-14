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
		assert.Contains(t, result, "| bar  | b & a | z                      |")
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
