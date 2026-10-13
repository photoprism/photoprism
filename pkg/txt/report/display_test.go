package report

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// hasControl reports whether s contains a C0 control other than the ones a format writes itself, or DEL.
func hasControl(s string, allowed string) bool {
	for _, r := range s {
		if (r < 0x20 || r == 0x7F) && !strings.ContainsRune(allowed, r) {
			return true
		}
	}

	return false
}

// formats maps a subtest name to each report format, since Default is the empty string.
var formats = map[string]Format{"Default": Default, "Markdown": Markdown, "CSV": CSV, "TSV": TSV, "JSON": JSON}

func TestDisplayRow(t *testing.T) {
	row := []string{"a\x00b", "c\nd", "日本"}
	assert.Equal(t, []string{"ab", "c d", "日本"}, displayRow(row))
	assert.Equal(t, "a\x00b", row[0])
}

func TestDisplayRows(t *testing.T) {
	t.Run("Copy", func(t *testing.T) {
		rows := [][]string{{"a\x1b[31mb", "c|d"}}
		out := displayRows(rows)
		assert.Equal(t, [][]string{{"a[31mb", "c|d"}}, out)
		assert.Equal(t, "a\x1b[31mb", rows[0][0])
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Empty(t, displayRows(nil))
		assert.Empty(t, displayRow(nil))
	})
}

func TestRenderControlCharacters(t *testing.T) {
	cols := []string{"Name\x1b", "Value"}
	unsafe := "Ab\x00c\x01d\te\x1bf\x07g\rh\x7f\nx\u202Ey"

	for name, format := range formats {
		t.Run(name, func(t *testing.T) {
			rows := [][]string{{unsafe, "a|b<c>"}}
			result, err := RenderFormat(rows, cols, format)
			assert.NoError(t, err)
			assert.False(t, hasControl(result, "\n\t"), "%q", result)
			assert.NotContains(t, result, "\u202E")
			assert.NotContains(t, result, "\x7f")
			assert.Equal(t, unsafe, rows[0][0])
			assert.Equal(t, "a|b<c>", rows[0][1])
			assert.Equal(t, "Name\x1b", cols[0])

			if format == JSON {
				var decoded []map[string]string
				require.NoError(t, json.Unmarshal([]byte(result), &decoded))
				require.Len(t, decoded, 1)
				assert.Equal(t, unsafe, decoded[0]["name"])
			} else {
				assert.Contains(t, result, "Abcd ef")
				assert.Contains(t, result, "h x?y")
			}
		})
	}
	t.Run("Caption", func(t *testing.T) {
		result := MarkdownTable([][]string{{"a", "b"}}, []string{"A", "B"}, Options{Caption: "Cap\x1btion\x00"})
		assert.False(t, hasControl(result, "\n"), "%q", result)
		assert.Contains(t, result, "Capti")
	})
}

func TestRenderUnchanged(t *testing.T) {
	cols := []string{"Name", "Info"}
	rows := [][]string{
		{"صورة.heic", "日本語 δοκιμή"},
		{"👩\u200D👩\u200D👧", "*optional*\u2002value 100\u202F%"},
		{"plain", "a, b; c"},
	}

	for name, format := range formats {
		t.Run(name, func(t *testing.T) {
			result, err := RenderFormat(rows, cols, format)
			assert.NoError(t, err)

			for _, row := range rows {
				for _, cell := range row {
					assert.Contains(t, result, cell)
				}
			}
		})
	}
}

func TestRenderUnchangedExact(t *testing.T) {
	cols := []string{"Name", "Info"}
	rows := [][]string{{"صورة.heic", "*optional*\u2002value 100\u202F%"}, {"👩\u200D👩\u200D👧", "a, b; c"}}

	t.Run("CSV", func(t *testing.T) {
		result, err := RenderFormat(rows, cols, CSV)
		assert.NoError(t, err)
		assert.Equal(t, "Name;Info\nصورة.heic;*optional*\u2002value 100\u202F%\n👩\u200D👩\u200D👧;\"a, b; c\"\n", result)
	})
	t.Run("JSON", func(t *testing.T) {
		result, err := RenderFormat(rows, cols, JSON)
		assert.NoError(t, err)
		assert.Equal(t, `[{"info":"*optional*`+"\u2002"+`value 100`+"\u202F"+`%","name":"صورة.heic"},{"info":"a, b; c","name":"👩`+"\u200D👩\u200D"+`👧"}]`, result)
	})
}
