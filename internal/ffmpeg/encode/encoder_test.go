package encode

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindEncoder(t *testing.T) {
	t.Run("Software", func(t *testing.T) {
		assert.Equal(t, "libx264", FindEncoder("software").String())
	})
	t.Run("Apple", func(t *testing.T) {
		assert.Equal(t, "h264_videotoolbox", FindEncoder("apple").String())
	})
	t.Run("Unsupported", func(t *testing.T) {
		assert.Equal(t, "libx264", FindEncoder("xxx").String())
	})
}
