package config

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_RawEnabled(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.NotEqual(t, c.DisableRaw(), c.RawEnabled())
}

func TestConfig_RawTherapeeBin(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.True(t, strings.Contains(c.RawTherapeeBin(), "/bin/rawtherapee-cli"))
}

func TestConfig_RawTherapeeExclude(t *testing.T) {
	c := NewConfig(CliTestContext())

	c.options.RawTherapeeExclude = "foo,bar"
	assert.Equal(t, "foo,bar", c.RawTherapeeExclude())
	c.options.RawTherapeeExclude = ""
	assert.Equal(t, "", c.RawTherapeeExclude())
}

func TestConfig_RawTherapeeEnabled(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.True(t, c.RawTherapeeEnabled())

	c.options.DisableRawTherapee = true
	assert.False(t, c.RawTherapeeEnabled())
}

func TestConfig_DarktableBin(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.True(t, strings.Contains(c.DarktableBin(), "/bin/darktable-cli"))
}

func TestConfig_DarktableExclude(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "raf, cr3", c.DarktableExclude())
}

func TestConfig_DarktablePresets(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.False(t, c.RawPresets())
}

func TestConfig_DarktableEnabled(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.True(t, c.DarktableEnabled())

	c.options.DisableDarktable = true
	assert.False(t, c.DarktableEnabled())
}

func TestConfig_CreateDarktableCachePath(t *testing.T) {
	c := NewConfig(CliTestContext())
	path, err := c.CreateDarktableCachePath()

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, path, "")

	c.options.DarktableCachePath = "test"

	path, err = c.CreateDarktableCachePath()

	if err != nil {
		t.Fatal(err)
	}

	assert.Contains(t, path, "test")

	c.options.DarktableCachePath = ""
}

func TestConfig_CreateDarktableConfigPath(t *testing.T) {
	c := NewConfig(CliTestContext())
	path, err := c.CreateDarktableConfigPath()

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, path, "")

	c.options.DarktableConfigPath = "test"

	path, err = c.CreateDarktableConfigPath()

	if err != nil {
		t.Fatal(err)
	}

	assert.Contains(t, path, "test")

	c.options.DarktableConfigPath = ""
}
