package encode

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOptions_VideoFilter(t *testing.T) {
	opt := &Options{
		Bin:         "",
		Encoder:     "intel",
		DestSize:    1500,
		DestBitrate: "50M",
		MapVideo:    "",
		MapAudio:    "",
	}

	t.Run("Empty", func(t *testing.T) {
		r := opt.VideoFilter("")
		assert.NotContains(t, r, "format")
		assert.Contains(t, r, "min(1500, iw)")
	})
	t.Run("Rgb32", func(t *testing.T) {
		r := opt.VideoFilter("rgb32")
		assert.Contains(t, r, "format=rgb32")
		assert.Contains(t, r, "min(1500, iw)")
	})
}
