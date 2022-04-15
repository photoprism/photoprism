package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_ExifBruteForce(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, false, c.ExifBruteForce())
}

func TestConfig_ExifToolBin(t *testing.T) {
	c := NewConfig(CliTestContext())

	bin := c.ExifToolBin()
	assert.Equal(t, "/usr/bin/exiftool", bin)
}

func TestConfig_ExifToolJson(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, true, c.ExifToolJson())
	assert.Equal(t, c.DisableExifTool(), !c.ExifToolJson())

	c.options.DisableExifTool = true

	assert.Equal(t, false, c.ExifToolJson())
	assert.Equal(t, c.DisableExifTool(), !c.ExifToolJson())
}

func TestConfig_SidecarYaml(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, true, c.BackupYaml())
	assert.Equal(t, c.DisableBackups(), !c.BackupYaml())

	c.options.DisableBackups = true

	assert.Equal(t, false, c.BackupYaml())
	assert.Equal(t, c.DisableBackups(), !c.BackupYaml())
}
