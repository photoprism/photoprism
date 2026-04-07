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
		Version: mcpVersion(ctx),
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))

	logger.Info("starting mcp prototype", "transport", "stdio", "tools", 2, "resources", 2)

	return internalmcp.NewServer(implementation, mcpEdition(ctx)).Run(context.Background(), &sdkmcp.StdioTransport{})
}

// mcpEdition returns the current build edition for MCP responses.
func mcpEdition(ctx *cli.Context) string {
	if ctx != nil && ctx.App != nil && ctx.App.Metadata != nil {
		if edition, ok := ctx.App.Metadata["Edition"].(string); ok && edition != "" {
			return edition
		}
	}

	return "unknown"
}

// mcpVersion returns the current application version for MCP metadata.
func mcpVersion(ctx *cli.Context) string {
	if ctx != nil && ctx.App != nil && ctx.App.Metadata != nil {
		if version, ok := ctx.App.Metadata["Version"].(string); ok && version != "" {
			return version
		}
	}

	return "development"
}
