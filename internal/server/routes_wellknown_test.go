package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/http/header"
)

func TestRegisterWellknownRoutes(t *testing.T) {
	conf := config.TestConfig()
	require.False(t, conf.Portal())

	router := gin.New()
	registerWellknownRoutes(router, conf)

	// request sends a GET request to the router and returns the recorded response.
	request := func(path string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, conf.BaseUri(path), nil))
		return w
	}

	t.Run("OAuthAuthorizationServer", func(t *testing.T) {
		w := request("/.well-known/oauth-authorization-server")
		require.Equal(t, http.StatusOK, w.Code)

		var result map[string]any
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
		assert.NotEmpty(t, result["issuer"])
	})
	t.Run("OpenIDConfiguration", func(t *testing.T) {
		w := request("/.well-known/openid-configuration")
		require.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "max-age=300, public", w.Header().Get(header.CacheControl))

		var result map[string]any
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
		assert.NotEmpty(t, result["issuer"])
	})
	t.Run("OAuthProtectedResourceDisabled", func(t *testing.T) {
		opt := conf.Options()
		apps, public, authMode := opt.OAuthApps, opt.Public, opt.AuthMode
		t.Cleanup(func() { opt.OAuthApps, opt.Public, opt.AuthMode = apps, public, authMode })

		opt.OAuthApps, opt.Public, opt.AuthMode = "none", false, config.AuthModePasswd

		w := request("/.well-known/oauth-protected-resource/api/v1/mcp")
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
	t.Run("OAuthProtectedResource", func(t *testing.T) {
		opt := conf.Options()
		apps, public, authMode, disableMCP := opt.OAuthApps, opt.Public, opt.AuthMode, opt.DisableMCP
		t.Cleanup(func() { opt.OAuthApps, opt.Public, opt.AuthMode, opt.DisableMCP = apps, public, authMode, disableMCP })

		opt.OAuthApps, opt.Public, opt.AuthMode, opt.DisableMCP = "registered", false, config.AuthModePasswd, false

		w := request("/.well-known/oauth-protected-resource/api/v1/mcp")
		require.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, header.Any, w.Header().Get(header.AccessControlAllowOrigin))
		assert.Equal(t, "max-age=300, public", w.Header().Get(header.CacheControl))

		var result map[string]any
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
		assert.Equal(t, conf.OAuthResourceUrl(), result["resource"])
		assert.Equal(t, []any{conf.SiteUrl()}, result["authorization_servers"])
		assert.Equal(t, []any{"mcp"}, result["scopes_supported"])
		assert.Equal(t, []any{"header"}, result["bearer_methods_supported"])

		opt.DisableMCP = true
		assert.Equal(t, http.StatusNotFound, request("/.well-known/oauth-protected-resource/api/v1/mcp").Code)

		opt.DisableMCP, opt.Public = false, true
		assert.Equal(t, http.StatusNotFound, request("/.well-known/oauth-protected-resource/api/v1/mcp").Code)
	})
	t.Run("JWKSNotFoundOnInstance", func(t *testing.T) {
		w := request("/.well-known/jwks.json")
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Empty(t, w.Header().Get(header.ETag))
	})
}
