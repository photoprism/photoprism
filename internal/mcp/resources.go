package mcp

import (
	"context"
	"encoding/json"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/photoprism/photoprism/pkg/http/header"
)

// registerResources registers the resources in ResourceURIs against the shared *Dataset.
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

	server.AddResource(&sdkmcp.Resource{
		URI:         cliCommandsURI,
		Name:        "cli-commands",
		Title:       "PhotoPrism CLI Commands",
		Description: "Read-only list of the CLI commands of this edition derived from the existing command catalog.",
		MIMEType:    header.ContentTypeJson,
	}, func(_ context.Context, req *sdkmcp.ReadResourceRequest) (*sdkmcp.ReadResourceResult, error) {
		items := data.CliCommands

		if items == nil {
			items = []CliCommand{}
		}

		return newResourceResult(req.Params.URI, CliCommandsResource{
			Edition: data.CurrentEdition,
			Items:   items,
		})
	})

	server.AddResource(&sdkmcp.Resource{
		URI:         fileFormatsURI,
		Name:        "file-formats",
		Title:       "PhotoPrism File Formats",
		Description: "Read-only list of supported file formats derived from the existing file format report.",
		MIMEType:    header.ContentTypeJson,
	}, func(_ context.Context, req *sdkmcp.ReadResourceRequest) (*sdkmcp.ReadResourceResult, error) {
		return newResourceResult(req.Params.URI, FileFormatsResource{
			Edition: data.CurrentEdition,
			Items:   data.FileFormats,
		})
	})

	server.AddResource(&sdkmcp.Resource{
		URI:         metadataFieldsURI,
		Name:        "metadata-fields",
		Title:       "PhotoPrism Metadata Fields",
		Description: "Read-only list of supported metadata fields and tags derived from the existing metadata report.",
		MIMEType:    header.ContentTypeJson,
	}, func(_ context.Context, req *sdkmcp.ReadResourceRequest) (*sdkmcp.ReadResourceResult, error) {
		return newResourceResult(req.Params.URI, MetadataFieldsResource{
			Edition: data.CurrentEdition,
			Items:   data.MetadataFields,
			Docs:    data.MetadataDocs,
		})
	})

	server.AddResource(&sdkmcp.Resource{
		URI:         thumbnailSizesURI,
		Name:        "thumbnail-sizes",
		Title:       "PhotoPrism Thumbnail Sizes",
		Description: "Read-only list of thumbnail and video sizes derived from the existing thumbnail size report.",
		MIMEType:    header.ContentTypeJson,
	}, func(_ context.Context, req *sdkmcp.ReadResourceRequest) (*sdkmcp.ReadResourceResult, error) {
		return newResourceResult(req.Params.URI, ThumbnailSizesResource{
			Edition: data.CurrentEdition,
			Items:   data.ThumbnailSizes,
		})
	})
}

// newResourceResult marshals payload to indented JSON and wraps it in an
// MCP ReadResourceResult with the given URI and header.ContentTypeJson as
// the advertised MIME type. Returns an error if JSON marshaling fails.
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
