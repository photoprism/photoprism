package commands

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

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
	require.Len(t, tools.Tools, 2, "server should advertise the two read-only tools over stdio-equivalent transport")

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
