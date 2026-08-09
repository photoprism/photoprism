package encode

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewOptions(t *testing.T) {
	t.Run("Defaults", func(t *testing.T) {
		opt := NewVideoOptions("", "", 0, 50, "", "", "", "")
		assert.Equal(t, "ffmpeg", opt.Bin)
		assert.Equal(t, FFmpegBin, opt.Bin)
		assert.Equal(t, DefaultAvcEncoder(), opt.Encoder)
		assert.Equal(t, 1920, opt.SizeLimit)
		assert.Equal(t, DefaultQuality, opt.Quality)
		assert.Equal(t, "50", opt.QvQuality())
		assert.Equal(t, "25", opt.GlobalQuality())
		assert.Equal(t, "25", opt.CrfQuality())
		assert.Equal(t, "25", opt.QpQuality())
		assert.Equal(t, "25", opt.CqQuality())
		assert.Equal(t, PresetFast, opt.Preset)
		assert.Equal(t, "", opt.Device)
		assert.Equal(t, "0:v:0", opt.MapVideo)
		assert.Equal(t, "0:a:0?", opt.MapAudio)
		assert.Equal(t, DefaultMapVideo, opt.MapVideo)
		assert.Equal(t, DefaultMapAudio, opt.MapAudio)
	})
}

func TestOptions_VideoFilter(t *testing.T) {
	opt := &Options{
		Bin:       "",
		Encoder:   "intel",
		SizeLimit: 1500,
		MapVideo:  "",
		MapAudio:  "",
		MovFlags:  "",
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
	t.Run("V360", func(t *testing.T) {
		v360Opt := &Options{SizeLimit: 1500, V360: "v360=input=dfisheye:output=e"}
		r := v360Opt.VideoFilter("")
		assert.True(t, strings.HasPrefix(r, "v360=input=dfisheye:output=e,"))
		assert.Contains(t, r, "min(1500, iw)")
	})
	t.Run("NoV360", func(t *testing.T) {
		r := opt.VideoFilter("")
		assert.NotContains(t, r, "v360")
	})
}
