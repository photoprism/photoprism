package api

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"

	"github.com/photoprism/photoprism/internal/config"
)

// TestOAuthClusterHooks verifies the shared /api/v1/oauth/* handlers delegate to
// the injected cluster hooks when set, and run their default path otherwise.
func TestOAuthClusterHooks(t *testing.T) {
	t.Run("AuthorizeDelegates", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		OAuthClusterAuthorize = func(c *gin.Context) bool {
			c.JSON(http.StatusOK, gin.H{"handled": "cluster-authorize"})
			return true
		}
		defer func() { OAuthClusterAuthorize = nil }()
		OAuthAuthorize(router)

		r := PerformRequest(app, http.MethodGet, "/api/v1/oauth/authorize?response_type=code&client_id=cs5gfen1bgxz7s9i")
		assert.Equal(t, http.StatusOK, r.Code)
		assert.Equal(t, "cluster-authorize", gjson.Get(r.Body.String(), "handled").String())
	})
	t.Run("TokenDelegates", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		OAuthClusterToken = func(c *gin.Context) bool {
			c.JSON(http.StatusOK, gin.H{"handled": "cluster-token"})
			return true
		}
		defer func() { OAuthClusterToken = nil }()
		OAuthToken(router)

		w := postOAuthToken(app, url.Values{"grant_type": {"authorization_code"}})
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "cluster-token", gjson.Get(w.Body.String(), "handled").String())
	})
	t.Run("TokenDefersWhenFalse", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		called := false
		OAuthClusterToken = func(c *gin.Context) bool { called = true; return false }
		defer func() { OAuthClusterToken = nil }()
		OAuthToken(router)

		// An incomplete authorization_code request falls through to CE, which
		// rejects it — proving the hook deferred rather than short-circuited.
		w := postOAuthToken(app, url.Values{"grant_type": {"authorization_code"}})
		assert.True(t, called, "hook must be consulted")
		assert.NotEqual(t, http.StatusOK, w.Code)
	})
	t.Run("UserinfoDelegates", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		OAuthClusterUserinfo = func(c *gin.Context) bool {
			c.JSON(http.StatusOK, gin.H{"handled": "cluster-userinfo"})
			return true
		}
		defer func() { OAuthClusterUserinfo = nil }()
		OAuthUserinfo(router)

		r := PerformRequest(app, http.MethodGet, "/api/v1/oauth/userinfo")
		assert.Equal(t, http.StatusOK, r.Code)
		assert.Equal(t, "cluster-userinfo", gjson.Get(r.Body.String(), "handled").String())
	})
}
