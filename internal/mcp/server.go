package mcp

import sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

// ToolNames lists the read-only tools registered on the MCP server in the
// order they are added. Callers use it for startup logging and tests that
// need to assert the registered tool surface.
var ToolNames = []string{
	"list_config_keys",
	"find_search_filters",
	"list_file_formats",
	"list_metadata_fields",
	"list_thumbnail_sizes",
}

// ResourceURIs lists the read-only resource URIs registered on the MCP
// server in the order they are added. Callers use it for startup logging
// and tests that need to assert the registered resource surface.
var ResourceURIs = []string{
	configOptionsURI,
	searchFiltersURI,
	fileFormatsURI,
	metadataFieldsURI,
	thumbnailSizesURI,
}

// NewServer returns a read-only MCP server with the tools and resources in ToolNames and ResourceURIs.
// The implementation is reported to clients as the server's identity; currentEdition is normalized
// and embedded in the resource payloads and tool responses.
func NewServer(implementation *sdkmcp.Implementation, currentEdition string) *sdkmcp.Server {
	server := sdkmcp.NewServer(implementation, nil)
	data := NewDataset(currentEdition)

	registerResources(server, data)
	registerTools(server, data)

	return server
}
