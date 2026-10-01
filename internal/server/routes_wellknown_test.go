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
	t.Run("JWKSNotFoundOnInstance", func(t *testing.T) {
		w := request("/.well-known/jwks.json")
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Empty(t, w.Header().Get(header.ETag))
	})
}
