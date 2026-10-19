package mcp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/auth"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
	logtest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/auth/oauth"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/authn"
)

// Test-only tools declared with a library resource.
const (
	testSearchLabels   = "test_search_labels"
	testManageSettings = "test_manage_settings"
)

var (
	testSearchLabelsAccess   = ToolAccess{Resource: acl.ResourceLabels, Action: acl.ActionSearch}
	testManageSettingsAccess = ToolAccess{Resource: acl.ResourceSettings, Action: acl.ActionManage}
)

// testToolOutput is the output of the test-only tools.
type testToolOutput struct {
	Called bool `json:"called"`
}

// addTestTools registers the test-only tools on the server.
func addTestTools(server *sdkmcp.Server) {
	handler := func(context.Context, *sdkmcp.CallToolRequest, struct{}) (*sdkmcp.CallToolResult, testToolOutput, error) {
		return nil, testToolOutput{Called: true}, nil
	}

	AddTool(server, testSearchLabelsAccess, &sdkmcp.Tool{Name: testSearchLabels, Description: "test"}, handler)
	AddTool(server, testManageSettingsAccess, &sdkmcp.Tool{Name: testManageSettings, Description: "test"}, handler)
}

// testUserSession returns a session of a registered account with the specified role and MCP setting.
func testUserSession(role acl.Role, setting int, scope string) *entity.Session {
	s := &entity.Session{AuthProvider: authn.ProviderLocal.String(), AuthScope: scope, RefID: "sesstestmcp1"}
	s.SetUser(&entity.User{ID: 1234567, UserUID: "urqdrfb72479n047", UserName: "mcp-test", UserRole: role.String(),
		AuthProvider: authn.ProviderLocal.String(), MCP: setting})

	return s
}

// testClientSession returns a session of an API client without a user.
func testClientSession(scope string) *entity.Session {
	return &entity.Session{AuthProvider: authn.ProviderClient.String(), AuthScope: scope, RefID: "sesstestmcp2"}
}

// connectAuthorized connects a client to the server over Streamable HTTP, with the session in the token info.
func connectAuthorized(t *testing.T, server *sdkmcp.Server, s *entity.Session) *sdkmcp.ClientSession {
	t.Helper()

	info := &auth.TokenInfo{UserID: "test", Extra: map[string]any{SessionKey: s, ClientIPKey: "203.0.113.7"}}
	verifier := func(context.Context, string, *http.Request) (*auth.TokenInfo, error) { return info, nil }
	handler := sdkmcp.NewStreamableHTTPHandler(func(*http.Request) *sdkmcp.Server { return server }, nil)
	ts := httptest.NewServer(auth.RequireBearerToken(verifier, &auth.RequireBearerTokenOptions{AllowMissingExpiration: true})(handler))
	t.Cleanup(ts.Close)

	transport := &sdkmcp.StreamableClientTransport{
		Endpoint:             ts.URL,
		HTTPClient:           &http.Client{Transport: bearerTransport{}},
		DisableStandaloneSSE: true,
	}

	client := sdkmcp.NewClient(&sdkmcp.Implementation{Name: "photoprism-mcp-test", Version: "test"}, nil)
	session, err := client.Connect(context.Background(), transport, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = session.Close() })

	return session
}

// bearerTransport adds a bearer token to each request, which the test verifier accepts.
type bearerTransport struct{}

// RoundTrip sets the Authorization header and sends the request.
func (bearerTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r = r.Clone(r.Context())
	r.Header.Set("Authorization", "Bearer test")
	return http.DefaultTransport.RoundTrip(r)
}

// toolNames returns the names of the tools listed by the session.
func toolNames(t *testing.T, session *sdkmcp.ClientSession) []string {
	t.Helper()

	tools, err := session.ListTools(context.Background(), nil)
	require.NoError(t, err)

	names := make([]string, 0, len(tools.Tools))

	for _, tool := range tools.Tools {
		names = append(names, tool.Name)
	}

	return names
}

