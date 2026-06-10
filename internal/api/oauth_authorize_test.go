package api

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
)

func TestOAuthAuthorize(t *testing.T) {
	t.Run("PublicMode", func(t *testing.T) {
		app, router, _ := NewApiTest()

		OAuthAuthorize(router)

		r := PerformRequest(app, http.MethodGet, "/api/v1/oauth/authorize")
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
	t.Run("DefaultNotImplemented", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		OAuthAuthorize(router)

		r := PerformRequest(app, http.MethodGet, "/api/v1/oauth/authorize")
		assert.Equal(t, http.StatusMethodNotAllowed, r.Code)
	})
	t.Run("OverrideHookDelegates", func(t *testing.T) {
		app, router, _ := NewApiTest()

		OAuthAuthorize(router)

		prev := OAuthAuthorizeHandler
		OAuthAuthorizeHandler = func(c *gin.Context) { c.String(http.StatusOK, "delegated") }
		defer func() { OAuthAuthorizeHandler = prev }()

		r := PerformRequest(app, http.MethodGet, "/api/v1/oauth/authorize")
		assert.Equal(t, http.StatusOK, r.Code)
		assert.Equal(t, "delegated", r.Body.String())
	})
}
