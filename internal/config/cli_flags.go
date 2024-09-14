package config

import (
	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/list"
)

// CliFlags represents a list of command-line parameters.
type CliFlags []CliFlag

// Cli returns the currently active command-line parameters.
func (f CliFlags) Cli() (result []cli.Flag) {
	result = make([]cli.Flag, 0, len(f))

	for _, flag := range f {
		result = append(result, flag.Flag)
	}

	return result
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

// Remove removes command flags by name.
func (f CliFlags) Remove(names []string) (result CliFlags) {
	result = make(CliFlags, 0, len(f))

	for _, flag := range f {
		if list.Contains(names, flag.Name()) {
			continue
		}

		result = append(result, flag)
	}

	return result
}

// Replace replaces an existing command flag by name and returns true if successful.
func (f CliFlags) Replace(name string, replacement CliFlag) CliFlags {
	done := false

	for i, flag := range f {
		if !done && flag.Name() == name {
			f[i] = replacement
			done = true
		}
	}

	if !done {
		log.Warnf("config: failed to replace cli flag %s", clean.Log(name))
	}

	return f
}

// Insert inserts command flags, if possible after the flag specified by name.
func (f CliFlags) Insert(name string, insert []CliFlag) (result CliFlags) {
	result = make(CliFlags, 0, len(f)+len(insert))

	done := false

	for _, flag := range f {
		result = append(result, flag)

		if !done && flag.Name() == name {
			result = append(result, insert...)
			done = true
		}
	}

	if !done {
		log.Warnf("config: failed to insert cli flags after %s", clean.Log(name))
		result = append(result, insert...)
	}

	return result
}

// InsertBefore inserts command flags, if possible before the flag specified by name.
func (f CliFlags) InsertBefore(name string, insert []CliFlag) (result CliFlags) {
	result = make(CliFlags, 0, len(f)+len(insert))

	done := false

	for _, flag := range f {
		if !done && flag.Name() == name {
			result = append(result, insert...)
			done = true
		}

		result = append(result, flag)
	}

	if !done {
		log.Warnf("config: failed to insert cli flags before %s", clean.Log(name))
		result = append(result, insert...)
	}

	return result
}

// Prepend adds command flags at the beginning.
func (f CliFlags) Prepend(el []CliFlag) (result CliFlags) {
	result = make(CliFlags, 0, len(f)+len(el))

	result = append(result, el...)
	return append(result, f...)
}
