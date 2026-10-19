package commands

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"reflect"
	"strings"
	"testing"
	"time"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/commands/catalog"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/mcp"
	"github.com/photoprism/photoprism/internal/photoprism/get"
)

// TestMCPCommandRegistered ensures the MCP command is present in the CLI catalog.
func TestMCPCommandRegistered(t *testing.T) {
	found := false

	for _, cmd := range PhotoPrism {
		if cmd.Name == "mcp" {
			found = true
			break
		}
	}

	if !found {
		t.Fatal("expected mcp command to be registered")
	}
}

// TestShowCommandsIncludesMCP ensures the command catalog exposes the MCP command.
func TestShowCommandsIncludesMCP(t *testing.T) {
	out, err := RunWithTestContext(ShowCommandsCommand, []string{"commands"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "## photoprism mcp") {
		t.Fatalf("expected MCP heading in output\n%s", out[:min(400, len(out))])
	}
}

// TestRunMCPServerOverInMemoryTransport exercises the full Action-side wiring
// (implementation metadata, server construction, Run lifecycle) against an
// in-memory MCP client. It is the stdio-path counterpart to the in-process
// tests in internal/mcp/server_test.go and catches regressions in the CLI
// glue that the SDK's unit tests cannot reach.
func TestRunMCPServerOverInMemoryTransport(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	serverTransport, clientTransport := sdkmcp.NewInMemoryTransports()

	appCtx := NewTestContext(nil)
	appCtx.App.Metadata["Version"] = "42.1.0"
	appCtx.App.Metadata["Edition"] = "pro"

	serverDone := make(chan error, 1)
	go func() { serverDone <- runMCPServer(ctx, appCtx, serverTransport) }()

	client := sdkmcp.NewClient(&sdkmcp.Implementation{Name: "photoprism-mcp-test", Version: "1.0"}, nil)

	session, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)

	info := session.InitializeResult()
	require.NotNil(t, info, "client should receive an InitializeResult from the server")
	require.Equal(t, "photoprism-mcp", info.ServerInfo.Name)
	require.Equal(t, "42.1.0", info.ServerInfo.Version, "Version must be sourced from App.Metadata")

	tools, err := session.ListTools(ctx, nil)
	require.NoError(t, err)
	names := make([]string, 0, len(tools.Tools))
	for _, tool := range tools.Tools {
		names = append(names, tool.Name)
	}
	require.ElementsMatch(t, mcp.ToolNames, names, "server should advertise every read-only tool over stdio-equivalent transport")

	for name, args := range map[string]map[string]any{
		"list_file_formats":    {"type": "raw", "limit": 5},
		"list_metadata_fields": {"query": "exif", "limit": 5},
		"list_thumbnail_sizes": {"type": "video"},
	} {
		res, callErr := session.CallTool(ctx, &sdkmcp.CallToolParams{Name: name, Arguments: args})
		require.NoError(t, callErr, name)
		require.False(t, res.IsError, name)
		out, ok := res.StructuredContent.(map[string]any)
		require.True(t, ok, name)
		require.NotEmpty(t, out["items"], name)
	}

	res, err := session.CallTool(ctx, &sdkmcp.CallToolParams{Name: "list_cli_commands", Arguments: map[string]any{"query": "mcp serve"}})
	require.NoError(t, err)
	require.False(t, res.IsError)
	raw, err := json.Marshal(res.StructuredContent)
	require.NoError(t, err)
	require.Contains(t, string(raw), `"command":"photoprism mcp serve"`)

	// The resource lists the commands built from the app's command list.
	resource, err := session.ReadResource(ctx, &sdkmcp.ReadResourceParams{URI: "photoprism://cli-commands"})
	require.NoError(t, err)
	require.Len(t, resource.Contents, 1)

	var payload mcp.CliCommandsResource
	require.NoError(t, json.Unmarshal([]byte(resource.Contents[0].Text), &payload))
	require.Equal(t, mcp.NewDataset("pro", mcp.Options{Commands: appCtx.App.Commands}).CliCommands, payload.Items)

	require.NoError(t, session.Close())
	cancel()

	select {
	case err := <-serverDone:
		if err != nil && !errors.Is(err, context.Canceled) {
			t.Errorf("runMCPServer returned unexpected error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("runMCPServer did not shut down within 5 seconds")
	}
}

// TestMCPDisabledExit covers the gate helper that mcpServeAction calls
// before starting the stdio transport: DisableMCP=true must return a
// cli.ExitCoder error whose message points at --disable-mcp=false as
// the documented override; DisableMCP=false must return nil so the
// server can start; nil config must be tolerated.
func TestMCPDisabledExit(t *testing.T) {
	conf := get.Config()
	require.NotNil(t, conf, "test harness must provide a shared config")

	original := conf.Options().DisableMCP
	t.Cleanup(func() { conf.Options().DisableMCP = original })

	t.Run("DisableMCPTrue_Refuses", func(t *testing.T) {
		conf.Options().DisableMCP = true

		err := mcpDisabledExit(conf)
		require.Error(t, err)

		coder, ok := err.(cli.ExitCoder)
		require.True(t, ok, "expected cli.ExitCoder, got %T", err)
		require.Equal(t, 1, coder.ExitCode())
		require.Contains(t, err.Error(), "--disable-mcp=false")
	})
	t.Run("DisableMCPFalse_Allows", func(t *testing.T) {
		conf.Options().DisableMCP = false

		require.NoError(t, mcpDisabledExit(conf))
	})
	t.Run("NilConfig_Allows", func(t *testing.T) {
		require.NoError(t, mcpDisabledExit(nil))
	})
}

// TestMCPServeRefusesWhenDisabled drives the full mcp serve action through
// RunWithTestContext to confirm the gate trips end-to-end before the stdio
// transport is started. The DisableMCP=false branch is not exercised here
// because mcpServeAction would block on stdin once it reaches the transport;
// the gate-level coverage in TestMCPDisabledExit is sufficient to assert the
// pre-transport branch.
func TestMCPServeRefusesWhenDisabled(t *testing.T) {
	conf := get.Config()
	require.NotNil(t, conf, "test harness must provide a shared config")

	original := conf.Options().DisableMCP
	t.Cleanup(func() { conf.Options().DisableMCP = original })

	conf.Options().DisableMCP = true

	_, err := RunWithTestContext(MCPServeCommand, []string{"serve"})
	require.Error(t, err)

	coder, ok := err.(cli.ExitCoder)
	require.True(t, ok, "expected cli.ExitCoder, got %T", err)
	require.Equal(t, 1, coder.ExitCode())
	require.Contains(t, err.Error(), "--disable-mcp=false")
}

// TestMCPAppMetadata covers the happy path and every fallback branch of
// mcpAppMetadata: missing key, non-string value, empty-string value, and
// nil metadata map.
func TestMCPAppMetadata(t *testing.T) {
	cases := []struct {
		name     string
		metadata map[string]any
		key      string
		fallback string
		want     string
	}{
		{"ValidString", map[string]any{"Version": "1.2.3"}, "Version", "development", "1.2.3"},
		{"MissingKey", map[string]any{"Edition": "ce"}, "Version", "development", "development"},
		{"NonStringValue", map[string]any{"Version": 42}, "Version", "development", "development"},
		{"EmptyString", map[string]any{"Version": ""}, "Version", "development", "development"},
		{"NilMetadata", nil, "Version", "development", "development"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			app := &cli.App{Metadata: tc.metadata}
			ctx := cli.NewContext(app, nil, nil)
			require.Equal(t, tc.want, mcpAppMetadata(ctx, tc.key, tc.fallback))
		})
	}
}

// TestMCPCliCommandsMatchShowCommands checks that list_cli_commands lists the same commands and flags
// as `photoprism show commands --json` for the CE command list.
func TestMCPCliCommandsMatchShowCommands(t *testing.T) {
	out, err := RunWithTestContext(ShowCommandsCommand, []string{"commands", "--json"})
	require.NoError(t, err)

	var report struct {
		Commands []catalog.Command `json:"commands"`
	}
	require.NoError(t, json.Unmarshal([]byte(out), &report))
	require.NotEmpty(t, report.Commands)

	ctx := NewTestContext(nil)
	data := mcp.NewDataset("ce", mcp.Options{Commands: ctx.App.Commands})
	require.Len(t, data.CliCommands, len(report.Commands))

	flagCount := 0

	for i, cmd := range data.CliCommands {
		expected := report.Commands[i]
		require.Equal(t, expected.FullName, cmd.Command)
		require.Equal(t, expected.Usage, cmd.Usage, cmd.Command)
		require.Len(t, cmd.Flags, len(expected.Flags), cmd.Command)

		for j, f := range cmd.Flags {
			require.Equal(t, expected.Flags[j].Name, f.Name, cmd.Command)
		}

		flagCount += len(cmd.Flags)
	}

	require.Positive(t, flagCount)
	require.Equal(t, "photoprism mcp serve", findCliCommand(t, data.CliCommands, "photoprism mcp serve").Command)
	require.Equal(t, `"admin"`, findCliFlag(t, findCliCommand(t, data.CliCommands, "photoprism users add"), "--role").Default)
}

// TestMCPCliCommandsOmitEnvironmentValues sets a sentinel in the environment variable of every CE
// command flag and every secret config flag, applies the command flags twice, and checks that no
// list_cli_commands row carries the value or a default for these flags.
func TestMCPCliCommandsOmitEnvironmentValues(t *testing.T) {
	const sentinel = "TestMCPCliCommandsOmitEnvironmentValues"

	for _, f := range config.Flags {
		if f.Secret {
			t.Setenv(f.EnvVar(), sentinel)
		}
	}

	envFlags := applyEnvFlagsTwice(t, PhotoPrism, sentinel)
	require.Contains(t, envFlags, "photoprism database transfer --target-password")
	require.Contains(t, catalog.DescribeFlag(envFlags["photoprism database transfer --target-password"]).Default, sentinel,
		"a second Apply must report the environment value, or this test cannot fail")

	data := mcp.NewDataset("ce", mcp.Options{Commands: PhotoPrism})

	raw, err := json.Marshal(data.CliCommands)
	require.NoError(t, err)
	require.NotContains(t, string(raw), sentinel)

	for _, cmd := range data.CliCommands {
		for _, f := range cmd.Flags {
			if _, ok := envFlags[cmd.Command+" "+f.Name]; ok {
				require.Empty(t, f.Default, "%s %s", cmd.Command, f.Name)
				require.NotEmpty(t, f.Env, "%s %s", cmd.Command, f.Name)
			}
		}
	}
}

// applyEnvFlagsTwice sets the environment variables of every flag in the command tree, applies each
// of these flags to two fresh flag sets, and restores the flag definitions when the test ends. It
// returns the flags by command path and flag name.
func applyEnvFlagsTwice(t *testing.T, commands []*cli.Command, sentinel string) map[string]cli.Flag {
	t.Helper()

	result := make(map[string]cli.Flag)

	var walk func(cmds []*cli.Command, parent string)

	walk = func(cmds []*cli.Command, parent string) {
		for _, c := range cmds {
			if c == nil {
				continue
			}

			name := parent + " " + c.Name

			for _, f := range c.Flags {
				e, ok := f.(interface{ GetEnvVars() []string })

				if !ok || len(e.GetEnvVars()) == 0 {
					continue
				}

				value := sentinel

				switch catalog.DescribeFlag(f).Type {
				case "bool":
					value = "true"
				case "int", "int64", "uint", "uint64", "float64":
					value = "4242"
				case "duration":
					value = "4242s"
				}

				for _, env := range e.GetEnvVars() {
					t.Setenv(env, value)
				}

				t.Cleanup(restoreFlag(f))

				for range 2 {
					require.NoError(t, f.Apply(flag.NewFlagSet("test", flag.ContinueOnError)), name)
				}

				result[name+" "+catalog.DescribeFlag(f).Name] = f
			}

			walk(c.Subcommands, name)
		}
	}

	walk(commands, "photoprism")

	return result
}

// restoreFlag returns a function that restores the flag definition, including the state that Apply sets.
func restoreFlag(f cli.Flag) func() {
	if w, ok := f.(interface{ Unwrap() cli.Flag }); ok {
		f = w.Unwrap()
	}

	v := reflect.ValueOf(f).Elem()
	saved := reflect.New(v.Type()).Elem()
	saved.Set(v)

	return func() { v.Set(saved) }
}

// findCliCommand returns the command with the specified path, failing the test if there is none.
func findCliCommand(t *testing.T, commands []mcp.CliCommand, path string) mcp.CliCommand {
	t.Helper()

	for _, cmd := range commands {
		if cmd.Command == path {
			return cmd
		}
	}

	require.FailNow(t, "command not found", path)

	return mcp.CliCommand{}
}

// findCliFlag returns the flag with the specified name, failing the test if there is none.
func findCliFlag(t *testing.T, cmd mcp.CliCommand, name string) mcp.CliFlag {
	t.Helper()

	for _, f := range cmd.Flags {
		if f.Name == name {
			return f
		}
	}

	require.FailNow(t, "flag not found", "%s %s", cmd.Command, name)

	return mcp.CliFlag{}
}
