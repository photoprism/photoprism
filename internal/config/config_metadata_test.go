package config

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConfig_ExifBruteForce(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, false, c.ExifBruteForce())
}

func TestConfig_ExifToolBin(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.True(t, strings.Contains(c.ExifToolBin(), "/bin/exiftool"))
}

func TestConfig_ExifToolJson(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, true, c.ExifToolJson())
	assert.Equal(t, c.DisableExifTool(), !c.ExifToolJson())

	c.options.DisableExifTool = true

	assert.Equal(t, false, c.ExifToolJson())
	assert.Equal(t, c.DisableExifTool(), !c.ExifToolJson())
}

func TestConfig_ConvertTimeout(t *testing.T) {
	c := NewConfig(CliTestContext())
	t.Run("Default", func(t *testing.T) {
		c.options.ConvertTimeout = 0
		assert.Equal(t, time.Duration(DefaultConvertTimeout)*time.Minute, c.ConvertTimeout())
	})
	t.Run("Configured", func(t *testing.T) {
		c.options.ConvertTimeout = 5
		assert.Equal(t, 5*time.Minute, c.ConvertTimeout())
	})
	t.Run("Disabled", func(t *testing.T) {
		c.options.ConvertTimeout = -1
		assert.Equal(t, time.Duration(0), c.ConvertTimeout())
	})
	t.Run("AboveMaximum", func(t *testing.T) {
		c.options.ConvertTimeout = MaxConvertTimeout + 1
		assert.Equal(t, time.Duration(MaxConvertTimeout)*time.Minute, c.ConvertTimeout())
	})
}

func TestConfig_TranscodeTimeout(t *testing.T) {
	c := NewConfig(CliTestContext())
	t.Run("DefaultIsUnlimited", func(t *testing.T) {
		// Transcoding time grows with the length of the source, so a default limit would fail
		// long recordings on slow hardware rather than bound anything.
		c.options.TranscodeTimeout = 0
		assert.Equal(t, time.Duration(0), c.TranscodeTimeout())
	})
	t.Run("Configured", func(t *testing.T) {
		c.options.TranscodeTimeout = 120
		assert.Equal(t, 120*time.Minute, c.TranscodeTimeout())
	})
	t.Run("Disabled", func(t *testing.T) {
		c.options.TranscodeTimeout = -1
		assert.Equal(t, time.Duration(0), c.TranscodeTimeout())
	})
	t.Run("AboveMaximum", func(t *testing.T) {
		c.options.TranscodeTimeout = MaxConvertTimeout + 1
		assert.Equal(t, time.Duration(MaxConvertTimeout)*time.Minute, c.TranscodeTimeout())
	})
	t.Run("IndependentOfConversion", func(t *testing.T) {
		c.options.ConvertTimeout = 5
		c.options.TranscodeTimeout = 600
		assert.Equal(t, 5*time.Minute, c.ConvertTimeout())
		assert.Equal(t, 600*time.Minute, c.TranscodeTimeout())
	})
}
