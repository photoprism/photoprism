package report

import (
	"testing"

	"github.com/olekukonko/tablewriter/tw"
	"github.com/stretchr/testify/assert"
)

func TestRowAlignment(t *testing.T) {
	cols := []string{"A", "B", "C"}

	t.Run("NoneRequested", func(t *testing.T) {
		assert.Nil(t, rowAlignment(cols, nil))
	})
	t.Run("PaddedToTheColumns", func(t *testing.T) {
		// Shorter than the columns is allowed, and the rest have to default rather than stay unset,
		// since the table writer indexes this positionally.
		got := rowAlignment(cols, []Align{AlignRight})
		assert.Equal(t, []tw.Align{tw.AlignRight, tw.AlignLeft, tw.AlignLeft}, got)
	})
	t.Run("EachAlignment", func(t *testing.T) {
		got := rowAlignment(cols, []Align{AlignRight, AlignCenter, AlignDefault})
		assert.Equal(t, []tw.Align{tw.AlignRight, tw.AlignCenter, tw.AlignLeft}, got)
	})
}

func TestRenderFormatOptionsAlignMarkdown(t *testing.T) {
	rows := [][]string{{"a", "1"}, {"b", "1000"}}
	cols := []string{"Name", "Count"}

	out, err := RenderFormatOptions(rows, cols, Markdown, Options{Align: []Align{AlignDefault, AlignRight}})
	assert.NoError(t, err)

	// Markdown carries alignment in the delimiter row, and the renderer resolves it from the header
	// rather than the body - so setting it on the rows alone reaches the output nowhere.
	assert.Contains(t, out, "---:", "a right-aligned column has to mark the delimiter row")
	assert.NotContains(t, out, ":----------:", "nothing here asked to be centered")
}

func TestRenderFormatOptionsAlign(t *testing.T) {
	rows := [][]string{{"a", "1"}, {"b", "1000"}}
	cols := []string{"Name", "Count"}

	out, err := RenderFormatOptions(rows, cols, Default, Options{Align: []Align{AlignDefault, AlignRight}})
	assert.NoError(t, err)

	// The short value is padded on the left, which is what makes a column of numbers comparable.
	assert.Contains(t, out, "    1 ")
	assert.Contains(t, out, " 1000 ")
}
