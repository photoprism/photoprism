package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_RawtherapeeBin(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "/usr/bin/rawtherapee-cli", c.RawtherapeeBin())
}

func TestConfig_DarktableBin(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "/usr/bin/darktable-cli", c.DarktableBin())
}

func TestConfig_DarktablePresets(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.False(t, c.DarktablePresets())
}

func TestConfig_SipsBin(t *testing.T) {
	c := NewConfig(CliTestContext())

	bin := c.SipsBin()
	assert.Equal(t, "", bin)
}

func TestConfig_HeifConvertBin(t *testing.T) {
	c := NewConfig(CliTestContext())

	bin := c.HeifConvertBin()
	assert.Equal(t, "/usr/bin/heif-convert", bin)
}
