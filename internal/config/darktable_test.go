package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConfig_DarktableUnlock(t *testing.T) {
	c := NewTestConfig()
	assert.False(t, c.DarktableUnlock())
}

func TestConfig_Darktablebin(t *testing.T) {
	c := NewTestConfig()
	assert.Equal(t, "/usr/bin/darktable-cli", c.DarktableBin())
}
