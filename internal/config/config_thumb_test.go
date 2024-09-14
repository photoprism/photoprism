package config

import (
	"testing"

	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/stretchr/testify/assert"
)

func TestConfig_ConvertSize(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, 720, c.JpegSize())
	c.options.JpegSize = 31000
	assert.Equal(t, 30000, c.JpegSize())
	c.options.JpegSize = 800
	assert.Equal(t, 800, c.JpegSize())
}

func TestConfig_JpegQuality(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, thumb.QualityMedium, c.JpegQuality())
	c.options.JpegQuality = 110
	assert.Equal(t, thumb.QualityMax, c.JpegQuality())
	c.options.JpegQuality = 98
	assert.Equal(t, thumb.Quality(98), c.JpegQuality())
	c.options.JpegQuality = -1
	assert.Equal(t, thumb.QualityMedium, c.JpegQuality())
	c.options.JpegQuality = 0
	assert.Equal(t, thumb.QualityMedium, c.JpegQuality())
	c.options.JpegQuality = 25
	assert.Equal(t, thumb.Quality(25), c.JpegQuality())
	c.options.JpegQuality = 85
	assert.Equal(t, thumb.Quality(85), c.JpegQuality())
	c.options.JpegQuality = 0
	assert.Equal(t, thumb.QualityMedium, c.JpegQuality())
}

func TestConfig_ThumbFilter(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, thumb.ResampleLanczos, c.ThumbFilter())
	c.options.ThumbFilter = "blackman"
	assert.Equal(t, thumb.ResampleBlackman, c.ThumbFilter())
	c.options.ThumbFilter = "lanczos"
	assert.Equal(t, thumb.ResampleLanczos, c.ThumbFilter())
	c.options.ThumbFilter = "linear"
	assert.Equal(t, thumb.ResampleLinear, c.ThumbFilter())
	c.options.ThumbFilter = "auto"
	assert.Equal(t, thumb.ResampleLanczos, c.ThumbFilter())
	c.options.ThumbFilter = ""
	assert.Equal(t, thumb.ResampleLanczos, c.ThumbFilter())
}

func TestConfig_ThumbSizeUncached(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.False(t, c.ThumbUncached())
}

func TestConfig_ThumbSize(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, 720, c.ThumbSizePrecached())
	c.options.ThumbSize = 7681
	assert.Equal(t, 7680, c.ThumbSizePrecached())
}

func TestConfig_ThumbSizeUncached2(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, 720, c.ThumbSizeUncached())
	c.options.ThumbSizeUncached = 7681
	assert.Equal(t, 7680, c.ThumbSizeUncached())
	c.options.ThumbSizeUncached = 800
	c.options.ThumbSize = 900
	assert.Equal(t, int(900), c.ThumbSizeUncached())
}

func TestConfig_PngSize(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, 720, c.PngSize())
	c.options.PngSize = 700681
	assert.Equal(t, 30000, c.PngSize())
	c.options.PngSize = 1240
	assert.Equal(t, 1240, c.PngSize())
}

func TestConfig_ThumbLibrary(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.False(t, c.DisableVips())
	c.options.ThumbLibrary = "auto"
	assert.Equal(t, "vips", c.ThumbLibrary())
	c.options.DisableVips = true
	assert.Equal(t, "imaging", c.ThumbLibrary())
	c.options.DisableVips = false
	c.options.ThumbLibrary = "libvips"
	assert.Equal(t, "vips", c.ThumbLibrary())
	c.options.ThumbLibrary = "xxx"
	assert.Equal(t, "vips", c.ThumbLibrary())
}
