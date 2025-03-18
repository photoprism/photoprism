package config

import (
	"reflect"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/pkg/list"
)

// CliFlag represents a command-line parameter.
type CliFlag struct {
	Flag       cli.DocGenerationFlag
	Tags       []string
	DocDefault string
}

// Skip checks if the parameter should be skipped based on a list of tags.
func (f CliFlag) Skip(tags []string) bool {
	return len(f.Tags) > 0 && !list.ContainsAny(f.Tags, tags)
}

// Fields returns the flag struct fields.
func (f CliFlag) Fields() reflect.Value {
	fields := reflect.ValueOf(f.Flag)

	for fields.Kind() == reflect.Ptr {
		fields = reflect.Indirect(fields)
	}

	return fields
}

// Default returns the default value.
func (f CliFlag) Default() string {
	if f.DocDefault != "" {
		return f.DocDefault
	}

	return f.Flag.GetValue()
}

// Hidden checks if the flag is hidden.
func (f CliFlag) Hidden() bool {
	field := f.Fields().FieldByName("Hidden")

	if !field.IsValid() || !field.Bool() {
		return false
	}

	return true
}

// EnvVar returns the names of the environment variables as a comma separated string.
func (f CliFlag) EnvVar() string {
	return strings.Join(f.EnvVars(), ", ")
}

// EnvVars returns the environment variable names as string slice.
func (f CliFlag) EnvVars() []string {
	field := f.Fields().FieldByName("EnvVars")

	if !field.IsValid() {
		return []string{}
	}

	if vars := field.Interface().([]string); len(vars) > 0 {
		return vars
	}

	return []string{}
}

// Name returns the command flag names as a comma-separated string.
func (f CliFlag) String() string {
	return strings.Join(f.Names(), ", ")
}

// Name returns the default command flag name.
func (f CliFlag) Name() string {
	if n := f.Flag.Names(); len(n) > 0 {
		return n[0]
	}

	return ""
}

// Names returns the command flag names as string slice.
func (f CliFlag) Names() []string {
	return f.Flag.Names()
}

// CommandFlag returns the full command flag based on the name.
func (f CliFlag) CommandFlag() string {
	n := strings.Split(f.String(), ",")
	return "--" + strings.TrimSpace(n[0])
}

// Usage returns the command flag usage.
func (f CliFlag) Usage() string {
	if list.Contains(f.Tags, EnvSponsor) {
		return f.Flag.GetUsage() + " *members only*"
	} else if list.Contains(f.Tags, Pro) {
		return f.Flag.GetUsage() + " *pro*"
	} else if list.Contains(f.Tags, Plus) {
		return f.Flag.GetUsage() + " *plus*"
	} else if list.Contains(f.Tags, Essentials) {
		return f.Flag.GetUsage() + " *essentials*"
	} else {
		return f.Flag.GetUsage()
	}
}
