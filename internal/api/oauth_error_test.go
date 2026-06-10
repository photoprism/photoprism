package api

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/http/header"
)

func TestOAuthWantsHTML(t *testing.T) {
	newCtx := func(headers map[string]string) *gin.Context {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/oauth/authorize", nil)
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		c.Request = req
		return c
	}
	t.Run("NavigateHeader", func(t *testing.T) {
		assert.True(t, oauthWantsHTML(newCtx(map[string]string{header.FetchMode: "navigate"})))
	})
	t.Run("AcceptHTML", func(t *testing.T) {
		assert.True(t, oauthWantsHTML(newCtx(map[string]string{header.Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8"})))
	})
	t.Run("AcceptJSON", func(t *testing.T) {
		assert.False(t, oauthWantsHTML(newCtx(map[string]string{header.Accept: "application/json"})))
	})
	t.Run("NoHeaders", func(t *testing.T) {
		assert.False(t, oauthWantsHTML(newCtx(nil)))
	})
	t.Run("NilContext", func(t *testing.T) {
		assert.False(t, oauthWantsHTML(nil))
	})
}

func TestRenderOAuthError(t *testing.T) {
	t.Run("JSONForApiClient", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/oauth/authorize", nil)
		req.Header.Set(header.Accept, "application/json")
		c.Request = req
		RenderOAuthError(c, http.StatusNotFound, "invalid_client", "unknown client")
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "invalid_client")
		assert.Contains(t, w.Body.String(), "unknown client")
		assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
		assert.Equal(t, header.CacheControlNoStore, w.Header().Get(header.CacheControl))
	})
	t.Run("HTMLForBrowser", func(t *testing.T) {
		app := gin.New()
		app.LoadHTMLFiles(get.Config().TemplateFiles()...)
		app.GET("/x", func(c *gin.Context) {
			RenderOAuthError(c, http.StatusForbidden, "access_denied", "redirect_uri not registered")
		})
		req, _ := http.NewRequest(http.MethodGet, "/x", nil)
		req.Header.Set(header.FetchMode, "navigate")
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Contains(t, w.Header().Get("Content-Type"), "text/html")
		assert.Contains(t, w.Body.String(), "redirect_uri not registered")
		assert.Contains(t, w.Body.String(), "access_denied")
	})
	t.Run("NilContext", func(t *testing.T) {
		assert.NotPanics(t, func() { RenderOAuthError(nil, http.StatusBadRequest, "x", "y") })
	})
}

func TestRedirectOAuthError(t *testing.T) {
	newCtx := func() (*gin.Context, *httptest.ResponseRecorder) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/oauth/authorize", nil)
		c.Request = req
		return c, w
	}
	t.Run("RedirectsWithErrorAndState", func(t *testing.T) {
		c, w := newCtx()
		RedirectOAuthError(c, "https://photos.example.com/api/v1/oidc/redirect", "s-123", "access_denied", "no access to the requested instance")
		require.Equal(t, http.StatusFound, w.Code)
		loc, err := url.Parse(w.Header().Get("Location"))
		require.NoError(t, err)
		assert.Equal(t, "photos.example.com", loc.Host)
		assert.Equal(t, "access_denied", loc.Query().Get("error"))
		assert.Equal(t, "no access to the requested instance", loc.Query().Get("error_description"))
		assert.Equal(t, "s-123", loc.Query().Get("state"))
	})
	t.Run("PreservesExistingQueryAndOmitsEmptyState", func(t *testing.T) {
		c, w := newCtx()
		RedirectOAuthError(c, "https://photos.example.com/cb?tenant=acme", "", "server_error", "")
		loc, _ := url.Parse(w.Header().Get("Location"))
		assert.Equal(t, "acme", loc.Query().Get("tenant"))
		assert.Equal(t, "server_error", loc.Query().Get("error"))
		assert.False(t, loc.Query().Has("state"), "an empty state must not be echoed")
		assert.False(t, loc.Query().Has("error_description"), "an empty description must not be added")
	})
	t.Run("MalformedURIFallsBackToJSON", func(t *testing.T) {
		c, w := newCtx() // no Accept header → JSON branch
		RedirectOAuthError(c, "://not-a-url", "s1", "access_denied", "x")
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid_request")
	})
	t.Run("NilContext", func(t *testing.T) {
		assert.NotPanics(t, func() { RedirectOAuthError(nil, "https://x/", "s", "e", "d") })
	})
}
