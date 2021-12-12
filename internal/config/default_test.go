package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_DefaultTheme(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "default", c.DefaultTheme())
	c.options.Sponsor = false
	c.options.DefaultTheme = "grayscale"
	assert.Equal(t, "default", c.DefaultTheme())
	c.options.Sponsor = true
	assert.Equal(t, "grayscale", c.DefaultTheme())
	c.options.DefaultTheme = ""
	assert.Equal(t, "default", c.DefaultTheme())
	c.options.Sponsor = false
	assert.Equal(t, "default", c.DefaultTheme())
}

func TestConfig_DefaultLocale(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "en", c.DefaultLocale())
	c.options.DefaultLocale = "de"
	assert.Equal(t, "de", c.DefaultLocale())
	c.options.DefaultLocale = ""
	assert.Equal(t, "en", c.DefaultLocale())
}
