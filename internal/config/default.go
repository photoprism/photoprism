package config

import (
	"github.com/photoprism/photoprism/internal/i18n"
)

// DefaultTheme returns the default user interface theme name.
func (c *Config) DefaultTheme() string {
	if c.options.DefaultTheme == "" || !c.Sponsor() {
		return "default"
	}

	return c.options.DefaultTheme
}

// DefaultLocale returns the default user interface language locale name.
func (c *Config) DefaultLocale() string {
	if c.options.DefaultLocale == "" {
		return i18n.Default.Locale()
	}

	return c.options.DefaultLocale
}
