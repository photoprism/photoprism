package mcp

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

// connectTestClient connects an in-memory client to an MCP server with the synthetic test commands.
func connectTestClient(t *testing.T) *sdkmcp.ClientSession {
	t.Helper()

	return connectServer(t, NewServer(&sdkmcp.Implementation{Name: "photoprism-mcp", Version: "test"}, "ce", Options{Commands: testCommands()}))
}

// connectServer connects an in-memory client to the specified MCP server.
func connectServer(t *testing.T, server *sdkmcp.Server) *sdkmcp.ClientSession {
	t.Helper()

	ctx := context.Background()
	client := sdkmcp.NewClient(&sdkmcp.Implementation{Name: "photoprism-mcp-test", Version: "test"}, nil)
	serverTransport, clientTransport := sdkmcp.NewInMemoryTransports()

	_, err := server.Connect(ctx, serverTransport, nil)
	require.NoError(t, err)

	session, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, session.Close())
	})

	return session
}

// TestNewServerResources lists and reads the registered MCP resources,
// asserting that exactly the URIs in ResourceURIs are exposed and return
// the expected JSON payloads.
func TestNewServerResources(t *testing.T) {
	ctx := context.Background()
	session := connectTestClient(t)

	resources, err := session.ListResources(ctx, nil)
	require.NoError(t, err)
	uris := make([]string, 0, len(resources.Resources))
	for _, resource := range resources.Resources {
		uris = append(uris, resource.URI)
	}
	require.ElementsMatch(t, ResourceURIs, uris)

	configResource, err := session.ReadResource(ctx, &sdkmcp.ReadResourceParams{URI: configOptionsURI})
	require.NoError(t, err)
	require.Len(t, configResource.Contents, 1)

	var configPayload ConfigOptionsResource
	require.NoError(t, json.Unmarshal([]byte(configResource.Contents[0].Text), &configPayload))
	require.NotEmpty(t, configPayload.Items)
	require.Equal(t, "ce", configPayload.Edition)

	filterResource, err := session.ReadResource(ctx, &sdkmcp.ReadResourceParams{URI: searchFiltersURI})
	require.NoError(t, err)
	require.Len(t, filterResource.Contents, 1)

	var filterPayload SearchFiltersResource
	require.NoError(t, json.Unmarshal([]byte(filterResource.Contents[0].Text), &filterPayload))
	require.NotEmpty(t, filterPayload.Items)
	require.Equal(t, "ce", filterPayload.Edition)

	var commandsPayload CliCommandsResource
	readResourcePayload(t, session, cliCommandsURI, &commandsPayload)
	require.Equal(t, buildCliCommands(testCommands()), commandsPayload.Items)
	require.Equal(t, "ce", commandsPayload.Edition)

	var formatsPayload FileFormatsResource
	readResourcePayload(t, session, fileFormatsURI, &formatsPayload)
	require.Equal(t, buildFileFormats(), formatsPayload.Items)
	require.Equal(t, "ce", formatsPayload.Edition)

	var metadataPayload MetadataFieldsResource
	readResourcePayload(t, session, metadataFieldsURI, &metadataPayload)
	require.Equal(t, buildMetadataFields(), metadataPayload.Items)
	require.Equal(t, buildMetadataDocs(), metadataPayload.Docs)
	require.Equal(t, "ce", metadataPayload.Edition)

	var sizesPayload ThumbnailSizesResource
	readResourcePayload(t, session, thumbnailSizesURI, &sizesPayload)
	require.Equal(t, buildThumbnailSizes(), sizesPayload.Items)
	require.Equal(t, "ce", sizesPayload.Edition)
}

// readResourcePayload reads the resource with the specified URI and decodes its JSON payload.
func readResourcePayload(t *testing.T, session *sdkmcp.ClientSession, uri string, payload any) {
	t.Helper()

	res, err := session.ReadResource(context.Background(), &sdkmcp.ReadResourceParams{URI: uri})
	require.NoError(t, err)
	require.Len(t, res.Contents, 1)
	require.Equal(t, uri, res.Contents[0].URI)
	require.NoError(t, json.Unmarshal([]byte(res.Contents[0].Text), payload))
}

