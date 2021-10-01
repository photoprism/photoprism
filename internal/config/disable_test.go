package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_DisableBackups(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.False(t, c.DisableBackups())
}

func TestConfig_DisableWebDAV(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.False(t, c.DisableWebDAV())

	c.options.ReadOnly = true
	c.options.Demo = true
	assert.True(t, c.DisableWebDAV())
}

func TestConfig_DisableExifTool(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.False(t, c.DisableExifTool())

	c.options.ExifToolBin = "XXX"
	assert.True(t, c.DisableExifTool())
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
