package encode

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewOptions(t *testing.T) {
	t.Run("Defaults", func(t *testing.T) {
		opt := NewVideoOptions("", "", 0, "", "", "")
		assert.Equal(t, "ffmpeg", opt.Bin)
		assert.Equal(t, FFmpegBin, opt.Bin)
		assert.Equal(t, DefaultAvcEncoder(), opt.Encoder)
		assert.Equal(t, 1920, opt.SizeLimit)
		assert.Equal(t, "50M", opt.BitrateLimit)
		assert.Equal(t, "0:v:0", opt.MapVideo)
		assert.Equal(t, "0:a:0?", opt.MapAudio)
		assert.Equal(t, MapVideo, opt.MapVideo)
		assert.Equal(t, MapAudio, opt.MapAudio)
	})

}

func TestOptions_VideoFilter(t *testing.T) {
	opt := &Options{
		Bin:          "",
		Encoder:      "intel",
		SizeLimit:    1500,
		BitrateLimit: "50M",
		MapVideo:     "",
		MapAudio:     "",
		MovFlags:     "",
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
