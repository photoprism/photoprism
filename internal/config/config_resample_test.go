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
	c.options.JpegQuality = "110"
	assert.Equal(t, thumb.QualityMedium, c.JpegQuality())
	c.options.JpegQuality = "98"
	assert.Equal(t, thumb.Quality(98), c.JpegQuality())
	c.options.JpegQuality = ""
	assert.Equal(t, thumb.QualityMedium, c.JpegQuality())
	c.options.JpegQuality = "best "
	assert.Equal(t, thumb.QualityMax, c.JpegQuality())
	c.options.JpegQuality = "high"
	assert.Equal(t, thumb.QualityHigh, c.JpegQuality())
	c.options.JpegQuality = "med "
	assert.Equal(t, thumb.QualityMedium, c.JpegQuality())
	c.options.JpegQuality = "medium "
	assert.Equal(t, thumb.QualityMedium, c.JpegQuality())
	c.options.JpegQuality = "low "
	assert.Equal(t, thumb.QualityLow, c.JpegQuality())
	c.options.JpegQuality = "max"
	assert.Equal(t, thumb.QualityMax, c.JpegQuality())
	c.options.JpegQuality = "min "
	assert.Equal(t, thumb.QualityMin, c.JpegQuality())
	c.options.JpegQuality = "default"
	assert.Equal(t, thumb.QualityMedium, c.JpegQuality())
}

func TestConfig_ThumbFilter(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, thumb.ResampleAuto, c.ThumbFilter())
	c.options.ThumbFilter = "blackman"
	assert.Equal(t, thumb.ResampleBlackman, c.ThumbFilter())
	c.options.ThumbFilter = "lanczos"
	assert.Equal(t, thumb.ResampleLanczos, c.ThumbFilter())
	c.options.ThumbFilter = "linear"
	assert.Equal(t, thumb.ResampleLinear, c.ThumbFilter())
	c.options.ThumbFilter = "auto"
	assert.Equal(t, thumb.ResampleAuto, c.ThumbFilter())
	c.options.ThumbFilter = ""
	assert.Equal(t, thumb.ResampleAuto, c.ThumbFilter())
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
