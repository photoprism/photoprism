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

func TestConfig_RawtherapeeBin(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.True(t, strings.Contains(c.RawtherapeeBin(), "/bin/rawtherapee-cli"))
}

func TestConfig_RawtherapeeBlacklist(t *testing.T) {
	c := NewConfig(CliTestContext())

	c.options.RawtherapeeBlacklist = "foo,bar"
	assert.Equal(t, "foo,bar", c.RawtherapeeBlacklist())
	c.options.RawtherapeeBlacklist = ""
	assert.Equal(t, "", c.RawtherapeeBlacklist())
}

func TestConfig_RawtherapeeEnabled(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.True(t, c.RawtherapeeEnabled())

	c.options.DisableRawtherapee = true
	assert.False(t, c.RawtherapeeEnabled())
}

func TestConfig_DarktableBin(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.True(t, strings.Contains(c.DarktableBin(), "/bin/darktable-cli"))
}

func TestConfig_DarktableBlacklist(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "raf,cr3", c.DarktableBlacklist())
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

func TestConfig_SipsBin(t *testing.T) {
	c := NewConfig(CliTestContext())

	bin := c.SipsBin()
	assert.Equal(t, "", bin)
}

func TestConfig_SipsEnabled(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.NotEqual(t, c.DisableSips(), c.SipsEnabled())
}

func TestConfig_HeifConvertBin(t *testing.T) {
	c := NewConfig(CliTestContext())

	bin := c.HeifConvertBin()
	assert.Contains(t, bin, "/bin/heif-convert")
}

func TestConfig_HeifConvertScript(t *testing.T) {
	c := NewConfig(CliTestContext())

	bin := c.HeifConvertScript()
	assert.Contains(t, bin, "/bin/heif-convert.sh")
}

func TestConfig_HeifConvertEnabled(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.True(t, c.HeifConvertEnabled())

	c.options.DisableHeifConvert = true
	assert.False(t, c.HeifConvertEnabled())
}