// TestNewServerToolList asserts that the server advertises exactly the
// tools listed in ToolNames and that every tool carries a JSON schema
// and the read-only annotation.
func TestNewServerToolList(t *testing.T) {
	ctx := context.Background()
	session := connectTestClient(t)

	tools, err := session.ListTools(ctx, nil)
	require.NoError(t, err)
	require.Len(t, tools.Tools, len(ToolNames))

	names := make([]string, 0, len(tools.Tools))
	for _, tool := range tools.Tools {
		names = append(names, tool.Name)
		require.NotNil(t, tool.InputSchema)
		require.NotNil(t, tool.Annotations, tool.Name)
		require.True(t, tool.Annotations.ReadOnlyHint, tool.Name)
	}
	for _, expected := range ToolNames {
		require.Contains(t, names, expected)
	}
}

// TestListConfigKeysTool exercises successful, empty, and invalid tool calls.
func TestListConfigKeysTool(t *testing.T) {
	ctx := context.Background()
	session := connectTestClient(t)

	t.Run("Success", func(t *testing.T) {
		res, err := session.CallTool(ctx, &sdkmcp.CallToolParams{
			Name: "list_config_keys",
			Arguments: map[string]any{
				"query": "http",
				"limit": 5,
			},
		})
		require.NoError(t, err)
		require.False(t, res.IsError)
		require.NotNil(t, res.StructuredContent)

		raw, err := json.Marshal(res.StructuredContent)
		require.NoError(t, err)

		var out ListConfigKeysOutput
		require.NoError(t, json.Unmarshal(raw, &out))
		require.NotEmpty(t, out.Matches)
		require.LessOrEqual(t, len(out.Matches), 5)
		require.Equal(t, "all", out.EditionApplied)
	})

	t.Run("NoResults", func(t *testing.T) {
		res, err := session.CallTool(ctx, &sdkmcp.CallToolParams{
			Name: "list_config_keys",
			Arguments: map[string]any{
				"query": "definitely-no-match",
			},
		})
		require.NoError(t, err)
		require.False(t, res.IsError)

		raw, err := json.Marshal(res.StructuredContent)
		require.NoError(t, err)

		var out ListConfigKeysOutput
		require.NoError(t, json.Unmarshal(raw, &out))
		require.Empty(t, out.Matches)
		require.Zero(t, out.TotalMatches)
	})

	t.Run("InvalidInput", func(t *testing.T) {
		res, err := session.CallTool(ctx, &sdkmcp.CallToolParams{
			Name: "list_config_keys",
			Arguments: map[string]any{
				"edition": "enterprise",
			},
		})
		require.NoError(t, err)
		require.True(t, res.IsError)
		require.Len(t, res.Content, 1)

		text, ok := res.Content[0].(*sdkmcp.TextContent)
		require.True(t, ok)
		require.Contains(t, text.Text, "edition must be one of ce, plus, pro, portal, or all")
	})

	t.Run("RejectUnknownEdition", func(t *testing.T) {
		res, err := session.CallTool(ctx, &sdkmcp.CallToolParams{
			Name: "list_config_keys",
			Arguments: map[string]any{
				"edition": "unknown",
			},
		})
		require.NoError(t, err)
		require.True(t, res.IsError)
	})
}

