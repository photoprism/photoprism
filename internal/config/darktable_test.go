package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConfig_DarktableUnlock(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.False(t, c.DarktableUnlock())
}

func TestConfig_Darktablebin(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "/usr/bin/darktable-cli", c.DarktableBin())
}
