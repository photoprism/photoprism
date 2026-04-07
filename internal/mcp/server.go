package mcp

import sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

// NewServer returns a read-only MCP server for local internal evaluation.
func NewServer(implementation *sdkmcp.Implementation, currentEdition string) *sdkmcp.Server {
	server := sdkmcp.NewServer(implementation, nil)
	data := NewDataset(currentEdition)

	registerResources(server, data)
	registerTools(server, data)

	return server
}
