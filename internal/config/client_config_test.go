package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_PublicConfig(t *testing.T) {
	config := TestConfig()
	result := config.PublicConfig()
	assert.IsType(t, ClientConfig{}, result)
	assert.Equal(t, true, result.Public)
	c2 := NewTestErrorConfig()
	result2 := c2.PublicConfig()
	assert.IsType(t, ClientConfig{}, result2)
	assert.Equal(t, false, result2.Public)
}

func TestConfig_GuestConfig(t *testing.T) {
	config := TestConfig()
	result := config.GuestConfig()
	assert.IsType(t, ClientConfig{}, result)
	assert.Equal(t, true, result.Public)
	assert.Equal(t, false, result.Experimental)
	assert.Equal(t, true, result.ReadOnly)
}

func TestConfig_Flags(t *testing.T) {
	config := TestConfig()
	config.options.Experimental = true
	config.options.ReadOnly = true
	config.settings.UI.Scrollbar = false

	result := config.Flags()
	assert.Equal(t, []string{"public", "debug", "test", "sponsor", "experimental", "readonly", "settings", "hide-scrollbar"}, result)

	config.options.Experimental = false
	config.options.ReadOnly = false
}
