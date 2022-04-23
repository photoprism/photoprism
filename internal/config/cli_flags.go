package config

import (
	"github.com/photoprism/photoprism/pkg/list"
	"github.com/urfave/cli"
)

// CliFlags represents a list of command-line parameters.
type CliFlags []CliFlag

// Cli returns the currently active command-line parameters.
func (f CliFlags) Cli() (result []cli.Flag) {
	var tags []string

	switch {
	case Sponsor():
		tags = []string{EnvSponsor}
	}

	return f.Find(tags)
}

// Find finds command-line parameters based on a list of tags.
func (f CliFlags) Find(tags []string) (result []cli.Flag) {
	result = make([]cli.Flag, 0, len(f))

	for _, flag := range f {
		if len(flag.Tags) > 0 && !list.ContainsAny(flag.Tags, tags) {
			continue
		}

		result = append(result, flag.Flag)
	}

	return result
}