// TestListConfigKeysValidation exercises edge cases for input validation.
func TestListConfigKeysValidation(t *testing.T) {
	ctx := context.Background()
	session := connectTestClient(t)

	t.Run("QueryTooLong", func(t *testing.T) {
		longQuery := strings.Repeat("a", maxQueryLength+1)

		res, err := session.CallTool(ctx, &sdkmcp.CallToolParams{
			Name: "list_config_keys",
			Arguments: map[string]any{
				"query": longQuery,
			},
		})
		require.NoError(t, err)
		require.True(t, res.IsError)
	})

	t.Run("WhitespaceQuery", func(t *testing.T) {
		res, err := session.CallTool(ctx, &sdkmcp.CallToolParams{
			Name: "list_config_keys",
			Arguments: map[string]any{
				"query": "   ",
			},
		})
		require.NoError(t, err)
		require.False(t, res.IsError)

		raw, err := json.Marshal(res.StructuredContent)
		require.NoError(t, err)

		var out ListConfigKeysOutput
		require.NoError(t, json.Unmarshal(raw, &out))
		require.NotEmpty(t, out.Matches, "whitespace-only query should match all")
	})

	t.Run("UnicodeQuery", func(t *testing.T) {
		res, err := session.CallTool(ctx, &sdkmcp.CallToolParams{
			Name: "list_config_keys",
			Arguments: map[string]any{
				"query": "日本語",
			},
		})
		require.NoError(t, err)
		require.False(t, res.IsError)
	})

	t.Run("NegativeLimit", func(t *testing.T) {
		res, err := session.CallTool(ctx, &sdkmcp.CallToolParams{
			Name: "list_config_keys",
			Arguments: map[string]any{
				"limit": -1,
			},
		})
		require.NoError(t, err)
		require.True(t, res.IsError)
	})

	t.Run("LimitCappedAtMax", func(t *testing.T) {
		res, err := session.CallTool(ctx, &sdkmcp.CallToolParams{
			Name: "list_config_keys",
			Arguments: map[string]any{
				"limit": maxResultLimit + 10,
			},
		})
		require.NoError(t, err)
		require.False(t, res.IsError)

		raw, err := json.Marshal(res.StructuredContent)
		require.NoError(t, err)

		var out ListConfigKeysOutput
		require.NoError(t, json.Unmarshal(raw, &out))
		require.LessOrEqual(t, len(out.Matches), maxResultLimit)
	})

	t.Run("LimitExactlyMax", func(t *testing.T) {
		res, err := session.CallTool(ctx, &sdkmcp.CallToolParams{
			Name: "list_config_keys",
			Arguments: map[string]any{
				"limit": maxResultLimit,
			},
		})
		require.NoError(t, err)
		require.False(t, res.IsError)

		raw, err := json.Marshal(res.StructuredContent)
		require.NoError(t, err)

		var out ListConfigKeysOutput
		require.NoError(t, json.Unmarshal(raw, &out))
		require.LessOrEqual(t, len(out.Matches), maxResultLimit)
	})

	t.Run("LimitOne", func(t *testing.T) {
		res, err := session.CallTool(ctx, &sdkmcp.CallToolParams{
			Name: "list_config_keys",
			Arguments: map[string]any{
				"limit": 1,
			},
		})
		require.NoError(t, err)
		require.False(t, res.IsError)

		raw, err := json.Marshal(res.StructuredContent)
		require.NoError(t, err)

		var out ListConfigKeysOutput
		require.NoError(t, json.Unmarshal(raw, &out))
		require.Len(t, out.Matches, 1)
	})
}

// TestFindSearchFiltersValidation exercises edge cases for search filter input validation.
func TestFindSearchFiltersValidation(t *testing.T) {
	ctx := context.Background()
	session := connectTestClient(t)

	t.Run("QueryTooLong", func(t *testing.T) {
		longQuery := strings.Repeat("a", maxQueryLength+1)

		res, err := session.CallTool(ctx, &sdkmcp.CallToolParams{
			Name: "find_search_filters",
			Arguments: map[string]any{
				"query": longQuery,
			},
		})
		require.NoError(t, err)
		require.True(t, res.IsError)
	})

	t.Run("WhitespaceQuery", func(t *testing.T) {
		res, err := session.CallTool(ctx, &sdkmcp.CallToolParams{
			Name: "find_search_filters",
			Arguments: map[string]any{
				"query": "   ",
			},
		})
		require.NoError(t, err)
		require.False(t, res.IsError)

		raw, err := json.Marshal(res.StructuredContent)
		require.NoError(t, err)

		var out FindSearchFiltersOutput
		require.NoError(t, json.Unmarshal(raw, &out))
		require.NotEmpty(t, out.Matches)
	})

	t.Run("UnicodeQuery", func(t *testing.T) {
		res, err := session.CallTool(ctx, &sdkmcp.CallToolParams{
			Name: "find_search_filters",
			Arguments: map[string]any{
				"query": "日本語",
			},
		})
		require.NoError(t, err)
		require.False(t, res.IsError)
	})

	t.Run("NegativeLimit", func(t *testing.T) {
		res, err := session.CallTool(ctx, &sdkmcp.CallToolParams{
			Name: "find_search_filters",
			Arguments: map[string]any{
				"limit": -1,
			},
		})
		require.NoError(t, err)
		require.True(t, res.IsError)
	})

	t.Run("LimitCappedAtMax", func(t *testing.T) {
		res, err := session.CallTool(ctx, &sdkmcp.CallToolParams{
			Name: "find_search_filters",
			Arguments: map[string]any{
				"limit": maxResultLimit + 10,
			},
		})
		require.NoError(t, err)
		require.False(t, res.IsError)

		raw, err := json.Marshal(res.StructuredContent)
		require.NoError(t, err)

		var out FindSearchFiltersOutput
		require.NoError(t, json.Unmarshal(raw, &out))
		require.LessOrEqual(t, len(out.Matches), maxResultLimit)
	})

	t.Run("LimitOne", func(t *testing.T) {
		res, err := session.CallTool(ctx, &sdkmcp.CallToolParams{
			Name: "find_search_filters",
			Arguments: map[string]any{
				"limit": 1,
			},
		})
		require.NoError(t, err)
		require.False(t, res.IsError)

		raw, err := json.Marshal(res.StructuredContent)
		require.NoError(t, err)

		var out FindSearchFiltersOutput
		require.NoError(t, json.Unmarshal(raw, &out))
		require.Len(t, out.Matches, 1)
		require.True(t, out.Truncated, "limit=1 should truncate the full result set")
		require.NotEmpty(t, out.Warnings, "truncated results should surface a warning")
		require.Contains(t, out.Warnings[0], "results truncated to 1")
	})
}

