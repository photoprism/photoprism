package api

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
)

func TestOAuthUserinfo(t *testing.T) {
	t.Run("PublicMode", func(t *testing.T) {
		app, router, _ := NewApiTest()

		OAuthUserinfo(router)

		r := PerformRequest(app, http.MethodGet, "/api/v1/oauth/userinfo")
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
	t.Run("DefaultNotImplemented", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		OAuthUserinfo(router)

		r := PerformRequest(app, http.MethodGet, "/api/v1/oauth/userinfo")
		assert.Equal(t, http.StatusMethodNotAllowed, r.Code)
	})
	t.Run("PostRouteRegistered", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		OAuthUserinfo(router)

		// POST must route to the same handler per OIDC Core §5.3.1.
		r := PerformRequest(app, http.MethodPost, "/api/v1/oauth/userinfo")
		assert.Equal(t, http.StatusMethodNotAllowed, r.Code)
	})
	t.Run("OverrideHookDelegates", func(t *testing.T) {
		app, router, _ := NewApiTest()

		OAuthUserinfo(router)

		prev := OAuthUserinfoHandler
		OAuthUserinfoHandler = func(c *gin.Context) { c.String(http.StatusOK, "delegated:"+c.Request.Method) }
		defer func() { OAuthUserinfoHandler = prev }()

		get := PerformRequest(app, http.MethodGet, "/api/v1/oauth/userinfo")
		assert.Equal(t, http.StatusOK, get.Code)
		assert.Equal(t, "delegated:GET", get.Body.String())

		post := PerformRequest(app, http.MethodPost, "/api/v1/oauth/userinfo")
		assert.Equal(t, http.StatusOK, post.Code)
		assert.Equal(t, "delegated:POST", post.Body.String())
	})
}
