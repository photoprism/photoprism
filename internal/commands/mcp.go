package commands

import (
	"context"
	"log/slog"
	"os"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/urfave/cli/v2"

	internalmcp "github.com/photoprism/photoprism/internal/mcp"
)

// MCPCommands configures the MCP prototype command group.
var MCPCommands = &cli.Command{
	Name:  "mcp",
	Usage: "Runs the internal read-only MCP prototype",
	Subcommands: []*cli.Command{
		MCPServeCommand,
	},
}

// MCPServeCommand starts the MCP prototype over stdio.
var MCPServeCommand = &cli.Command{
	Name:   "serve",
	Usage:  "Starts the internal read-only MCP prototype over stdio",
	Action: mcpServeAction,
}

// mcpServeAction starts the MCP server using the stdio transport.
func mcpServeAction(ctx *cli.Context) error {
	implementation := &sdkmcp.Implementation{
		Name:    "photoprism-mcp",
		Version: mcpAppMetadata(ctx, "Version", "development"),
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))

	logger.Info("starting mcp prototype", "transport", "stdio", "tools", 2, "resources", 2)

	edition := mcpAppMetadata(ctx, "Edition", "unknown")

	return internalmcp.NewServer(implementation, edition).Run(context.Background(), &sdkmcp.StdioTransport{})
}

// mcpAppMetadata returns the named string entry from the CLI app metadata,
// falling back to the supplied default if it is missing or not a string.
func mcpAppMetadata(ctx *cli.Context, key, fallback string) string {
	if value, ok := ctx.App.Metadata[key].(string); ok && value != "" {
		return value
	}

	return fallback
}
