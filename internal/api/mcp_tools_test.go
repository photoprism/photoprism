package api

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/gin-gonic/gin"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/mcp"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/http/header"
)

// Test-only tools declared with library resources.
const (
	testSearchPhotosTool   = "test_search_photos"
	testSearchLabelsTool   = "test_search_labels"
	testManageSettingsTool = "test_manage_settings"
)

// testToolOutput is the output of the test-only tools.
type testToolOutput struct {
	Called bool `json:"called"`
}

// testToolHandler is the handler of the test-only tools.
func testToolHandler(context.Context, *sdkmcp.CallToolRequest, struct{}) (*sdkmcp.CallToolResult, testToolOutput, error) {
	return nil, testToolOutput{Called: true}, nil
}

// withTestLibraryTool adds testSearchPhotosTool to the library server that ServeMCP creates.
func withTestLibraryTool(t *testing.T) {
	t.Helper()

	prev := newMcpLibraryServer
	t.Cleanup(func() { newMcpLibraryServer = prev })

	newMcpLibraryServer = func(implementation *sdkmcp.Implementation, edition string, opt mcp.Options) *sdkmcp.Server {
		server := prev(implementation, edition, opt)
		mcp.AddTool(server, mcp.ToolAccess{Resource: acl.ResourcePhotos, Action: acl.ActionSearch}, &sdkmcp.Tool{Name: testSearchPhotosTool, Description: "test"}, testToolHandler)
		mcp.AddTool(server, mcp.ToolAccess{Resource: acl.ResourceLabels, Action: acl.ActionSearch}, &sdkmcp.Tool{Name: testSearchLabelsTool, Description: "test"}, testToolHandler)
		mcp.AddTool(server, mcp.ToolAccess{Resource: acl.ResourceSettings, Action: acl.ActionManage}, &sdkmcp.Tool{Name: testManageSettingsTool, Description: "test"}, testToolHandler)
		return server
	}
}

// mcpCallBody returns the JSON-RPC request that calls the tool.
func mcpCallBody(id int, name string) string {
	return fmt.Sprintf(`{"jsonrpc":"2.0","id":%d,"method":"tools/call","params":{"name":"%s","arguments":{}}}`, id, name)
}