func TestToolAccess(t *testing.T) {
	t.Run("DefaultScope", func(t *testing.T) {
		assert.Equal(t, "labels", testSearchLabelsAccess.TokenScope())
		assert.Equal(t, []string{"mcp", "labels"}, testSearchLabelsAccess.TokenScopes())
	})
	t.Run("OwnScope", func(t *testing.T) {
		access := ToolAccess{Resource: acl.ResourceMCP, Action: acl.ActionView, Scope: "metrics"}
		assert.Equal(t, "metrics", access.TokenScope())
		assert.Equal(t, []string{"mcp", "metrics"}, access.TokenScopes())
	})
	t.Run("Static", func(t *testing.T) {
		assert.Equal(t, []string{"mcp"}, staticAccess.TokenScopes())
	})
}

func TestDeclareTool(t *testing.T) {
	t.Run("SameAccess", func(t *testing.T) {
		declareTool("test_declare_same", testSearchLabelsAccess)
		declareTool("test_declare_same", testSearchLabelsAccess)
		access, ok := ToolDeclaration("test_declare_same")
		assert.True(t, ok)
		assert.Equal(t, testSearchLabelsAccess, access)
	})
	t.Run("OtherAccess", func(t *testing.T) {
		declareTool("test_declare_other", testSearchLabelsAccess)
		assert.Panics(t, func() { declareTool("test_declare_other", testManageSettingsAccess) })
	})
	t.Run("Missing", func(t *testing.T) {
		_, ok := ToolDeclaration("test_not_declared")
		assert.False(t, ok)
	})
}

// TestLibraryServerDeclarations lists every tool of the library server without its middleware and checks that
// each has a declaration that only reads, so a write tool cannot be added without changing this test.
func TestLibraryServerDeclarations(t *testing.T) {
	session := connectServer(t, newLibraryServer(&sdkmcp.Implementation{Name: "photoprism-mcp", Version: "test"}, "ce", Options{}))
	tools, err := session.ListTools(context.Background(), nil)
	require.NoError(t, err)

	names := make([]string, 0, len(tools.Tools))

	for _, tool := range tools.Tools {
		names = append(names, tool.Name)
		access, ok := ToolDeclaration(tool.Name)
		require.True(t, ok, tool.Name)
		assert.Contains(t, []acl.Permission{acl.ActionView, acl.ActionSearch, acl.ActionDownload}, access.Action, tool.Name)
		assert.Equal(t, access.Static, slices.Contains(ToolNames, tool.Name), tool.Name)
		assert.Contains(t, oauth.AppScopes, access.TokenScope(), tool.Name)
		require.NotNil(t, tool.Annotations, tool.Name)
		assert.True(t, tool.Annotations.ReadOnlyHint, tool.Name)
	}

	assert.ElementsMatch(t, append(append([]string{}, ToolNames...), LibraryToolNames...), names)

	for _, name := range ToolNames {
		access, _ := ToolDeclaration(name)
		assert.Equal(t, staticAccess, access, name)
	}

	// Resources, templates, and prompts pass the middleware unchecked, so the library server must offer only
	// the static resources of the reference server.
	resources, err := session.ListResources(context.Background(), nil)
	require.NoError(t, err)
	uris := make([]string, 0, len(resources.Resources))
	for _, r := range resources.Resources {
		uris = append(uris, r.URI)
	}
	assert.ElementsMatch(t, ResourceURIs, uris)

	templates, err := session.ListResourceTemplates(context.Background(), nil)
	require.NoError(t, err)
	assert.Empty(t, templates.ResourceTemplates)

	prompts, err := session.ListPrompts(context.Background(), nil)
	if err == nil {
		assert.Empty(t, prompts.Prompts)
	}
}

func TestAddTool(t *testing.T) {
	server := sdkmcp.NewServer(&sdkmcp.Implementation{Name: "test"}, nil)
	AddTool(server, testSearchLabelsAccess, &sdkmcp.Tool{Name: "test_add_tool", Description: "test"},
		func(context.Context, *sdkmcp.CallToolRequest, struct{}) (*sdkmcp.CallToolResult, testToolOutput, error) {
			return nil, testToolOutput{Called: true}, nil
		})

	access, ok := ToolDeclaration("test_add_tool")
	assert.True(t, ok)
	assert.Equal(t, testSearchLabelsAccess, access)
	assert.Equal(t, []string{"test_add_tool"}, toolNames(t, connectServer(t, server)))
}

