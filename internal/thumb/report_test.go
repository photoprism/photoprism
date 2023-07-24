package thumb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReport(t *testing.T) {
	t.Run("Videos", func(t *testing.T) {
		rows, cols := Report(VideoSizes, true)
		assert.Equal(t, 2, len(cols))
		assert.Equal(t, len(VideoSizes), len(rows))
	})
	t.Run("Thumbs", func(t *testing.T) {
		rows, cols := Report(Sizes.All(), false)
		assert.Equal(t, 5, len(cols))
		assert.Equal(t, len(Sizes), len(rows))
	})
}
