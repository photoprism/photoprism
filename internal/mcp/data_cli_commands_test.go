package mcp

import (
	"context"
	"encoding/json"
	"flag"
	"strings"
	"testing"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/commands/catalog"
)

// testGeneric is a cli.Generic value for tests.
type testGeneric struct {
	value string
}

// Set sets the value.
func (g *testGeneric) Set(value string) error {
	g.value = value
	return nil
}

// String returns the value.
func (g *testGeneric) String() string {
	return g.value
}

// testWrappedFlag wraps a string flag like commands.LocalStringFlag.
type testWrappedFlag struct {
	*cli.StringFlag
}

// Unwrap returns the wrapped string flag.
func (f *testWrappedFlag) Unwrap() cli.Flag {
	return f.StringFlag
}

// testFieldWrappedFlag wraps a string flag in a named field, so nothing of the inner flag is promoted.
type testFieldWrappedFlag struct {
	inner *cli.StringFlag
}

// Apply applies the inner flag.
func (f *testFieldWrappedFlag) Apply(set *flag.FlagSet) error {
	return f.inner.Apply(set)
}

// Names returns the names of the inner flag.
func (f *testFieldWrappedFlag) Names() []string {
	return f.inner.Names()
}

// IsSet reports whether the inner flag is set.
func (f *testFieldWrappedFlag) IsSet() bool {
	return f.inner.IsSet()
}

// String returns the help text of the inner flag.
func (f *testFieldWrappedFlag) String() string {
	return f.inner.String()
}

// Unwrap returns the inner flag.
func (f *testFieldWrappedFlag) Unwrap() cli.Flag {
	return f.inner
}

// testEnvVars lists the environment variables of the env-backed flags in testCommands.
var testEnvVars = []string{
	"PHOTOPRISM_TEST_MCP_PASSWORD",
	"PHOTOPRISM_TEST_MCP_COUNT",
	"PHOTOPRISM_TEST_MCP_PATH",
	"PHOTOPRISM_TEST_MCP_GENERIC",
	"PHOTOPRISM_TEST_MCP_WRAPPED",
}

// testCommands returns a new synthetic command list with visible, hidden, and env-backed commands and flags.
func testCommands() []*cli.Command {
	return []*cli.Command{
		{
			Name:    "users",
			Usage:   "Manages user accounts",
			Aliases: []string{"u"},
			Subcommands: []*cli.Command{
				{
					Name:        "add",
					Usage:       "Adds a new user account",
					Description: "Creates an account with the specified role.",
					ArgsUsage:   "[username]",
					Flags: []cli.Flag{
						&cli.StringFlag{Name: "role", Aliases: []string{"r"}, Usage: "account role", Value: "admin"},
						&cli.StringFlag{Name: "secret", Usage: "hidden option", Hidden: true},
						&cli.StringFlag{Name: "password", Usage: "account password", Value: "literal", EnvVars: []string{"PHOTOPRISM_TEST_MCP_PASSWORD"}},
						&cli.IntFlag{Name: "count", Usage: "number of accounts", Value: 3, EnvVars: []string{"PHOTOPRISM_TEST_MCP_COUNT"}},
						&cli.PathFlag{Name: "path", Usage: "storage path", Value: "/literal", EnvVars: []string{"PHOTOPRISM_TEST_MCP_PATH"}},
						&cli.GenericFlag{Name: "generic", Usage: "generic option", Value: &testGeneric{"literal"}, EnvVars: []string{"PHOTOPRISM_TEST_MCP_GENERIC"}},
						&cli.GenericFlag{Name: "parsed", Usage: "generic option set by arguments", Value: &testGeneric{"literal"}},
						&cli.StringFlag{Name: "name-file", Usage: "file option", Value: "literal", FilePath: "/nonexistent/photoprism-test-mcp"},
						&testWrappedFlag{&cli.StringFlag{Name: "wrapped", Usage: "wrapped option", Value: "literal", EnvVars: []string{"PHOTOPRISM_TEST_MCP_WRAPPED"}}},
						&cli.BoolFlag{Name: "debug", Usage: "repeats a global flag"},
					},
				},
				{Name: "purge", Usage: "Hidden subcommand", Hidden: true},
				{Name: "help", Usage: "Nested help"},
			},
		},
		{Name: "maintenance", Usage: "Hidden top-level command", Hidden: true, Subcommands: []*cli.Command{{Name: "run", Usage: "Runs maintenance"}}},
		{Name: "help", Aliases: []string{"h"}, Usage: "Shows a list of commands or help for one command"},
	}
}

