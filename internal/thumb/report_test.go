package thumb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReport(t *testing.T) {
	t.Run("Videos", func(t *testing.T) {
		rows, cols := Report(VideoSizes, true)
		assert.GreaterOrEqual(t, 2, len(cols))
		assert.Equal(t, len(VideoSizes), len(rows))
	})
	t.Run("Thumbs", func(t *testing.T) {
		rows, cols := Report(Sizes.All(), false)
		assert.GreaterOrEqual(t, 6, len(cols))
		assert.Equal(t, len(Sizes), len(rows))
	})
}
