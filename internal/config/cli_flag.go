package config

import (
	"reflect"
	"strings"

	"github.com/photoprism/photoprism/pkg/list"
	"github.com/urfave/cli"
)

// CliFlag represents a command-line parameter.
type CliFlag struct {
	Flag cli.DocGenerationFlag
	Tags []string
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

// EnvVar returns the flag environment variable name.
func (f CliFlag) EnvVar() string {
	field := f.Fields().FieldByName("EnvVar")

	if !field.IsValid() {
		return ""
	}

	return field.String()
}

// Name returns the command flag name.
func (f CliFlag) Name() string {
	return f.Flag.GetName()
}

// CommandFlag returns the full command flag based on the name.
func (f CliFlag) CommandFlag() string {
	n := strings.Split(f.Name(), ",")
	return "--" + n[0]
}

// Usage returns the command flag usage.
func (f CliFlag) Usage() string {
	if list.Contains(f.Tags, EnvSponsor) {
		return f.Flag.GetUsage() + " *members only*"
	} else {
		return f.Flag.GetUsage()
	}
}