// testGlobalFlags returns the synthetic global flags for testCommands.
func testGlobalFlags() []cli.Flag {
	return []cli.Flag{&cli.BoolFlag{Name: "debug", Usage: "enables debug mode"}}
}

// applyTwice applies each env-backed flag to two fresh flag sets, which makes urfave/cli keep the
// environment value as the default of scalar flags.
func applyTwice(t *testing.T, commands []*cli.Command) {
	t.Helper()

	var walk func(cmds []*cli.Command)

	walk = func(cmds []*cli.Command) {
		for _, c := range cmds {
			for _, f := range c.Flags {
				if e, ok := f.(interface{ GetEnvVars() []string }); !ok || len(e.GetEnvVars()) == 0 {
					continue
				}

				for range 2 {
					require.NoError(t, f.Apply(flag.NewFlagSet("test", flag.ContinueOnError)))
				}
			}

			walk(c.Subcommands)
		}
	}

	walk(commands)
}

// TestBuildCliCommands checks which commands and flags are listed and how they are described.
func TestBuildCliCommands(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		items := buildCliCommands(testCommands(), testGlobalFlags())

		names := make([]string, 0, len(items))
		for _, item := range items {
			names = append(names, item.Command)
		}
		require.Equal(t, []string{"photoprism users", "photoprism users add", "photoprism help"}, names)

		add := items[1]
		require.Equal(t, "Adds a new user account", add.Usage)
		require.Equal(t, "Creates an account with the specified role.", add.Description)
		require.Equal(t, "[username]", add.ArgsUsage)
		require.Equal(t, []string{"u"}, items[0].Aliases)
		require.Equal(t, []string{}, add.Aliases)

		flags := make(map[string]CliFlag, len(add.Flags))
		for _, f := range add.Flags {
			flags[f.Name] = f
		}
		require.NotContains(t, flags, "--secret", "hidden flag")
		require.NotContains(t, flags, "--debug", "global flag")
		require.Len(t, flags, 8)

		require.Equal(t, CliFlag{Name: "--role", Aliases: []string{"-r"}, Type: "string", Usage: "account role", Default: `"admin"`, Env: []string{}}, flags["--role"])
		require.Equal(t, []string{"PHOTOPRISM_TEST_MCP_PASSWORD"}, flags["--password"].Env)

		for _, name := range []string{"--password", "--count", "--path", "--generic", "--parsed", "--name-file", "--wrapped"} {
			require.Empty(t, flags[name].Default, name)
		}
	})
	t.Run("NilList", func(t *testing.T) {
		require.Nil(t, buildCliCommands(nil, testGlobalFlags()))
	})
	t.Run("EmptyList", func(t *testing.T) {
		items := buildCliCommands([]*cli.Command{}, nil)
		require.NotNil(t, items)
		require.Empty(t, items)
	})
	t.Run("NilEntries", func(t *testing.T) {
		items := buildCliCommands([]*cli.Command{nil, {Name: "start", Usage: "Starts the server"}}, nil)
		require.Len(t, items, 1)
		require.Equal(t, "photoprism start", items[0].Command)
		require.Equal(t, []CliFlag{}, items[0].Flags)
	})
	t.Run("NoGlobalFlags", func(t *testing.T) {
		items := buildCliCommands(testCommands(), nil)
		require.Len(t, items[1].Flags, 9, "--debug is kept without global flags")
	})
}

