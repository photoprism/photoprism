package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/mcp"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/http/header"
)

// prepareMCPTest sets auth mode to password and disables public mode so that
// MCP auth checks behave as in production.
func prepareMCPTest(t *testing.T) *config.Config {
	t.Helper()

	conf := get.Config()
	originalOptions := *conf.Options()

	t.Cleanup(func() {
		*conf.Options() = originalOptions
	})

	conf.Options().AuthMode = config.AuthModePasswd
	conf.Options().Public = false

	return conf
}

// mcpPost sends a JSON-RPC POST to /api/v1/mcp with MCP-required headers.
func mcpPost(app http.Handler, body, authToken, sessionID string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/mcp", strings.NewReader(body))

	if authToken != "" {
		header.SetAuthorization(req, authToken)
	}

	req.Header.Set(header.ContentType, header.ContentTypeJson)
	req.Header.Set(header.Accept, header.ContentTypeJson+", "+header.ContentTypeEventStream)

	if sessionID != "" {
		req.Header.Set("Mcp-Session-Id", sessionID)
	}

	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	return w
}

// assertMCPReferenceTools lists the tools of an initialized MCP session and calls each reference tool
// that returns a list of items, asserting a successful result with rows.
func assertMCPReferenceTools(t *testing.T, app http.Handler, authToken, sessionID string) {
	t.Helper()

	w := mcpPost(app, `{"jsonrpc":"2.0","id":10,"method":"tools/list"}`, authToken, sessionID)
	require.Equal(t, http.StatusOK, w.Code)

	for _, name := range mcp.ToolNames {
		assert.Contains(t, w.Body.String(), `"name":"`+name+`"`)
	}

	calls := map[string]string{
		"list_file_formats":    `{"type":"raw","limit":5}`,
		"list_metadata_fields": `{"query":"exif","limit":5}`,
		"list_thumbnail_sizes": `{"type":"video"}`,
	}

	for name, args := range calls {
		w = mcpPost(app, `{"jsonrpc":"2.0","id":11,"method":"tools/call","params":{"name":"`+name+`","arguments":`+args+`}}`, authToken, sessionID)
		require.Equal(t, http.StatusOK, w.Code, name)
		assert.Contains(t, w.Body.String(), `"items":[{`, name)
		assert.NotContains(t, w.Body.String(), `"isError":true`, name)
	}

	// The test config's CLI app has no commands, so the tool returns no items and a warning.
	w = mcpPost(app, `{"jsonrpc":"2.0","id":12,"method":"tools/call","params":{"name":"list_cli_commands","arguments":{"query":"users"}}}`, authToken, sessionID)
	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"items":[]`)
	assert.Contains(t, w.Body.String(), "the command list of this edition is not available")
	assert.NotContains(t, w.Body.String(), `"isError":true`)
}

// mcpInitSession initializes an MCP session and returns its ID.
func mcpInitSession(t *testing.T, app http.Handler, authToken string) string {
	t.Helper()

	w := mcpPost(app, `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0"}}}`, authToken, "")
	require.Equal(t, http.StatusOK, w.Code)

	sessionID := w.Header().Get("Mcp-Session-Id")
	require.NotEmpty(t, sessionID)

	w = mcpPost(app, `{"jsonrpc":"2.0","method":"notifications/initialized"}`, authToken, sessionID)
	require.Less(t, w.Code, 300)

	return sessionID
}