func TestNewLibraryServer(t *testing.T) {
	server := NewLibraryServer(&sdkmcp.Implementation{Name: "photoprism-mcp", Version: "test"}, "ce", Options{})
	session := connectAuthorized(t, server, testUserSession(acl.RoleAdmin, 0, ""))
	assert.ElementsMatch(t, append(append([]string{}, ToolNames...), LibraryToolNames...), toolNames(t, session))
}

func TestRegisterLibraryTools(t *testing.T) {
	server := sdkmcp.NewServer(&sdkmcp.Implementation{Name: "test"}, nil)
	registerLibraryTools(server, Options{})
	assert.ElementsMatch(t, LibraryToolNames, toolNames(t, connectServer(t, server)))
}

func TestToolNotAvailable(t *testing.T) {
	result := toolNotAvailable("search_media")
	assert.True(t, result.IsError)
	require.Len(t, result.Content, 1)
	assert.Equal(t, "tool 'search_media' is not available", result.Content[0].(*sdkmcp.TextContent).Text)
}

func TestRequestClientIP(t *testing.T) {
	newRequest := func(extra map[string]any) *sdkmcp.CallToolRequest {
		return &sdkmcp.CallToolRequest{Extra: &sdkmcp.RequestExtra{TokenInfo: &auth.TokenInfo{Extra: extra}}}
	}

	assert.Equal(t, "", RequestClientIP(nil))
	assert.Equal(t, "", RequestClientIP(&sdkmcp.CallToolRequest{}))
	assert.Equal(t, "", RequestClientIP(newRequest(nil)))
	assert.Equal(t, "203.0.113.7", RequestClientIP(newRequest(map[string]any{ClientIPKey: "203.0.113.7"})))
	assert.Equal(t, "", RequestClientIP(newRequest(map[string]any{ClientIPKey: "not an address"})))
	assert.Equal(t, "", RequestClientIP(newRequest(map[string]any{ClientIPKey: 7})))
}

