package mcp

import (
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/config/customize"
)

// ToolNames lists the reference tools in the order they are added, which both the reference and the library
// server register. Callers use it for startup logging and tests that assert the registered tool surface.
var ToolNames = []string{
	"list_config_keys",
	"find_search_filters",
	"list_cli_commands",
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
	cliCommandsURI,
	fileFormatsURI,
	metadataFieldsURI,
	thumbnailSizesURI,
}

// Options contains the reference data and settings that only the caller can provide.
type Options struct {
	// Commands is the CLI command list of the running edition.
	Commands []*cli.Command
	// Settings returns the current settings, whose features change search results like in the web app.
	Settings func() *customize.Settings
}

// features returns the current feature settings, or the defaults if no settings are available.
func (opt Options) features() customize.FeatureSettings {
	if opt.Settings != nil {
		if settings := opt.Settings(); settings != nil {
			return settings.Features
		}
	}

	return customize.NewDefaultSettings().Features
}

// LibraryToolNames lists the tools that read library data. They are registered only on the library server, whose
// middleware checks each call against the tool's declaration.
var LibraryToolNames = []string{
	"search_media",
	"get_media",
	"list_collections",
	"get_collection",
	"list_people",
	"list_labels",
}

// NewServer returns a read-only MCP server with the tools and resources in ToolNames and ResourceURIs.
// The implementation is reported to clients as the server's identity; currentEdition is normalized
// and embedded in the resource payloads and tool responses.
func NewServer(implementation *sdkmcp.Implementation, currentEdition string, opt Options) *sdkmcp.Server {
	server := sdkmcp.NewServer(implementation, nil)
	data := NewDataset(currentEdition, opt)

	registerResources(server, data)
	registerTools(server, data)

	return server
}

// NewLibraryServer returns the MCP server for authenticated sessions, with the tools of NewServer and the tools
// in LibraryToolNames. It checks every tools/call against the tool's declaration and the session in the token
// info of the request, and lists only the tools the caller's role and account may use.
func NewLibraryServer(implementation *sdkmcp.Implementation, currentEdition string, opt Options) *sdkmcp.Server {
	server := newLibraryServer(implementation, currentEdition, opt)
	server.AddReceivingMiddleware(authorizeTools)

	return server
}

// newLibraryServer registers the tools and resources of the library server without its middleware.
func newLibraryServer(implementation *sdkmcp.Implementation, currentEdition string, opt Options) *sdkmcp.Server {
	server := NewServer(implementation, currentEdition, opt)
	registerLibraryTools(server, opt)

	return server
}

// registerLibraryTools registers the tools in LibraryToolNames.
func registerLibraryTools(server *sdkmcp.Server, opt Options) {
	registerSearchMedia(server, opt)
	registerGetMedia(server)
	registerCollectionTools(server)
	registerPeopleAndLabelTools(server)
}
