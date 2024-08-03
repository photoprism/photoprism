package config

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/media"
)

func TestConfig_HeifConvertBin(t *testing.T) {
	c := NewConfig(CliTestContext())

	bin := c.HeifConvertBin()
	assert.Contains(t, bin, "/bin/heif-")
}

func TestConfig_HeifConvertOrientation(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, media.ResetOrientation, c.HeifConvertOrientation())
	c.Options().HeifConvertOrientation = media.KeepOrientation
	assert.Equal(t, media.KeepOrientation, c.HeifConvertOrientation())
	c.Options().HeifConvertOrientation = ""
	assert.Equal(t, media.ResetOrientation, c.HeifConvertOrientation())
}

func TestConfig_HeifConvertEnabled(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.True(t, c.HeifConvertEnabled())

	c.options.DisableHeifConvert = true
	assert.False(t, c.HeifConvertEnabled())
}

func TestConfig_SipsBin(t *testing.T) {
	c := NewConfig(CliTestContext())

	bin := c.SipsBin()
	assert.Equal(t, "", bin)
}

func TestConfig_SipsEnabled(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.NotEqual(t, c.DisableSips(), c.SipsEnabled())
}

func TestConfig_SipsExclude(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "avif, avifs, thm", c.SipsExclude())
}

func TestConfig_RsvgConvertBin(t *testing.T) {
	c := NewConfig(CliTestContext())

	bin := c.RsvgConvertBin()
	assert.Contains(t, bin, "/bin/rsvg-convert")
}

func TestConfig_RsvgConvertEnabled(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.True(t, c.RsvgConvertEnabled())

	c.options.DisableVectors = true
	assert.False(t, c.RsvgConvertEnabled())
}

func TestConfig_VectorEnabled(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.True(t, c.VectorEnabled())
	c.options.DisableVectors = true
	assert.False(t, c.VectorEnabled())
	c.options.DisableVectors = false
}

func TestConfig_RsvgConvertBin2(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Contains(t, c.RsvgConvertBin(), "rsvg-convert")
}

func TestConfig_ImageMagickBin(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Contains(t, c.ImageMagickBin(), "convert")
}

func TestConfig_ImageMagickEnabled(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.True(t, c.ImageMagickEnabled())
	c.options.DisableImageMagick = true
	assert.False(t, c.ImageMagickEnabled())
	c.options.DisableImageMagick = false
}

func TestConfig_JpegXLDecoderBin(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Contains(t, c.JpegXLDecoderBin(), "djxl")
}

func TestConfig_JpegXLEnabled(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.True(t, c.JpegXLEnabled())
	c.options.DisableJpegXL = true
	assert.False(t, c.JpegXLEnabled())
	c.options.DisableJpegXL = false
}

func TestConfig_DisableJpegXL(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.False(t, c.DisableJpegXL())
	c.options.DisableJpegXL = true
	assert.True(t, c.DisableJpegXL())
	c.options.DisableJpegXL = false
}
