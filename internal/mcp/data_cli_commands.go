package mcp

import (
	"reflect"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/commands/catalog"
)

// cliRootName is the name of the root command that prefixes every command path.
const cliRootName = "photoprism"

// CliCommand represents a visible CLI command with its own visible flags.
type CliCommand struct {
	Command     string    `json:"command"`
	Usage       string    `json:"usage"`
	Description string    `json:"description,omitempty"`
	ArgsUsage   string    `json:"args_usage,omitempty"`
	Aliases     []string  `json:"aliases"`
	Flags       []CliFlag `json:"flags"`
}

// CliFlag represents a visible flag of a CLI command.
type CliFlag struct {
	Name     string   `json:"name"`
	Aliases  []string `json:"aliases"`
	Type     string   `json:"type"`
	Usage    string   `json:"usage"`
	Required bool     `json:"required"`
	Default  string   `json:"default,omitempty"`
	Env      []string `json:"env"`
}

// CliCommandsResource contains the JSON payload for the cli-commands resource.
type CliCommandsResource struct {
	Edition string       `json:"edition"`
	Items   []CliCommand `json:"items"`
}

// buildCliCommands returns the visible commands with their own visible flags in the order of
// `photoprism show commands`, or nil if no command list was passed.
func buildCliCommands(commands []*cli.Command, globalFlags []cli.Flag) []CliCommand {
	if commands == nil {
		return nil
	}

	global := catalog.FlagsToCatalog(globalFlags, false)
	sources := cliCommandSources(commands)
	items := make([]CliCommand, 0, len(commands))

	for _, root := range commands {
		if root == nil || root.Hidden {
			continue
		}

		for _, cmd := range catalog.BuildFlat(root, 1, cliRootName, false, global) {
			items = append(items, newCliCommand(cmd, sources[cmd.FullName]))
		}
	}

	return items
}

// cliCommandSources maps each command path to its command, or to nil if the path is not unique.
func cliCommandSources(commands []*cli.Command) map[string]*cli.Command {
	sources := make(map[string]*cli.Command)

	var walk func(cmds []*cli.Command, parent string)

	walk = func(cmds []*cli.Command, parent string) {
		for _, c := range cmds {
			if c == nil {
				continue
			}

			name := strings.TrimSpace(parent + " " + c.Name)

			if _, found := sources[name]; found {
				sources[name] = nil
			} else {
				sources[name] = c
			}

			walk(c.Subcommands, name)
		}
	}

	walk(commands, cliRootName)

	return sources
}

// newCliCommand converts a catalog command into a CliCommand. A flag keeps its default only if
// its definition in src shows that the default is a literal.
func newCliCommand(cmd catalog.Command, src *cli.Command) CliCommand {
	flags := make([]CliFlag, 0, len(cmd.Flags))

	for _, f := range cmd.Flags {
		flag := CliFlag{
			Name:     f.Name,
			Aliases:  cloneStrings(f.Aliases),
			Type:     f.Type,
			Usage:    f.Usage,
			Required: f.Required,
			Env:      cloneStrings(f.Env),
		}

		if src != nil && flagDefaultAllowed(findFlag(src.Flags, f.Name)) {
			flag.Default = f.Default
		}

		flags = append(flags, flag)
	}

	return CliCommand{
		Command:     cmd.FullName,
		Usage:       cmd.Usage,
		Description: cmd.Description,
		ArgsUsage:   cmd.ArgsUsage,
		Aliases:     cloneStrings(cmd.Aliases),
		Flags:       flags,
	}
}

// findFlag returns the flag that the catalog lists under the specified name, or nil if there is none.
func findFlag(flags []cli.Flag, name string) cli.Flag {
	for _, f := range flags {
		if f != nil && catalog.DescribeFlag(f).Name == name {
			return f
		}
	}

	return nil
}

// flagDefaultAllowed reports whether the default of a flag is a literal of its definition. It is not
// if the flag can be set through an environment variable or a file, or holds a generic value that
// parsed arguments change.
func flagDefaultAllowed(f cli.Flag) bool {
	for f != nil {
		if e, ok := f.(interface{ GetEnvVars() []string }); ok && len(e.GetEnvVars()) > 0 {
			return false
		}

		if _, ok := f.(*cli.GenericFlag); ok || flagFilePath(f) != "" {
			return false
		}

		w, ok := f.(interface{ Unwrap() cli.Flag })

		if !ok {
			return true
		}

		f = w.Unwrap()
	}

	return false
}

// flagFilePath returns the FilePath field of a flag struct, or an empty string if it has none.
func flagFilePath(f cli.Flag) string {
	v := reflect.ValueOf(f)

	if v.Kind() != reflect.Pointer || v.IsNil() {
		return ""
	}

	v = v.Elem()

	if v.Kind() != reflect.Struct {
		return ""
	}

	field, ok := v.Type().FieldByName("FilePath")

	if !ok || field.Type.Kind() != reflect.String {
		return ""
	}

	value, err := v.FieldByIndexErr(field.Index)

	if err != nil {
		return ""
	}

	return value.String()
}

// cloneStrings returns a copy of the specified strings that is never nil.
func cloneStrings(s []string) []string {
	return append(make([]string, 0, len(s)), s...)
}
