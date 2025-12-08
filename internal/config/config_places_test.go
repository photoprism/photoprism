package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_GeoApi(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "places", c.GeoApi())
	c.options.DisablePlaces = true
	assert.Equal(t, "", c.GeoApi())
}

func TestConfig_PlacesLocale(t *testing.T) {
	c := NewConfig(CliTestContext())

	c.options.PlacesLocale = ""
	assert.Equal(t, "local", c.PlacesLocale())
	c.options.PlacesLocale = "local"
	assert.Equal(t, "local", c.PlacesLocale())
	c.options.PlacesLocale = "EN"
	assert.Equal(t, "en", c.PlacesLocale())
	c.options.PlacesLocale = "EN_US"
	assert.Equal(t, "en-US", c.PlacesLocale())
	c.options.PlacesLocale = ""
	assert.Equal(t, "local", c.PlacesLocale())
}