// TestServeMCP exercises the HTTP handler installed by ServeMCP: the
// public-mode anonymous path, unauthenticated and non-admin denial, and
// the full admin round-trip through initialize, notifications/initialized,
// and tools/call.
func TestServeMCP(t *testing.T) {
	t.Run("AllowedPublicMode", func(t *testing.T) {
		// In public mode, Session() returns the default public session and
		// the currently registered MCP tools only surface static reference
		// data. Anonymous callers must therefore be able to initialize an
		// MCP session and call every tool without a token - this is what
		// lets the MCP server run on demo.photoprism.app. Guard the policy
		// here so it regresses loudly if a future change tightens it.
		app, router, _ := NewApiTest()
		conf := prepareMCPTest(t)
		conf.Options().Public = true
		ServeMCP(router)

		initBody := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0"}}}`
		w := mcpPost(app, initBody, "", "")

		assert.Equal(t, http.StatusOK, w.Code)
		sessionID := w.Header().Get("Mcp-Session-Id")
		assert.NotEmpty(t, sessionID)
		assert.Contains(t, w.Body.String(), "photoprism-mcp")

		w2 := mcpPost(app, `{"jsonrpc":"2.0","method":"notifications/initialized"}`, "", sessionID)
		assert.Less(t, w2.Code, 300, "notification should succeed, got %d: %s", w2.Code, w2.Body.String())

		w3 := mcpPost(app, `{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"list_config_keys","arguments":{"query":"database","limit":5}}}`, "", sessionID)
		assert.Equal(t, http.StatusOK, w3.Code)
		assert.Contains(t, w3.Body.String(), "matches")

		w4 := mcpPost(app, `{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"find_search_filters","arguments":{"query":"color","limit":5}}}`, "", sessionID)
		assert.Equal(t, http.StatusOK, w4.Code)
		assert.Contains(t, w4.Body.String(), "matches")

		assertMCPReferenceTools(t, app, "", sessionID)
	})
	t.Run("EditionCommandList", func(t *testing.T) {
		// ServeMCP passes the command list of the running edition from the CLI app to the server.
		app, router, _ := NewApiTest()
		conf := prepareMCPTest(t)
		conf.Options().Public = true

		cliApp := conf.CliApp()
		require.NotNil(t, cliApp)

		commands := cliApp.Commands
		t.Cleanup(func() { cliApp.Commands = commands })

		cliApp.Commands = []*cli.Command{
			{Name: "ldap", Usage: "Directory commands", Subcommands: []*cli.Command{
				{Name: "sync", Usage: "Synchronizes accounts", Flags: []cli.Flag{
					&cli.StringFlag{Name: "role", Usage: "account role", Value: "user"},
					&cli.BoolFlag{Name: "debug", Usage: "has the name of a global flag"},
					cli.HelpFlag,
				}},
			}},
			{Name: "secret", Usage: "Hidden command", Hidden: true},
		}

		ServeMCP(router)

		sessionID := mcpInitSession(t, app, "")
		w := mcpPost(app, `{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"list_cli_commands","arguments":{}}}`, "", sessionID)
		require.Equal(t, http.StatusOK, w.Code)

		body := w.Body.String()
		assert.Contains(t, body, `"command":"photoprism ldap"`)
		assert.Contains(t, body, `"command":"photoprism ldap sync"`)
		assert.Contains(t, body, `"name":"--role"`)
		assert.Contains(t, body, `\"user\"`)
		assert.Contains(t, body, `"total":2`)
		assert.NotContains(t, body, "photoprism secret")
		assert.Contains(t, body, `"name":"--debug"`)
		assert.NotContains(t, body, `"name":"--help"`)
		assert.NotContains(t, body, "not available")
	})
	t.Run("DisabledViaConfig", func(t *testing.T) {
		// When DisableMCP is set, ServeMCP must skip route registration so
		// the endpoint responds with the standard 404 — mirroring how the
		// other Disable* flags short-circuit route setup.
		app, router, _ := NewApiTest()
		conf := prepareMCPTest(t)
		conf.Options().DisableMCP = true
		ServeMCP(router)

		r := mcpPost(app, `{"jsonrpc":"2.0","id":1,"method":"initialize"}`, "", "")
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("UnauthorizedAnonymous", func(t *testing.T) {
		app, router, _ := NewApiTest()
		prepareMCPTest(t)
		ServeMCP(router)

		r := mcpPost(app, `{"jsonrpc":"2.0","id":1,"method":"initialize"}`, "", "")
		assert.Equal(t, http.StatusUnauthorized, r.Code)
	})
	t.Run("ForbiddenNonAdmin", func(t *testing.T) {
		app, router, _ := NewApiTest()
		prepareMCPTest(t)
		ServeMCP(router)

		authToken := AuthenticateUser(app, router, "gandalf", "Gandalf123!")
		r := mcpPost(app, `{"jsonrpc":"2.0","id":1,"method":"initialize"}`, authToken, "")
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
	t.Run("AdminAppPasswordWithoutMCPScope", func(t *testing.T) {
		app, router, _ := NewApiTest()
		prepareMCPTest(t)
		ServeMCP(router)

		user := entity.FindUserByName("alice")
		if user == nil {
			t.Fatal("admin fixture user 'alice' must not be nil")
		}

		sess, err := entity.AddClientSession("mcp-test-wrong-scope", 3600, "webdav", authn.GrantCLI, user)
		assert.NoError(t, err)
		if err != nil {
			return
		}

		r := mcpPost(app, `{"jsonrpc":"2.0","id":1,"method":"initialize"}`, sess.AuthToken(), "")
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
	t.Run("AdminAppPasswordWithMCPScope", func(t *testing.T) {
		app, router, _ := NewApiTest()
		prepareMCPTest(t)
		ServeMCP(router)

		user := entity.FindUserByName("alice")
		if user == nil {
			t.Fatal("admin fixture user 'alice' must not be nil")
		}

		sess, err := entity.AddClientSession("mcp-test-right-scope", 3600, "mcp", authn.GrantCLI, user)
		assert.NoError(t, err)
		if err != nil {
			return
		}

		initBody := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"qa","version":"1.0"}}}`
		r := mcpPost(app, initBody, sess.AuthToken(), "")
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("AccountSetting", func(t *testing.T) {
		app, router, _ := NewApiTest()
		conf := prepareMCPTest(t)
		ServeMCP(router)

		alice := entity.FindUserByName("alice")
		require.NotNil(t, alice)
		setMCP := func(value int) {
			alice.MCP = value
			require.NoError(t, alice.Save())
		}
		t.Cleanup(func() { setMCP(0) })

		appPassword, err := entity.AddClientSession("mcp-test-account-setting", 3600, "mcp", authn.GrantCLI, alice)
		require.NoError(t, err)
		t.Cleanup(func() { _ = appPassword.Delete() })

		client := entity.ClientFixtures.Pointer("alice").NewSession(nil, authn.GrantClientCredentials)
		require.NoError(t, client.Create())
		t.Cleanup(func() { _ = client.Delete() })
		require.True(t, client.IsClient())
		require.True(t, client.HasUser())

		initBody := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"qa","version":"1.0"}}}`

		for _, value := range []int{0, 1} {
			setMCP(value)
			assert.Equal(t, http.StatusOK, mcpPost(app, initBody, appPassword.AuthToken(), "").Code)
			assert.Equal(t, http.StatusOK, mcpPost(app, initBody, client.AuthToken(), "").Code)
		}

		// An open MCP session stops working once the setting turns MCP off.
		setMCP(0)
		sessionID := mcpInitSession(t, app, appPassword.AuthToken())
		setMCP(-1)

		audit := captureAuditLog(t)
		w := mcpPost(app, `{"jsonrpc":"2.0","id":2,"method":"tools/list"}`, appPassword.AuthToken(), sessionID)
		assert.Equal(t, http.StatusForbidden, w.Code)
		require.NotNil(t, audit.LastEntry())
		assert.Contains(t, audit.LastEntry().Message, "mcp › disabled for user alice › denied")

		assert.Equal(t, http.StatusForbidden, mcpPost(app, initBody, appPassword.AuthToken(), "").Code)
		assert.Equal(t, http.StatusForbidden, mcpPost(app, initBody, client.AuthToken(), "").Code)

		// The event stream and session tear-down are refused as well.
		for _, method := range []string{http.MethodGet, http.MethodDelete} {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			req, _ := http.NewRequestWithContext(ctx, method, "/api/v1/mcp", nil)
			header.SetAuthorization(req, appPassword.AuthToken())
			req.Header.Set(header.Accept, header.ContentTypeEventStream)
			req.Header.Set("Mcp-Session-Id", sessionID)
			w := httptest.NewRecorder()
			app.ServeHTTP(w, req)
			assert.Equal(t, http.StatusForbidden, w.Code, method)
			cancel()
		}

		// A client without a user does not depend on the setting.
		clientOnly, err := entity.AddClientSession("mcp-test-client-only", 3600, "mcp", authn.GrantClientCredentials, nil)
		require.NoError(t, err)
		t.Cleanup(func() { _ = clientOnly.Delete() })
		assert.Equal(t, http.StatusOK, mcpPost(app, initBody, clientOnly.AuthToken(), "").Code)

		// The public session stands for no account, so the setting of the admin it carries does not apply.
		publicMCP := entity.Admin.MCP
		t.Cleanup(func() { entity.Admin.MCP = publicMCP })
		entity.Admin.MCP = -1
		conf.Options().Public = true
		assert.Equal(t, http.StatusOK, mcpPost(app, initBody, "", "").Code)
	})
	t.Run("RequestTooLargeContentLength", func(t *testing.T) {
		// The handler must reject oversized POST bodies with a 413 before
		// the MCP SDK reads them into memory. This branch covers the
		// early-reject path that fires when the client sends a
		// Content-Length header larger than MaxMCPRequestBytes.
		app, router, _ := NewApiTest()
		prepareMCPTest(t)
		ServeMCP(router)

		authToken := AuthenticateAdmin(app, router)

		req, _ := http.NewRequest(http.MethodPost, "/api/v1/mcp", strings.NewReader(""))
		header.SetAuthorization(req, authToken)
		req.Header.Set(header.ContentType, header.ContentTypeJson)
		req.Header.Set(header.Accept, header.ContentTypeJson+", "+header.ContentTypeEventStream)
		req.ContentLength = MaxMCPRequestBytes + 1

		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
	})
	t.Run("RequestTooLargeStreamed", func(t *testing.T) {
		// A body larger than the cap without a Content-Length header, as sent with chunked transfer
		// encoding, is answered with the standard 413 before the SDK reads it, in public mode too.
		for _, public := range []bool{false, true} {
			app, router, _ := NewApiTest()
			conf := prepareMCPTest(t)
			ServeMCP(router)

			authToken := AuthenticateAdmin(app, router)
			conf.Options().Public = public

			oversizedBody := strings.Repeat("a", int(MaxMCPRequestBytes)+1024)
			req, _ := http.NewRequest(http.MethodPost, "/api/v1/mcp", strings.NewReader(oversizedBody))
			header.SetAuthorization(req, authToken)
			req.Header.Set(header.ContentType, header.ContentTypeJson)
			req.Header.Set(header.Accept, header.ContentTypeJson+", "+header.ContentTypeEventStream)
			req.ContentLength = -1

			audit := captureAuditLog(t)
			w := httptest.NewRecorder()
			app.ServeHTTP(w, req)

			assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
			assert.NotContains(t, w.Body.String(), "request body exceeds")
			require.NotNil(t, audit.LastEntry())
			assert.Contains(t, audit.LastEntry().Message, "mcp › request body too large › failed")
		}
	})
	t.Run("NoChallengeWhenAppsDisabled", func(t *testing.T) {
		app, router, _ := NewApiTest()
		conf := prepareMCPTest(t)
		conf.Options().OAuthApps = "none"
		ServeMCP(router)

		r := mcpPost(app, `{"jsonrpc":"2.0","id":1,"method":"initialize"}`, "", "")
		assert.Equal(t, http.StatusUnauthorized, r.Code)
		assert.Empty(t, r.Header().Get(header.WWWAuthenticate))
	})
	t.Run("ChallengeWhenAppsEnabled", func(t *testing.T) {
		app, router, _ := NewApiTest()
		conf := prepareMCPTest(t)
		conf.Options().OAuthApps = "registered"
		ServeMCP(router)

		r := mcpPost(app, `{"jsonrpc":"2.0","id":1,"method":"initialize"}`, "", "")
		assert.Equal(t, http.StatusUnauthorized, r.Code)
		assert.Equal(t, `Bearer resource_metadata="`+conf.OAuthResourceMetadataUrl()+`"`, r.Header().Get(header.WWWAuthenticate))
		assert.Contains(t, r.Body.String(), "error")

		r = mcpPost(app, `{"jsonrpc":"2.0","id":1,"method":"initialize"}`, "sessxkkcxxxz0000000000000000000000000000000000000000000000000000", "")
		assert.Equal(t, http.StatusUnauthorized, r.Code)
		assert.Contains(t, r.Header().Get(header.WWWAuthenticate), `error="invalid_token"`)

		authToken := AuthenticateUser(app, router, "gandalf", "Gandalf123!")
		r = mcpPost(app, `{"jsonrpc":"2.0","id":1,"method":"initialize"}`, authToken, "")
		assert.Equal(t, http.StatusForbidden, r.Code)
		assert.Empty(t, r.Header().Get(header.WWWAuthenticate))
	})
	t.Run("PublicModeWithAppsEnabled", func(t *testing.T) {
		app, router, _ := NewApiTest()
		conf := prepareMCPTest(t)
		conf.Options().Public = true
		conf.Options().OAuthApps = "registered"
		ServeMCP(router)

		initBody := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0"}}}`
		w := mcpPost(app, initBody, "", "")
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Empty(t, w.Header().Get(header.WWWAuthenticate))
		sessionID := w.Header().Get("Mcp-Session-Id")
		assert.NotEmpty(t, sessionID)

		w2 := mcpPost(app, `{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"list_config_keys","arguments":{"query":"database","limit":5}}}`, "", sessionID)
		assert.Equal(t, http.StatusOK, w2.Code)
	})
	t.Run("SessionBoundToUser", func(t *testing.T) {
		app, router, _ := NewApiTest()
		prepareMCPTest(t)
		ServeMCP(router)

		adminToken := AuthenticateAdmin(app, router)

		alice := entity.FindUserByName("alice")
		if alice == nil {
			t.Fatal("admin fixture user 'alice' must not be nil")
		}

		adminSess := Session("", adminToken)
		if adminSess == nil {
			t.Fatal("admin session must not be nil")
		}

		assert.NotEqual(t, alice.UserUID, adminSess.UserUID)

		aliceSess, err := entity.AddClientSession("mcp-test-session-binding", 3600, "mcp", authn.GrantCLI, alice)
		if err != nil {
			t.Fatal(err)
		}

		t.Cleanup(func() { _ = aliceSess.Delete() })

		initBody := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0"}}}`
		w := mcpPost(app, initBody, adminToken, "")
		assert.Equal(t, http.StatusOK, w.Code)
		sessionID := w.Header().Get("Mcp-Session-Id")
		assert.NotEmpty(t, sessionID)

		// The other user is authorized for MCP and can open a session of their own.
		w1 := mcpPost(app, initBody, aliceSess.AuthToken(), "")
		assert.Equal(t, http.StatusOK, w1.Code)
		assert.NotEmpty(t, w1.Header().Get("Mcp-Session-Id"))
		assert.NotEqual(t, sessionID, w1.Header().Get("Mcp-Session-Id"))

		callBody := `{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"list_config_keys","arguments":{"query":"database","limit":5}}}`

		// Another user presenting the session ID is refused.
		audit := captureAuditLog(t)
		w2 := mcpPost(app, callBody, aliceSess.AuthToken(), sessionID)
		assert.Equal(t, http.StatusForbidden, w2.Code)
		assert.Contains(t, w2.Body.String(), "session user mismatch")
		require.NotNil(t, audit.LastEntry())
		assert.Contains(t, audit.LastEntry().Message, "mcp › request refused › denied")

		// The same token reaches the session through the X-Auth-Token header.
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/mcp", strings.NewReader(callBody))
		req.Header.Set(header.XAuthToken, adminToken)
		req.Header.Set(header.ContentType, header.ContentTypeJson)
		req.Header.Set(header.Accept, header.ContentTypeJson+", "+header.ContentTypeEventStream)
		req.Header.Set("Mcp-Session-Id", sessionID)
		w3 := httptest.NewRecorder()
		app.ServeHTTP(w3, req)
		assert.Equal(t, http.StatusOK, w3.Code)
		assert.Contains(t, w3.Body.String(), "matches")

		// The X-Auth-Token header takes precedence over a bearer token of another user.
		req, _ = http.NewRequest(http.MethodPost, "/api/v1/mcp", strings.NewReader(callBody))
		req.Header.Set(header.XAuthToken, aliceSess.AuthToken())
		header.SetAuthorization(req, adminToken)
		req.Header.Set(header.ContentType, header.ContentTypeJson)
		req.Header.Set(header.Accept, header.ContentTypeJson+", "+header.ContentTypeEventStream)
		req.Header.Set("Mcp-Session-Id", sessionID)
		w4 := httptest.NewRecorder()
		app.ServeHTTP(w4, req)
		assert.Equal(t, http.StatusForbidden, w4.Code)
		assert.Contains(t, w4.Body.String(), "session user mismatch")

		// The owner still reaches the session with the bearer token.
		w5 := mcpPost(app, callBody, adminToken, sessionID)
		assert.Equal(t, http.StatusOK, w5.Code)
		assert.Contains(t, w5.Body.String(), "matches")
	})
	t.Run("InitializeAndCallTools", func(t *testing.T) {
		app, router, _ := NewApiTest()
		prepareMCPTest(t)
		ServeMCP(router)

		authToken := AuthenticateAdmin(app, router)

		// Initialize MCP session.
		initBody := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0"}}}`
		w := mcpPost(app, initBody, authToken, "")

		assert.Equal(t, http.StatusOK, w.Code)
		sessionID := w.Header().Get("Mcp-Session-Id")
		assert.NotEmpty(t, sessionID)
		assert.Contains(t, w.Body.String(), "photoprism-mcp")

		// Send initialized notification (status varies by SDK version).
		w2 := mcpPost(app, `{"jsonrpc":"2.0","method":"notifications/initialized"}`, authToken, sessionID)
		assert.Less(t, w2.Code, 300, "notification should succeed, got %d: %s", w2.Code, w2.Body.String())

		// Call list_config_keys.
		w3 := mcpPost(app, `{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"list_config_keys","arguments":{"query":"database","limit":5}}}`, authToken, sessionID)
		assert.Equal(t, http.StatusOK, w3.Code)
		assert.Contains(t, w3.Body.String(), "matches")

		// Call find_search_filters.
		w4 := mcpPost(app, `{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"find_search_filters","arguments":{"query":"color","limit":5}}}`, authToken, sessionID)
		assert.Equal(t, http.StatusOK, w4.Code)
		assert.Contains(t, w4.Body.String(), "matches")

		assertMCPReferenceTools(t, app, authToken, sessionID)
	})
}