// TestFindSearchFiltersTool exercises search filter lookup behavior.
func TestFindSearchFiltersTool(t *testing.T) {
	ctx := context.Background()
	session := connectTestClient(t)

	t.Run("Success", func(t *testing.T) {
		res, err := session.CallTool(ctx, &sdkmcp.CallToolParams{
			Name: "find_search_filters",
			Arguments: map[string]any{
				"query": "Berlin",
				"limit": 5,
			},
		})
		require.NoError(t, err)
		require.False(t, res.IsError)

		raw, err := json.Marshal(res.StructuredContent)
		require.NoError(t, err)

		var out FindSearchFiltersOutput
		require.NoError(t, json.Unmarshal(raw, &out))
		require.NotEmpty(t, out.Matches)
		require.Equal(t, "all", out.TypeApplied)
		require.False(t, out.Truncated, "Berlin query should fit in limit=5")
		require.Empty(t, out.Warnings, "non-truncated results should not emit warnings")
	})

	t.Run("FilterByType", func(t *testing.T) {
		res, err := session.CallTool(ctx, &sdkmcp.CallToolParams{
			Name: "find_search_filters",
			Arguments: map[string]any{
				"type":  "switch",
				"query": "hidden",
			},
		})
		require.NoError(t, err)
		require.False(t, res.IsError)

		raw, err := json.Marshal(res.StructuredContent)
		require.NoError(t, err)

		var out FindSearchFiltersOutput
		require.NoError(t, json.Unmarshal(raw, &out))
		require.NotEmpty(t, out.Matches)
		require.Equal(t, "switch", out.TypeApplied)
	})

	t.Run("NoResults", func(t *testing.T) {
		res, err := session.CallTool(ctx, &sdkmcp.CallToolParams{
			Name: "find_search_filters",
			Arguments: map[string]any{
				"query": "definitely-no-match",
			},
		})
		require.NoError(t, err)
		require.False(t, res.IsError)

		raw, err := json.Marshal(res.StructuredContent)
		require.NoError(t, err)

		var out FindSearchFiltersOutput
		require.NoError(t, json.Unmarshal(raw, &out))
		require.Empty(t, out.Matches)
		require.Zero(t, out.TotalMatches)
	})

	t.Run("InvalidType", func(t *testing.T) {
		res, err := session.CallTool(ctx, &sdkmcp.CallToolParams{
			Name: "find_search_filters",
			Arguments: map[string]any{
				"type": "boolean",
			},
		})
		require.NoError(t, err)
		require.True(t, res.IsError)
		require.Len(t, res.Content, 1)

		text, ok := res.Content[0].(*sdkmcp.TextContent)
		require.True(t, ok)
		require.Contains(t, text.Text, "type must be one of decimal, number, string, switch, timestamp, or all")
	})
}
