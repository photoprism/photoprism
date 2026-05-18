package commands

import (
	"context"
	"errors"
	"log/slog"
	"os"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/mcp"
)

// MCPCommands configures the Model Context Protocol (MCP) command group.
var MCPCommands = &cli.Command{
	Name:  "mcp",
	Usage: "Shows the Model Context Protocol (MCP) server subcommands",
	Subcommands: []*cli.Command{
		MCPServeCommand,
	},
}

// MCPServeCommand starts the MCP server over the stdio transport.
var MCPServeCommand = &cli.Command{
	Name:   "serve",
	Usage:  "Starts the internal MCP server via stdio for development and testing",
	Action: mcpServeAction,
}

// mcpServeAction starts the MCP server over the stdio transport,
// writing a startup line to stderr so the JSON-RPC stream on stdout
// stays clean for the MCP client. The action loads the active config
// and refuses to start when DisableMCP is set, mirroring how the
// HTTP transport at /api/v1/mcp is gated; pass --disable-mcp=false on
// the command line to override the operator setting for ad-hoc
// development runs.
func mcpServeAction(ctx *cli.Context) error {
	conf, err := InitConfig(ctx)

	if err != nil {
		return err
	}

	if err = mcpDisabledExit(conf); err != nil {
		return err
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	logger.Info("starting mcp server",
		"transport", "stdio",
		"tools", len(mcp.ToolNames),
		"resources", len(mcp.ResourceURIs),
	)

	return runMCPServer(context.Background(), ctx, &sdkmcp.StdioTransport{})
}

// mcpDisabledExit returns a cli.ExitCoder error when MCP is disabled
// by config, pointing the caller at --disable-mcp=false for ad-hoc
// development runs. It returns nil when MCP is enabled.
func mcpDisabledExit(conf *config.Config) error {
	if conf == nil || !conf.DisableMCP() {
		return nil
	}

	return cli.Exit(errors.New("mcp serve disabled by config; pass --disable-mcp=false to override"), 1)
}

// runMCPServer builds an MCP server from the CLI app metadata (Version,
// Edition) and runs it over the given transport until ctx is canceled or
// the transport closes. The transport is a parameter so tests can
// substitute an in-memory transport for the stdio one mcpServeAction uses
// in production.
func runMCPServer(ctx context.Context, appCtx *cli.Context, transport sdkmcp.Transport) error {
	implementation := &sdkmcp.Implementation{
		Name:    "photoprism-mcp",
		Version: mcpAppMetadata(appCtx, "Version", "development"),
	}
	edition := mcpAppMetadata(appCtx, "Edition", "unknown")

	return mcp.NewServer(implementation, edition).Run(ctx, transport)
}

// mcpAppMetadata returns the named string entry from the CLI app metadata,
// falling back to the supplied default if the key is missing, the value is
// not a string, or the value is an empty string.
func mcpAppMetadata(ctx *cli.Context, key, fallback string) string {
	if value, ok := ctx.App.Metadata[key].(string); ok && value != "" {
		return value
	}

	return fallback
}
