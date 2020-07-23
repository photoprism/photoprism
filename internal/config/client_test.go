package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
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
	config.params.Experimental = true
	config.params.ReadOnly = true

	result := config.Flags()
	assert.Equal(t, []string{"public", "debug", "experimental", "readonly", "settings"}, result)

	config.params.Experimental = false
	config.params.ReadOnly = false
}
