package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/http/header"
)

// prepareMCPTest sets auth mode to password, disables public mode, and enables
// experimental mode so that MCP auth checks behave as in production.
func prepareMCPTest(t *testing.T) *config.Config {
	t.Helper()

	conf := get.Config()
	originalOptions := *conf.Options()

	t.Cleanup(func() {
		*conf.Options() = originalOptions
	})

	conf.Options().AuthMode = config.AuthModePasswd
	conf.Options().Public = false
	conf.Options().Experimental = true

	return conf
}

// mcpPost sends a JSON-RPC POST to /api/v1/mcp with MCP-required headers.
func mcpPost(app http.Handler, body, authToken, sessionID string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/mcp", strings.NewReader(body))

	if authToken != "" {
		header.SetAuthorization(req, authToken)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")

	if sessionID != "" {
		req.Header.Set("Mcp-Session-Id", sessionID)
	}

	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	return w
}

func TestServeMCP(t *testing.T) {
	t.Run("ForbiddenPublicMode", func(t *testing.T) {
		app, router, _ := NewApiTest()
		conf := prepareMCPTest(t)
		conf.Options().Public = true
		ServeMCP(router)

		r := PerformRequest(app, http.MethodPost, "/api/v1/mcp")
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
	t.Run("ForbiddenUnauthenticated", func(t *testing.T) {
		app, router, _ := NewApiTest()
		prepareMCPTest(t)
		ServeMCP(router)

		r := mcpPost(app, `{"jsonrpc":"2.0","id":1,"method":"initialize"}`, "", "")
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
	t.Run("ForbiddenNonAdmin", func(t *testing.T) {
		app, router, _ := NewApiTest()
		prepareMCPTest(t)
		ServeMCP(router)

		authToken := AuthenticateUser(app, router, "gandalf", "Gandalf123!")
		r := mcpPost(app, `{"jsonrpc":"2.0","id":1,"method":"initialize"}`, authToken, "")
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
	t.Run("NotFoundExperimentalOff", func(t *testing.T) {
		app, router, _ := NewApiTest()
		conf := prepareMCPTest(t)
		conf.Options().Experimental = false
		ServeMCP(router)

		authToken := AuthenticateAdmin(app, router)
		r := mcpPost(app, `{"jsonrpc":"2.0","id":1,"method":"initialize"}`, authToken, "")
		assert.Equal(t, http.StatusNotFound, r.Code)
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
	})
}
