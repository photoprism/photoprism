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