func TestServeMCP_ToolAuthorization(t *testing.T) {
	callBody := `{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"` + testSearchPhotosTool + `","arguments":{}}}`
	listBody := `{"jsonrpc":"2.0","id":3,"method":"tools/list"}`

	alice := entity.FindUserByName("alice")
	require.NotNil(t, alice)

	newToken := func(t *testing.T, scope string) string {
		sess, err := entity.AddClientSession("mcp-test-tools-"+strings.ReplaceAll(scope, " ", "-"), 3600, scope, authn.GrantCLI, alice)
		require.NoError(t, err)
		t.Cleanup(func() { _ = sess.Delete() })
		return sess.AuthToken()
	}

	t.Run("StepUp", func(t *testing.T) {
		withTestLibraryTool(t)
		app, router, _ := NewApiTest()
		conf := prepareMCPTest(t)
		ServeMCP(router)

		token := newToken(t, "mcp")
		sessionID := mcpInitSession(t, app, token)

		// The tool is listed, so the client learns that it exists.
		w := mcpPost(app, listBody, token, sessionID)
		require.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"name":"`+testSearchPhotosTool+`"`)

		audit := captureAuditLog(t)
		w = mcpPost(app, callBody, token, sessionID)
		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Equal(t, McpChallenge(conf, McpErrInsufficientScope, "mcp", "photos"), w.Header().Get(header.WWWAuthenticate))
		require.NotNil(t, audit.LastEntry())
		assert.Contains(t, audit.LastEntry().Message, "mcp › call needs scope 'mcp photos' › denied")

		// Reference tools need no other scope.
		w = mcpPost(app, `{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"list_thumbnail_sizes","arguments":{}}}`, token, sessionID)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.NotContains(t, w.Body.String(), `"isError":true`)

		// A token with the scope of the tool may call it.
		photosToken := newToken(t, "mcp photos")
		photosSession := mcpInitSession(t, app, photosToken)
		w = mcpPost(app, callBody, photosToken, photosSession)
		require.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"called":true`)
		assert.NotContains(t, w.Body.String(), `"isError":true`)
	})
	t.Run("Batch", func(t *testing.T) {
		withTestLibraryTool(t)
		app, router, _ := NewApiTest()
		conf := prepareMCPTest(t)
		ServeMCP(router)

		token := newToken(t, "mcp")
		sessionID := mcpInitSession(t, app, token)

		// The SDK accepts batches from clients that send no protocol version header, as mcpPost does.
		batch := `[` + mcpCallBody(5, "list_thumbnail_sizes") + `,` + callBody + `]`
		w := mcpPost(app, batch, token, sessionID)
		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Equal(t, McpChallenge(conf, McpErrInsufficientScope, "mcp", "photos"), w.Header().Get(header.WWWAuthenticate))

		// The challenge names the union of the missing scopes once each.
		batch = `[` + callBody + `,` + mcpCallBody(6, testSearchLabelsTool) + `,` + mcpCallBody(7, testSearchPhotosTool) + `]`
		w = mcpPost(app, batch, token, sessionID)
		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Equal(t, McpChallenge(conf, McpErrInsufficientScope, "mcp", "photos", "labels"), w.Header().Get(header.WWWAuthenticate))

		// The scopes the token already has are named too, so a new authorization keeps them.
		labelsToken := newToken(t, "mcp labels")
		labelsSession := mcpInitSession(t, app, labelsToken)
		w = mcpPost(app, callBody, labelsToken, labelsSession)
		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Equal(t, McpChallenge(conf, McpErrInsufficientScope, "labels", "mcp", "photos"), w.Header().Get(header.WWWAuthenticate))

		// Batches are limited in size.
		messages := make([]string, MaxMCPBatchMessages+1)
		for i := range messages {
			messages[i] = mcpCallBody(10+i, "list_thumbnail_sizes")
		}
		w = mcpPost(app, `[`+strings.Join(messages, ",")+`]`, token, sessionID)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		w = mcpPost(app, `[`+strings.Join(messages, ",")+`] x`, token, sessionID)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		w = mcpPost(app, `[`+strings.Join(messages[:MaxMCPBatchMessages], ",")+`]`, token, sessionID)
		assert.Equal(t, http.StatusOK, w.Code)
	})
	t.Run("RoleCannotUseTool", func(t *testing.T) {
		// A role that cannot use a tool gets an error result, not a request for more scopes.
		withTestLibraryTool(t)
		app, router, _ := NewApiTest()
		prepareMCPTest(t)
		ServeMCP(router)

		client, err := entity.AddClientSession("mcp-test-tools-client", 3600, "*", authn.GrantClientCredentials, nil)
		require.NoError(t, err)
		t.Cleanup(func() { _ = client.Delete() })

		token := client.AuthToken()
		sessionID := mcpInitSession(t, app, token)

		w := mcpPost(app, listBody, token, sessionID)
		require.Equal(t, http.StatusOK, w.Code)
		assert.NotContains(t, w.Body.String(), testManageSettingsTool)
		assert.Contains(t, w.Body.String(), testSearchLabelsTool)

		w = mcpPost(app, mcpCallBody(8, testManageSettingsTool), token, sessionID)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Empty(t, w.Header().Get(header.WWWAuthenticate))
		assert.Contains(t, w.Body.String(), `"isError":true`)
		assert.Contains(t, w.Body.String(), "is not available")
		assert.NotContains(t, w.Body.String(), `"called":true`)
	})
	t.Run("UnreadableBody", func(t *testing.T) {
		withTestLibraryTool(t)
		app, router, _ := NewApiTest()
		prepareMCPTest(t)
		ServeMCP(router)

		token := newToken(t, "mcp")
		sessionID := mcpInitSession(t, app, token)

		// Bodies that are not a single JSON value are refused before the SDK runs them.
		for _, body := range []string{`{"jsonrpc":"2.0","id":9,"method":`, callBody + ` x`, `[` + callBody + `] x`} {
			w := mcpPost(app, body, token, sessionID)
			assert.Equal(t, http.StatusBadRequest, w.Code, body)
			assert.Empty(t, w.Header().Get(header.WWWAuthenticate), body)
			assert.NotContains(t, w.Body.String(), `"called":true`, body)
		}

		// Other methods need no scope.
		w := mcpPost(app, listBody, token, sessionID)
		assert.Equal(t, http.StatusOK, w.Code)
	})
	t.Run("AdminSession", func(t *testing.T) {
		withTestLibraryTool(t)
		app, router, _ := NewApiTest()
		prepareMCPTest(t)
		ServeMCP(router)

		token := AuthenticateAdmin(app, router)
		sessionID := mcpInitSession(t, app, token)
		w := mcpPost(app, callBody, token, sessionID)
		require.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"called":true`)

		assertMCPReferenceTools(t, app, token, sessionID)
	})
	t.Run("PublicMode", func(t *testing.T) {
		withTestLibraryTool(t)
		app, router, _ := NewApiTest()
		conf := prepareMCPTest(t)
		conf.Options().Public = true
		ServeMCP(router)

		sessionID := mcpInitSession(t, app, "")
		w := mcpPost(app, listBody, "", sessionID)
		require.Equal(t, http.StatusOK, w.Code)
		assert.NotContains(t, w.Body.String(), testSearchPhotosTool)
		assert.Equal(t, len(mcp.ToolNames), strings.Count(w.Body.String(), `"inputSchema"`))

		w = mcpPost(app, callBody, "", sessionID)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `unknown tool`)
		assert.NotContains(t, w.Body.String(), `"called":true`)
	})
}

func TestMcpReadCalls(t *testing.T) {
	newContext := func(body io.Reader) *gin.Context {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/mcp", body)
		return c
	}

	t.Run("Single", func(t *testing.T) {
		c := newContext(strings.NewReader(mcpCallBody(1, "search_media")))
		calls, ok := mcpReadCalls(c, &entity.Session{})
		require.True(t, ok)
		require.Len(t, calls, 1)
		assert.Equal(t, "tools/call", calls[0].Method)
		assert.Equal(t, "search_media", calls[0].Params.Name)

		// The body is restored for the SDK.
		body, err := io.ReadAll(c.Request.Body)
		require.NoError(t, err)
		assert.Equal(t, mcpCallBody(1, "search_media"), string(body))
	})
	t.Run("Batch", func(t *testing.T) {
		c := newContext(strings.NewReader(" [" + mcpCallBody(1, "a") + `,{"jsonrpc":"2.0","method":"notifications/initialized"},` + mcpCallBody(2, "b") + "]"))
		calls, ok := mcpReadCalls(c, &entity.Session{})
		require.True(t, ok)
		require.Len(t, calls, 3)
		assert.Equal(t, "b", calls[2].Params.Name)
	})
	t.Run("Empty", func(t *testing.T) {
		calls, ok := mcpReadCalls(newContext(strings.NewReader(" ")), &entity.Session{})
		assert.True(t, ok)
		assert.Empty(t, calls)
	})
	t.Run("Malformed", func(t *testing.T) {
		for _, body := range []string{"[{", "{", "[1,", mcpCallBody(1, "a") + " x", "[" + mcpCallBody(1, "a") + "] x", "[" + mcpCallBody(1, "a") + "]]", mcpCallBody(1, "a") + "\n0"} {
			c := newContext(strings.NewReader(body))
			_, ok := mcpReadCalls(c, &entity.Session{})
			assert.False(t, ok, body)
			assert.Equal(t, http.StatusBadRequest, c.Writer.Status(), body)
		}
	})
	t.Run("TrailingWhitespace", func(t *testing.T) {
		calls, ok := mcpReadCalls(newContext(strings.NewReader(mcpCallBody(1, "a")+"\n\t ")), &entity.Session{})
		assert.True(t, ok)
		assert.Len(t, calls, 1)
	})
	t.Run("BatchLimit", func(t *testing.T) {
		messages := make([]string, MaxMCPBatchMessages+1)
		for i := range messages {
			messages[i] = mcpCallBody(i, "a")
		}
		for _, suffix := range []string{"", " x", "\n0"} {
			c := newContext(strings.NewReader("[" + strings.Join(messages, ",") + "]" + suffix))
			_, ok := mcpReadCalls(c, &entity.Session{})
			assert.False(t, ok, suffix)
			assert.Equal(t, http.StatusBadRequest, c.Writer.Status(), suffix)
		}
	})
	t.Run("HelperCaps", func(t *testing.T) {
		// The helper caps the body itself.
		c := newContext(strings.NewReader("[" + strings.Repeat(" ", int(MaxMCPRequestBytes)) + "]"))
		_, ok := mcpReadCalls(c, &entity.Session{})
		assert.False(t, ok)
		assert.Equal(t, http.StatusRequestEntityTooLarge, c.Writer.Status())
	})
	t.Run("ReadError", func(t *testing.T) {
		c := newContext(iotest.ErrReader(errors.New("read failed")))
		_, ok := mcpReadCalls(c, &entity.Session{})
		assert.False(t, ok)
		assert.Equal(t, http.StatusBadRequest, c.Writer.Status())
	})
	t.Run("TooLarge", func(t *testing.T) {
		c := newContext(strings.NewReader(strings.Repeat("a", 64)))
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 16)
		_, ok := mcpReadCalls(c, &entity.Session{})
		assert.False(t, ok)
		assert.Equal(t, http.StatusRequestEntityTooLarge, c.Writer.Status())
	})
}

func TestMcpMissingScopes(t *testing.T) {
	withTestLibraryTool(t)
	newMcpLibraryServer(&sdkmcp.Implementation{Name: "test"}, "ce", mcp.Options{})

	alice := entity.FindUserByName("alice")
	require.NotNil(t, alice)

	s := entity.NewSession(3600, 0)
	s.SetUser(alice)
	s.AuthProvider = authn.ProviderApplication.String()
	s.AuthScope = "mcp"

	calls := []mcpCall{{Method: "tools/call"}, {Method: "tools/list"}, {Method: "tools/call"}}
	calls[0].Params.Name = testSearchPhotosTool
	calls[1].Params.Name = testSearchLabelsTool
	calls[2].Params.Name = testSearchPhotosTool

	t.Run("Union", func(t *testing.T) {
		assert.Equal(t, []string{"mcp", "photos"}, mcpMissingScopes(s, calls, false))
	})
	t.Run("PublicMode", func(t *testing.T) {
		assert.Empty(t, mcpMissingScopes(s, calls, true))
	})
	t.Run("ReferenceTool", func(t *testing.T) {
		call := mcpCall{Method: "tools/call"}
		call.Params.Name = "list_config_keys"
		assert.Empty(t, mcpMissingScopes(s, []mcpCall{call}, false))
	})
}

func TestMcpChallengeScopes(t *testing.T) {
	t.Run("Missing", func(t *testing.T) {
		assert.Equal(t, []string{"mcp", "photos"}, mcpChallengeScopes(&entity.Session{AuthScope: "mcp"}, []string{"mcp", "photos"}))
	})
	t.Run("KeepsAppScopes", func(t *testing.T) {
		assert.Equal(t, []string{"labels", "mcp", "photos"}, mcpChallengeScopes(&entity.Session{AuthScope: "mcp labels"}, []string{"mcp", "photos"}))
	})
	t.Run("IgnoresOtherScopes", func(t *testing.T) {
		assert.Equal(t, []string{"mcp", "photos"}, mcpChallengeScopes(&entity.Session{AuthScope: "read webdav"}, []string{"mcp", "photos"}))
	})
}
