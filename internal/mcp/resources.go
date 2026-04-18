package mcp

import (
	"context"
	"encoding/json"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/photoprism/photoprism/pkg/http/header"
)

// registerResources registers every read-only resource exposed by the
// server against the shared *Dataset. The order matches ResourceURIs so
// the startup log and the SDK's resources/list response stay in sync.
func registerResources(server *sdkmcp.Server, data *Dataset) {
	server.AddResource(&sdkmcp.Resource{
		URI:         configOptionsURI,
		Name:        "config-options",
		Title:       "PhotoPrism Config Options",
		Description: "Read-only config options derived from the existing config report.",
		MIMEType:    header.ContentTypeJson,
	}, func(_ context.Context, req *sdkmcp.ReadResourceRequest) (*sdkmcp.ReadResourceResult, error) {
		return newResourceResult(req.Params.URI, ConfigOptionsResource{
			Edition: data.CurrentEdition,
			Items:   data.ConfigOptions,
		})
	})

	server.AddResource(&sdkmcp.Resource{
		URI:         searchFiltersURI,
		Name:        "search-filters",
		Title:       "PhotoPrism Search Filters",
		Description: "Read-only search filter reference derived from the existing search report.",
		MIMEType:    header.ContentTypeJson,
	}, func(_ context.Context, req *sdkmcp.ReadResourceRequest) (*sdkmcp.ReadResourceResult, error) {
		return newResourceResult(req.Params.URI, SearchFiltersResource{
			Edition: data.CurrentEdition,
			Items:   data.SearchFilters,
		})
	})
}

// newResourceResult marshals payload to indented JSON and wraps it in an
// MCP ReadResourceResult with the given URI and header.ContentTypeJson as
// the advertised MIME type. Returns an error if JSON marshalling fails.
func newResourceResult(uri string, payload any) (*sdkmcp.ReadResourceResult, error) {
	body, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return nil, err
	}

	return &sdkmcp.ReadResourceResult{
		Contents: []*sdkmcp.ResourceContents{{
			URI:      uri,
			MIMEType: header.ContentTypeJson,
			Text:     string(body),
		}},
	}, nil
}
