package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"

	"github.com/photoprism/photoprism/internal/config"
)

func TestOAuthUserinfo(t *testing.T) {
	t.Run("PublicMode", func(t *testing.T) {
		app, router, _ := NewApiTest()

		OAuthUserinfo(router)

		r := PerformRequest(app, http.MethodGet, "/api/v1/oauth/userinfo")
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
	t.Run("Authenticated", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		OAuthUserinfo(router)
		token := AuthenticateAdmin(app, router)

		r := AuthenticatedRequest(app, http.MethodGet, "/api/v1/oauth/userinfo", token)
		assert.Equal(t, http.StatusOK, r.Code, "body=%s", r.Body.String())
		assert.NotEmpty(t, gjson.Get(r.Body.String(), "sub").String(), "sub must be the account UID")
		assert.Equal(t, "admin", gjson.Get(r.Body.String(), "preferred_username").String())
	})
	t.Run("NoToken", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		OAuthUserinfo(router)

		r := PerformRequest(app, http.MethodGet, "/api/v1/oauth/userinfo")
		assert.Equal(t, http.StatusUnauthorized, r.Code)
	})
	t.Run("InvalidToken", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		OAuthUserinfo(router)

		r := AuthenticatedRequest(app, http.MethodGet, "/api/v1/oauth/userinfo", "invalidtoken")
		assert.Equal(t, http.StatusUnauthorized, r.Code)
	})
}