func TestAuthorizeTool(t *testing.T) {
	NewServer(&sdkmcp.Implementation{Name: "photoprism-mcp", Version: "test"}, "ce", Options{})
	addTestTools(sdkmcp.NewServer(&sdkmcp.Implementation{Name: "test"}, nil))

	t.Run("Undeclared", func(t *testing.T) {
		decision, scopes := AuthorizeTool(testUserSession(acl.RoleAdmin, 0, ""), "test_not_declared")
		assert.Equal(t, ToolRefused, decision)
		assert.Empty(t, scopes)
	})
	t.Run("NoSession", func(t *testing.T) {
		decision, _ := AuthorizeTool(nil, "list_config_keys")
		assert.Equal(t, ToolRefused, decision)
	})
	t.Run("AdminWithoutScope", func(t *testing.T) {
		s := testUserSession(acl.RoleAdmin, 0, "")
		for _, name := range []string{"list_config_keys", testSearchLabels, testManageSettings} {
			decision, _ := AuthorizeTool(s, name)
			assert.Equal(t, ToolAllowed, decision, name)
		}
	})
	t.Run("AccountSetting", func(t *testing.T) {
		decision, _ := AuthorizeTool(testUserSession(acl.RoleAdmin, -1, ""), "list_config_keys")
		assert.Equal(t, ToolRefused, decision)
		decision, _ = AuthorizeTool(testUserSession(acl.RoleAdmin, 1, ""), "list_config_keys")
		assert.Equal(t, ToolAllowed, decision)
	})
	t.Run("RoleWithoutMCP", func(t *testing.T) {
		decision, _ := AuthorizeTool(testUserSession(acl.RoleGuest, 0, ""), "list_config_keys")
		assert.Equal(t, ToolRefused, decision)
	})
	t.Run("UserWithScope", func(t *testing.T) {
		decision, scopes := AuthorizeTool(testUserSession(acl.RoleAdmin, 0, "mcp"), testSearchLabels)
		assert.Equal(t, ToolMissingScope, decision)
		assert.Equal(t, []string{"mcp", "labels"}, scopes)
		decision, _ = AuthorizeTool(testUserSession(acl.RoleAdmin, 0, "mcp labels"), testSearchLabels)
		assert.Equal(t, ToolAllowed, decision)
	})
	t.Run("ClientScope", func(t *testing.T) {
		decision, _ := AuthorizeTool(testClientSession("mcp"), "list_config_keys")
		assert.Equal(t, ToolAllowed, decision)
		decision, scopes := AuthorizeTool(testClientSession("mcp"), testSearchLabels)
		assert.Equal(t, ToolMissingScope, decision)
		assert.Equal(t, []string{"mcp", "labels"}, scopes)
		decision, _ = AuthorizeTool(testClientSession("mcp labels"), testSearchLabels)
		assert.Equal(t, ToolAllowed, decision)
		decision, scopes = AuthorizeTool(testClientSession("labels"), testSearchLabels)
		assert.Equal(t, ToolMissingScope, decision)
		assert.Equal(t, []string{"mcp", "labels"}, scopes)
		decision, _ = AuthorizeTool(testClientSession(""), "list_config_keys")
		assert.Equal(t, ToolMissingScope, decision)
	})
	t.Run("ScopeVerbs", func(t *testing.T) {
		decision, _ := AuthorizeTool(testClientSession("read"), testSearchLabels)
		assert.Equal(t, ToolAllowed, decision)
		decision, _ = AuthorizeTool(testClientSession("*"), testSearchLabels)
		assert.Equal(t, ToolAllowed, decision)
		decision, scopes := AuthorizeTool(testClientSession("mcp read"), testSearchLabels)
		assert.Equal(t, ToolMissingScope, decision)
		assert.Equal(t, []string{"mcp", "labels"}, scopes)
	})
	t.Run("RoleWithoutMCPView", func(t *testing.T) {
		acl.RulesMutex.Lock()
		grant := acl.Rules[acl.ResourceMCP][acl.RoleClient]
		delete(acl.Rules[acl.ResourceMCP], acl.RoleClient)
		acl.RulesMutex.Unlock()
		t.Cleanup(func() {
			acl.RulesMutex.Lock()
			acl.Rules[acl.ResourceMCP][acl.RoleClient] = grant
			acl.RulesMutex.Unlock()
		})

		decision, scopes := AuthorizeTool(testClientSession("*"), testSearchLabels)
		assert.Equal(t, ToolRefused, decision)
		assert.Empty(t, scopes)
	})
	t.Run("ClientRoleWithoutAction", func(t *testing.T) {
		decision, scopes := AuthorizeTool(testClientSession("*"), testManageSettings)
		assert.Equal(t, ToolRefused, decision)
		assert.Empty(t, scopes)
	})
}

func TestRequestSession(t *testing.T) {
	s := testClientSession("mcp")

	t.Run("Nil", func(t *testing.T) {
		assert.Nil(t, RequestSession(nil))
		assert.Nil(t, RequestSession(&sdkmcp.CallToolRequest{}))
		assert.Nil(t, RequestSession(&sdkmcp.CallToolRequest{Extra: &sdkmcp.RequestExtra{}}))
		assert.Nil(t, RequestSession(&sdkmcp.CallToolRequest{Extra: &sdkmcp.RequestExtra{TokenInfo: &auth.TokenInfo{}}}))
	})
	t.Run("Session", func(t *testing.T) {
		req := &sdkmcp.CallToolRequest{Extra: &sdkmcp.RequestExtra{TokenInfo: &auth.TokenInfo{Extra: map[string]any{SessionKey: s}}}}
		assert.Same(t, s, RequestSession(req))
	})
	t.Run("OtherType", func(t *testing.T) {
		req := &sdkmcp.CallToolRequest{Extra: &sdkmcp.RequestExtra{TokenInfo: &auth.TokenInfo{Extra: map[string]any{SessionKey: "sess"}}}}
		assert.Nil(t, RequestSession(req))
	})
}

