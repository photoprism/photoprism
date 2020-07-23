package photoprism

import (
	"github.com/photoprism/photoprism/internal/config"
)

var conf *config.Config

func SetConfig(c *config.Config) {
	if c == nil {
		panic("config is nil")
	}

	conf = c
}

func Config() *config.Config {
	if conf == nil {
		panic("config is nil")
	}

	return conf
}
