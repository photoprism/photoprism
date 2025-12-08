package config

import (
	"github.com/photoprism/photoprism/internal/service/hub/places"
	"github.com/photoprism/photoprism/pkg/clean"
)

// GeoApi returns the preferred geocoding api (places, or none).
func (c *Config) GeoApi() string {
	if c.options.DisablePlaces {
		return ""
	}

	return "places"
}

// PlacesLocale returns the locale name used for geocoding.
func (c *Config) PlacesLocale() string {
	return clean.WebLocale(c.options.PlacesLocale, places.LocalLocale)
}