func TestAuthorizeTools(t *testing.T) {
	newServer := func() *sdkmcp.Server {
		server := NewLibraryServer(&sdkmcp.Implementation{Name: "photoprism-mcp", Version: "test"}, "ce", Options{})
		addTestTools(server)
		return server
	}

	call := func(t *testing.T, session *sdkmcp.ClientSession, name string) *sdkmcp.CallToolResult {
		t.Helper()
		result, err := session.CallTool(context.Background(), &sdkmcp.CallToolParams{Name: name, Arguments: map[string]any{}})
		require.NoError(t, err)
		return result
	}

	t.Run("Admin", func(t *testing.T) {
		session := connectAuthorized(t, newServer(), testUserSession(acl.RoleAdmin, 0, ""))
		assert.ElementsMatch(t, append(append(append([]string{}, ToolNames...), LibraryToolNames...), testSearchLabels, testManageSettings), toolNames(t, session))
		assert.False(t, call(t, session, testManageSettings).IsError)
		assert.False(t, call(t, session, testSearchLabels).IsError)
		assert.False(t, call(t, session, "list_thumbnail_sizes").IsError)
	})
	t.Run("ClientWithMCPScope", func(t *testing.T) {
		// Tools are listed by role, so a client learns which scope to request; the call is refused.
		session := connectAuthorized(t, newServer(), testClientSession("mcp"))
		names := toolNames(t, session)
		assert.Contains(t, names, testSearchLabels)
		assert.NotContains(t, names, testManageSettings)
		assert.Subset(t, names, ToolNames)

		result := call(t, session, testSearchLabels)
		assert.True(t, result.IsError)
		require.Len(t, result.Content, 1)
		text := result.Content[0].(*sdkmcp.TextContent).Text
		assert.Contains(t, text, testSearchLabels)
		assert.NotContains(t, text, "scope")
		assert.False(t, call(t, session, "list_config_keys").IsError)
	})
	t.Run("RoleWithoutAction", func(t *testing.T) {
		orig := event.AuditLog
		logger, hook := logtest.NewNullLogger()
		event.AuditLog = logger
		t.Cleanup(func() { event.AuditLog = orig })

		sub := event.Subscribe(string(acl.ChannelAudit) + ".log.*")
		t.Cleanup(func() { event.Unsubscribe(sub) })

		session := connectAuthorized(t, newServer(), testClientSession("*"))
		assert.NotContains(t, toolNames(t, session), testManageSettings)
		assert.True(t, call(t, session, testManageSettings).IsError)
		require.NotNil(t, hook.LastEntry())
		assert.Contains(t, hook.LastEntry().Message, "203.0.113.7 › session sesstestmcp2 › mcp › call test_manage_settings › denied")

		select {
		case msg := <-sub.Receiver:
			assert.Equal(t, "203.0.113.7", msg.Fields["ip"])
		case <-time.After(time.Second):
			t.Fatal("no audit event was published")
		}
		assert.False(t, call(t, session, testSearchLabels).IsError)
	})
	t.Run("Undeclared", func(t *testing.T) {
		orig := event.AuditLog
		logger, hook := logtest.NewNullLogger()
		event.AuditLog = logger
		t.Cleanup(func() { event.AuditLog = orig })

		server := newServer()
		sdkmcp.AddTool(server, &sdkmcp.Tool{Name: "test_undeclared", Description: "test"},
			func(context.Context, *sdkmcp.CallToolRequest, struct{}) (*sdkmcp.CallToolResult, testToolOutput, error) {
				return nil, testToolOutput{Called: true}, nil
			})
		session := connectAuthorized(t, server, testUserSession(acl.RoleAdmin, 0, ""))
		assert.NotContains(t, toolNames(t, session), "test_undeclared")
		assert.True(t, call(t, session, "test_undeclared").IsError)
		assert.True(t, call(t, session, "made_up_name").IsError)
		assert.Empty(t, hook.AllEntries(), "calls of undeclared names are not audited")
	})
	t.Run("AccountSetting", func(t *testing.T) {
		session := connectAuthorized(t, newServer(), testUserSession(acl.RoleAdmin, -1, ""))
		assert.Empty(t, toolNames(t, session))
		assert.True(t, call(t, session, "list_config_keys").IsError)
	})
	t.Run("NoSession", func(t *testing.T) {
		session := connectServer(t, newServer())
		assert.Empty(t, toolNames(t, session))
		assert.True(t, call(t, session, "list_config_keys").IsError)
	})
}
