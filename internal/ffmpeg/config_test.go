package ffmpeg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOptions_VideoFilter(t *testing.T) {
	Options := &Options{
		Bin:      "",
		Encoder:  "intel",
		Size:     1500,
		Bitrate:  "50",
		MapVideo: "",
		MapAudio: "",
	}

	t.Run("rgb32", func(t *testing.T) {
		r := Options.VideoFilter("rgb32")
		assert.Contains(t, r, "format=rgb32")
		assert.Contains(t, r, "min(1500, iw)")
	})
	t.Run("empty format", func(t *testing.T) {
		r := Options.VideoFilter("")
		assert.NotContains(t, r, "format")
		assert.Contains(t, r, "min(1500, iw)")
	})
}
