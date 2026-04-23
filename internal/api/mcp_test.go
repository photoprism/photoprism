package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/photoprism/get"
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

// TestServeMCP exercises the HTTP handler installed by ServeMCP: the
// public-mode anonymous path, unauthenticated and non-admin denial, and
// the full admin round-trip through initialize, notifications/initialized,
// and tools/call.
func TestServeMCP(t *testing.T) {
	t.Run("AllowedPublicMode", func(t *testing.T) {
		// In public mode, Session() returns the default public session and
		// the currently registered MCP tools only surface static reference
		// data. Anonymous callers must therefore be able to initialize an
		// MCP session and call both tools without a token — this is what
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
		// MaxBytesReader plus the response-writer wrapper must convert the
		// SDK's 400 "failed to read body" into the standard 413 when the
		// client streams a body larger than the cap without a
		// Content-Length header. This path is what protects deployments
		// that sit behind proxies using chunked transfer encoding.
		app, router, _ := NewApiTest()
		prepareMCPTest(t)
		ServeMCP(router)

		authToken := AuthenticateAdmin(app, router)

		oversizedBody := strings.Repeat("a", int(MaxMCPRequestBytes)+1024)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/mcp", strings.NewReader(oversizedBody))
		header.SetAuthorization(req, authToken)
		req.Header.Set(header.ContentType, header.ContentTypeJson)
		req.Header.Set(header.Accept, header.ContentTypeJson+", "+header.ContentTypeEventStream)
		// Force chunked transfer encoding semantics by clearing the
		// Content-Length signal, matching the behavior of a streaming
		// upstream proxy or client that sends Transfer-Encoding: chunked.
		req.ContentLength = -1

		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
		// The rewritten response must not leak the SDK's 400 phrasing.
		assert.NotContains(t, w.Body.String(), "failed to read body")
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

// TestMcpLimitReader exercises the body wrapper that sets the shared
// "tripped" flag when http.MaxBytesReader returns its sentinel error.
func TestMcpLimitReader(t *testing.T) {
	t.Run("FlagsOnMaxBytesError", func(t *testing.T) {
		rec := httptest.NewRecorder()
		body := io.NopCloser(strings.NewReader("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"))
		tripped := &atomic.Bool{}
		r := &mcpLimitReader{
			ReadCloser: http.MaxBytesReader(rec, body, 4),
			tripped:    tripped,
		}
		buf := make([]byte, 16)
		n, err := r.Read(buf)
		assert.Less(t, n, 16)
		assert.Error(t, err)
		assert.True(t, tripped.Load())
	})
	t.Run("NoFlagOnNormalRead", func(t *testing.T) {
		rec := httptest.NewRecorder()
		body := io.NopCloser(strings.NewReader("hello"))
		tripped := &atomic.Bool{}
		r := &mcpLimitReader{
			ReadCloser: http.MaxBytesReader(rec, body, 1024),
			tripped:    tripped,
		}
		buf := make([]byte, 64)
		n, _ := r.Read(buf)
		assert.Equal(t, 5, n)
		assert.False(t, tripped.Load())
	})
}

// TestMcpLimitWriter exercises the response-writer wrapper that translates
// the SDK's 400 into a 413 once the paired reader has tripped.
func TestMcpLimitWriter(t *testing.T) {
	t.Run("RewritesBadRequestWhenTripped", func(t *testing.T) {
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		tripped := &atomic.Bool{}
		tripped.Store(true)
		w := &mcpLimitWriter{ResponseWriter: c.Writer, tripped: tripped}
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("failed to read body\n"))
		assert.Equal(t, http.StatusRequestEntityTooLarge, c.Writer.Status())
		assert.True(t, w.suppress.Load())
		// The SDK-provided body must not reach the client on the 413
		// rewrite path; only the status code is preserved.
		assert.NotContains(t, rec.Body.String(), "failed to read body")
	})
	t.Run("PassesThroughWhenNotTripped", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		tripped := &atomic.Bool{}
		w := &mcpLimitWriter{ResponseWriter: c.Writer, tripped: tripped}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
		assert.Equal(t, http.StatusOK, c.Writer.Status())
		assert.False(t, w.suppress.Load())
	})
	t.Run("LeavesNonBadRequestAlone", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		tripped := &atomic.Bool{}
		tripped.Store(true)
		w := &mcpLimitWriter{ResponseWriter: c.Writer, tripped: tripped}
		w.WriteHeader(http.StatusInternalServerError)
		assert.Equal(t, http.StatusInternalServerError, c.Writer.Status())
		assert.False(t, w.suppress.Load())
	})
}
