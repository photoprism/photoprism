package mcp

import (
	"context"
	"encoding/json"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

const jsonMIMEType = "application/json"

// registerResources adds the static read-only resources to the server.
func registerResources(server *sdkmcp.Server, data *Dataset) {
	server.AddResource(&sdkmcp.Resource{
		URI:         configOptionsURI,
		Name:        "config-options",
		Title:       "PhotoPrism Config Options",
		Description: "Read-only config options derived from the existing config report.",
		MIMEType:    jsonMIMEType,
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
		MIMEType:    jsonMIMEType,
	}, func(_ context.Context, req *sdkmcp.ReadResourceRequest) (*sdkmcp.ReadResourceResult, error) {
		return newResourceResult(req.Params.URI, SearchFiltersResource{
			Edition: data.CurrentEdition,
			Items:   data.SearchFilters,
		})
	})
}

// newResourceResult marshals a JSON payload into an MCP read-resource response.
func newResourceResult(uri string, payload any) (*sdkmcp.ReadResourceResult, error) {
	body, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return nil, err
	}

	return &sdkmcp.ReadResourceResult{
		Contents: []*sdkmcp.ResourceContents{{
			URI:      uri,
			MIMEType: jsonMIMEType,
			Text:     string(body),
		}},
	}, nil
}
