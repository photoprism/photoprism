package fs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileFormats_Markdown(t *testing.T) {
	t.Run("All", func(t *testing.T) {
		f := Extensions.Formats(true)
		rows, cols := f.Report(true, true, true)
		assert.NotEmpty(t, rows)
		assert.NotEmpty(t, cols)
		assert.Len(t, cols, 4)
		assert.GreaterOrEqual(t, len(rows), 30)
	})
	t.Run("Compact", func(t *testing.T) {
		f := Extensions.Formats(true)
		rows, cols := f.Report(false, false, false)
		assert.NotEmpty(t, rows)
		assert.NotEmpty(t, cols)
		assert.Len(t, cols, 1)
		assert.GreaterOrEqual(t, len(rows), 30)
	})
}
