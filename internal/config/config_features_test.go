package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_DisableWebDAV(t *testing.T) {
	c := NewConfig(CliTestContext())

	c.options.Public = false
	c.options.ReadOnly = false
	c.options.Demo = false

	assert.False(t, c.DisableWebDAV())

	c.options.Public = true
	c.options.ReadOnly = false
	c.options.Demo = false

	assert.True(t, c.DisableWebDAV())

	c.options.Public = false
	c.options.ReadOnly = true
	c.options.Demo = false

	assert.False(t, c.DisableWebDAV())

	c.options.Public = false
	c.options.ReadOnly = false
	c.options.Demo = true

	assert.True(t, c.DisableWebDAV())

	c.options.Public = true
	c.options.ReadOnly = true
	c.options.Demo = true

	assert.True(t, c.DisableWebDAV())

	c.options.Public = false
	c.options.ReadOnly = false
	c.options.Demo = false

	assert.False(t, c.DisableWebDAV())
}

func TestConfig_DisableExifTool(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.False(t, c.DisableExifTool())

	c.options.ExifToolBin = "XXX"
	assert.True(t, c.DisableExifTool())
}

func TestConfig_ExifToolEnabled(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.True(t, c.ExifToolEnabled())

	c.options.ExifToolBin = "XXX"
	assert.False(t, c.ExifToolEnabled())
}

func TestConfig_DisableFaces(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.False(t, c.DisableFaces())
	c.options.DisableFaces = true
	assert.True(t, c.DisableFaces())
	c.options.DisableFaces = false
	c.options.DisableTensorFlow = true
	assert.True(t, c.DisableFaces())
	c.options.DisableTensorFlow = false
	assert.False(t, c.DisableFaces())
}

func TestConfig_DisableClassification(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.False(t, c.DisableClassification())
	c.options.DisableClassification = true
	assert.True(t, c.DisableClassification())
	c.options.DisableClassification = false
	c.options.DisableTensorFlow = true
	assert.True(t, c.DisableClassification())
	c.options.DisableTensorFlow = false
	assert.False(t, c.DisableClassification())
}

func TestConfig_DisableDarktable(t *testing.T) {
	c := NewConfig(CliTestContext())
	missing := c.DarktableBin() == ""

	assert.Equal(t, missing, c.DisableDarktable())
	c.options.DisableRaw = true
	assert.True(t, c.DisableDarktable())
	c.options.DisableRaw = false
	assert.Equal(t, missing, c.DisableDarktable())
	c.options.DisableDarktable = true
	assert.True(t, c.DisableDarktable())
	c.options.DisableDarktable = false
	assert.Equal(t, missing, c.DisableDarktable())
}

func TestConfig_DisableRawTherapee(t *testing.T) {
	c := NewConfig(CliTestContext())
	missing := c.RawTherapeeBin() == ""

	assert.Equal(t, missing, c.DisableRawTherapee())
	c.options.DisableRaw = true
	assert.True(t, c.DisableRawTherapee())
	c.options.DisableRaw = false
	assert.Equal(t, missing, c.DisableRawTherapee())
	c.options.DisableRawTherapee = true
	assert.True(t, c.DisableRawTherapee())
	c.options.DisableRawTherapee = false
	assert.Equal(t, missing, c.DisableRawTherapee())
}

func TestConfig_DisableImageMagick(t *testing.T) {
	c := NewConfig(CliTestContext())
	missing := c.ImageMagickBin() == ""

	assert.Equal(t, missing, c.DisableImageMagick())
	c.options.DisableRaw = true
	assert.Equal(t, missing, c.DisableImageMagick())
	c.options.DisableRaw = false
	assert.Equal(t, missing, c.DisableImageMagick())
	c.options.DisableImageMagick = true
	assert.True(t, c.DisableImageMagick())
	c.options.DisableImageMagick = false
	assert.Equal(t, missing, c.DisableImageMagick())
}

func TestConfig_DisableVips(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, false, c.DisableVips())
	c.options.DisableVips = true
	assert.True(t, c.DisableVips())
	c.options.DisableVips = false
	assert.Equal(t, false, c.DisableVips())
}

func TestConfig_DisableSips(t *testing.T) {
	c := NewConfig(CliTestContext())
	missing := c.SipsBin() == ""

	assert.Equal(t, missing, c.DisableSips())
	c.options.DisableSips = true
	assert.True(t, c.DisableSips())
	c.options.DisableSips = false
	assert.Equal(t, missing, c.DisableSips())
}

func TestConfig_DisableVector(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, c.Sponsor(), !c.DisableVectors())
	c.options.DisableVectors = true
	assert.True(t, c.DisableVectors())
	c.options.DisableVectors = false
	assert.Equal(t, c.Sponsor(), !c.DisableVectors())
}

func TestConfig_DisableRsvgConvert(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, c.Sponsor(), !c.DisableRsvgConvert())
	c.options.DisableVectors = true
	assert.True(t, c.DisableRsvgConvert())
	c.options.DisableVectors = false
	assert.Equal(t, c.Sponsor(), !c.DisableVectors())
}

func TestConfig_DisableRaw(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.False(t, c.DisableRaw())
	c.options.DisableRaw = true
	assert.True(t, c.DisableRaw())
	assert.True(t, c.DisableDarktable())
	assert.True(t, c.DisableRawTherapee())
	c.options.DisableRaw = false
	assert.False(t, c.DisableRaw())
	c.options.DisableDarktable = true
	c.options.DisableRawTherapee = true
	assert.False(t, c.DisableRaw())
	c.options.DisableDarktable = false
	c.options.DisableRawTherapee = false
	assert.False(t, c.DisableRaw())
	assert.False(t, c.DisableDarktable())
	assert.False(t, c.DisableRawTherapee())
}
