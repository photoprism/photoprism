package ffmpeg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindEncoder(t *testing.T) {
	t.Run("software", func(t *testing.T) {
		assert.Equal(t, "libx264", FindEncoder("software").String())
	})
	t.Run("apple", func(t *testing.T) {
		assert.Equal(t, "h264_videotoolbox", FindEncoder("apple").String())
	})
	t.Run("unsupported", func(t *testing.T) {
		assert.Equal(t, "libx264", FindEncoder("xxx").String())
	})
}
