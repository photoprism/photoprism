package config

import (
	"testing"

	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/stretchr/testify/assert"
)

func TestConfig_ConvertSize(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, int(720), c.JpegSize())
	c.options.JpegSize = 31000
	assert.Equal(t, int(30000), c.JpegSize())
	c.options.JpegSize = 800
	assert.Equal(t, int(800), c.JpegSize())
}

func TestConfig_JpegQuality(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, int(25), c.JpegQuality())
	c.options.JpegQuality = 110
	assert.Equal(t, int(100), c.JpegQuality())
	c.options.JpegQuality = 98
	assert.Equal(t, int(98), c.JpegQuality())
}

func TestConfig_ThumbFilter(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, thumb.ResampleFilter("cubic"), c.ThumbFilter())
	c.options.ThumbFilter = "blackman"
	assert.Equal(t, thumb.ResampleFilter("blackman"), c.ThumbFilter())
	c.options.ThumbFilter = "lanczos"
	assert.Equal(t, thumb.ResampleFilter("lanczos"), c.ThumbFilter())
	c.options.ThumbFilter = "linear"
	assert.Equal(t, thumb.ResampleFilter("linear"), c.ThumbFilter())
	c.options.ThumbFilter = "cubic"
	assert.Equal(t, thumb.ResampleFilter("cubic"), c.ThumbFilter())
}

func TestConfig_ThumbSizeUncached(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.False(t, c.ThumbUncached())
}

func TestConfig_ThumbSize(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, int(720), c.ThumbSizePrecached())
	c.options.ThumbSize = 7681
	assert.Equal(t, int(7680), c.ThumbSizePrecached())
}

func TestConfig_ThumbSizeUncached2(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, int(720), c.ThumbSizeUncached())
	c.options.ThumbSizeUncached = 7681
	assert.Equal(t, int(7680), c.ThumbSizeUncached())
	c.options.ThumbSizeUncached = 800
	c.options.ThumbSize = 900
	assert.Equal(t, int(900), c.ThumbSizeUncached())
}