// TestBuildCliCommandsOmitsEnvironmentValues sets a sentinel in every environment variable of the
// test commands, applies their flags twice, and checks that the sentinel appears in no output.
func TestBuildCliCommandsOmitsEnvironmentValues(t *testing.T) {
	const sentinel = "TestBuildCliCommandsOmitsEnvironmentValues"

	for _, name := range testEnvVars {
		if strings.HasSuffix(name, "_COUNT") {
			t.Setenv(name, "4242")
		} else {
			t.Setenv(name, sentinel)
		}
	}

	commands := testCommands()
	applyTwice(t, commands)

	// The catalog reports the environment values after a second Apply, so the test can fail.
	add := commands[0].Subcommands[0]
	require.Contains(t, catalog.DescribeFlag(findFlag(add.Flags, "--password")).Default, sentinel)
	require.Equal(t, "4242", catalog.DescribeFlag(findFlag(add.Flags, "--count")).Default)

	server := NewServer(&sdkmcp.Implementation{Name: "photoprism-mcp", Version: "test"}, "ce", Options{Commands: commands, GlobalFlags: testGlobalFlags()})
	session := connectServer(t, server)

	res, err := session.ReadResource(context.Background(), &sdkmcp.ReadResourceParams{URI: cliCommandsURI})
	require.NoError(t, err)
	require.Len(t, res.Contents, 1)
	require.Contains(t, res.Contents[0].Text, "photoprism users add")
	require.NotContains(t, res.Contents[0].Text, sentinel)
	require.NotContains(t, res.Contents[0].Text, "4242")

	call, err := session.CallTool(context.Background(), &sdkmcp.CallToolParams{Name: "list_cli_commands", Arguments: map[string]any{"limit": maxResultLimit}})
	require.NoError(t, err)
	require.False(t, call.IsError)

	raw, err := json.Marshal(call)
	require.NoError(t, err)
	require.Contains(t, string(raw), "photoprism users add")
	require.NotContains(t, string(raw), sentinel)
	require.NotContains(t, string(raw), "4242")
}

// TestCliCommandSources covers unique and ambiguous command paths.
func TestCliCommandSources(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		commands := testCommands()
		sources := cliCommandSources(commands)
		require.Same(t, commands[0].Subcommands[0], sources["photoprism users add"])
		require.Same(t, commands[1].Subcommands[0], sources["photoprism maintenance run"])
	})
	t.Run("Ambiguous", func(t *testing.T) {
		sources := cliCommandSources([]*cli.Command{{Name: "start"}, nil, {Name: "start"}})
		require.Contains(t, sources, "photoprism start")
		require.Nil(t, sources["photoprism start"])
	})
}

// TestNewCliCommand checks that defaults are omitted without a source command.
func TestNewCliCommand(t *testing.T) {
	cmd := catalog.Command{FullName: "photoprism users add", Usage: "Adds", Flags: []catalog.Flag{{Name: "--role", Type: "string", Default: `"admin"`}}}
	src := &cli.Command{Name: "add", Flags: []cli.Flag{&cli.StringFlag{Name: "role", Value: "admin"}}}

	t.Run("Success", func(t *testing.T) {
		result := newCliCommand(cmd, src)
		require.Equal(t, "photoprism users add", result.Command)
		require.Equal(t, `"admin"`, result.Flags[0].Default)
		require.Equal(t, []string{}, result.Flags[0].Aliases)
	})
	t.Run("NoSource", func(t *testing.T) {
		require.Empty(t, newCliCommand(cmd, nil).Flags[0].Default)
	})
	t.Run("UnknownFlag", func(t *testing.T) {
		require.Empty(t, newCliCommand(cmd, &cli.Command{Name: "add"}).Flags[0].Default)
	})
}

