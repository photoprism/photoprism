package photoprism

import (
	"github.com/photoprism/photoprism/internal/config"
)

var conf *config.Config

// SetConfig initialises package-level access to the shared Config.
func SetConfig(c *config.Config) {
	if c == nil {
		panic("config is missing")
	}

	conf = c
}

// Config returns the shared Config, panicking if it has not been set.
func Config() *config.Config {
	if conf == nil {
		panic("config is missing")
	}

	return conf
}
