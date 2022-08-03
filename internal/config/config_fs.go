package config

import (
	"github.com/photoprism/photoprism/pkg/fs"
	"regexp"
)

func (c *Config) StripSequenceRegex() *regexp.Regexp {
	if c.options.StripSequenceRegex == "" {
		return fs.StripSequenceRegex
	}

	exp, err := regexp.Compile(c.options.StripSequenceRegex)
	if err != nil {
		log.Errorf("config: can't parse StripSequenceRegex: %v", err)
		return fs.StripSequenceRegex
	}

	return exp
}
