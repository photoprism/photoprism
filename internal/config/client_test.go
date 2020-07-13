package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConfig_PublicConfig(t *testing.T) {
	config := NewTestConfig()
	result := config.PublicConfig()
	assert.IsType(t, ClientConfig{}, result)
	assert.Equal(t, true, result.Public)
	c2 := NewTestErrorConfig()
	result2 := c2.PublicConfig()
	assert.IsType(t, ClientConfig{}, result2)
	assert.Equal(t, false, result2.Public)
}

func TestConfig_GuestConfig(t *testing.T) {
	config := NewTestConfig()
	result := config.GuestConfig()
	assert.IsType(t, ClientConfig{}, result)
	assert.Equal(t, true, result.Public)
}
