package config

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/thumb"
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
	assert.Equal(t, 7681, c.ThumbSizePrecached())
	c.options.ThumbSize = 15361
	assert.Equal(t, 15360, c.ThumbSizePrecached())
}

func TestConfig_ThumbSizeUncached2(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, 720, c.ThumbSizeUncached())
	c.options.ThumbSizeUncached = 7681
	assert.Equal(t, 7681, c.ThumbSizeUncached())
	c.options.ThumbSizeUncached = 15361
	assert.Equal(t, 15360, c.ThumbSizeUncached())
	c.options.ThumbSizeUncached = 800
	c.options.ThumbSize = 900
	assert.Equal(t, int(900), c.ThumbSizeUncached())
}

func TestInitThumbs_AdvertisesLargeSizes(t *testing.T) {
	origCached := thumb.SizeCached
	origOnDemand := thumb.SizeOnDemand
	defer func() {
		thumb.SizeCached = origCached
		thumb.SizeOnDemand = origOnDemand
		initThumbs()
	}()

	names := func() []string {
		out := make([]string, 0, len(Thumbs))
		for _, s := range Thumbs {
			out = append(out, s.Size)
		}
		return out
	}

	// A high on-demand limit exposes the 8K and 16K sizes to client apps.
	thumb.SizeCached = 1920
	thumb.SizeOnDemand = 15360
	initThumbs()
	assert.Contains(t, names(), "fit_15360")
	assert.Contains(t, names(), "fit_7680")

	// A low limit keeps the large on-demand sizes out of the advertised list.
	thumb.SizeOnDemand = 2560
	initThumbs()
	assert.NotContains(t, names(), "fit_15360")
	assert.NotContains(t, names(), "fit_7680")
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
	c.options.ThumbLibrary = Auto
	assert.Equal(t, "vips", c.ThumbLibrary())
	c.options.ThumbLibrary = "libvips"
	assert.Equal(t, "vips", c.ThumbLibrary())
	c.options.ThumbLibrary = "imaging"
	assert.Equal(t, "vips", c.ThumbLibrary())
	c.options.ThumbLibrary = "xxx"
	assert.Equal(t, "vips", c.ThumbLibrary())
}
