package mcp

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

// connectTestClient connects an in-memory client to the MCP server.
func connectTestClient(t *testing.T) *sdkmcp.ClientSession {
	t.Helper()

	ctx := context.Background()
	server := NewServer(&sdkmcp.Implementation{Name: "photoprism-mcp", Version: "test"}, "ce")
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

// TestNewServerResources lists and reads the registered MCP resources.
func TestNewServerResources(t *testing.T) {
	ctx := context.Background()
	session := connectTestClient(t)

	resources, err := session.ListResources(ctx, nil)
	require.NoError(t, err)
	require.Len(t, resources.Resources, 2)
	require.Equal(t, configOptionsURI, resources.Resources[0].URI)
	require.Equal(t, searchFiltersURI, resources.Resources[1].URI)

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
}

// TestNewServerToolList exposes the expected list_config_keys tool.
func TestNewServerToolList(t *testing.T) {
	ctx := context.Background()
	session := connectTestClient(t)

	tools, err := session.ListTools(ctx, nil)
	require.NoError(t, err)
	require.Len(t, tools.Tools, 2)

	names := []string{tools.Tools[0].Name, tools.Tools[1].Name}
	require.Contains(t, names, "list_config_keys")
	require.Contains(t, names, "find_search_filters")

	for _, tool := range tools.Tools {
		require.NotNil(t, tool.InputSchema)
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