// TestFindFlag covers found, missing, and nil flags.
func TestFindFlag(t *testing.T) {
	role := &cli.StringFlag{Name: "role", Aliases: []string{"r"}}
	flags := []cli.Flag{nil, &cli.BoolFlag{Name: "yes"}, role}

	t.Run("Success", func(t *testing.T) {
		require.Same(t, role, findFlag(flags, "--role"))
	})
	t.Run("NotFound", func(t *testing.T) {
		require.Nil(t, findFlag(flags, "--name"))
		require.Nil(t, findFlag(flags, "-r"))
	})
}

// TestFlagDefaultAllowed covers literal, env-backed, file-backed, generic, wrapped, and missing flags.
func TestFlagDefaultAllowed(t *testing.T) {
	t.Run("Literal", func(t *testing.T) {
		require.True(t, flagDefaultAllowed(&cli.StringFlag{Name: "role", Value: "admin"}))
		require.True(t, flagDefaultAllowed(&cli.StringSliceFlag{Name: "tags"}))
		require.True(t, flagDefaultAllowed(&testWrappedFlag{&cli.StringFlag{Name: "wrapped"}}))
		require.True(t, flagDefaultAllowed(&testFieldWrappedFlag{&cli.StringFlag{Name: "wrapped"}}))
	})
	t.Run("Environment", func(t *testing.T) {
		require.False(t, flagDefaultAllowed(&cli.StringFlag{Name: "dsn", EnvVars: []string{"DSN"}}))
		require.False(t, flagDefaultAllowed(&cli.StringSliceFlag{Name: "tags", EnvVars: []string{"TAGS"}}))
		require.False(t, flagDefaultAllowed(&testWrappedFlag{&cli.StringFlag{Name: "wrapped", EnvVars: []string{"WRAPPED"}}}))
		require.False(t, flagDefaultAllowed(&testFieldWrappedFlag{&cli.StringFlag{Name: "wrapped", EnvVars: []string{"WRAPPED"}}}))
	})
	t.Run("File", func(t *testing.T) {
		require.False(t, flagDefaultAllowed(&cli.StringFlag{Name: "secret", FilePath: "/etc/secret"}))
		require.False(t, flagDefaultAllowed(&testWrappedFlag{&cli.StringFlag{Name: "wrapped", FilePath: "/etc/secret"}}))
		require.False(t, flagDefaultAllowed(&testFieldWrappedFlag{&cli.StringFlag{Name: "wrapped", FilePath: "/etc/secret"}}))
	})
	t.Run("Generic", func(t *testing.T) {
		require.False(t, flagDefaultAllowed(&cli.GenericFlag{Name: "generic", Value: &testGeneric{}}))
	})
	t.Run("Missing", func(t *testing.T) {
		require.False(t, flagDefaultAllowed(nil))
	})
}

// TestFlagFilePath covers flags with and without a FilePath field.
func TestFlagFilePath(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		require.Equal(t, "/etc/secret", flagFilePath(&cli.PathFlag{Name: "path", FilePath: "/etc/secret"}))
		require.Equal(t, "/etc/secret", flagFilePath(&testWrappedFlag{&cli.StringFlag{FilePath: "/etc/secret"}}))
	})
	t.Run("Empty", func(t *testing.T) {
		require.Empty(t, flagFilePath(&cli.StringFlag{Name: "name"}))
		require.Empty(t, flagFilePath(&testWrappedFlag{}), "nil embedded flag")
		require.Empty(t, flagFilePath(nil))
		require.Empty(t, flagFilePath((*cli.StringFlag)(nil)))
	})
}

// TestCloneStrings checks that the result is a non-nil copy.
func TestCloneStrings(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		s := []string{"a", "b"}
		c := cloneStrings(s)
		require.Equal(t, s, c)
		c[0] = "x"
		require.Equal(t, "a", s[0])
	})
	t.Run("Nil", func(t *testing.T) {
		require.Equal(t, []string{}, cloneStrings(nil))
	})
}
