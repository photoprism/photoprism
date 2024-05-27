package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

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
