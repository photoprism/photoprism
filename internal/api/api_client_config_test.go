package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/http/header"
)

func TestGetClientConfig(t *testing.T) {
	t.Run("Public", func(t *testing.T) {
		app, router, _ := NewApiTest()
		GetClientConfig(router)
		r := PerformRequest(app, "GET", "/api/v1/config")
		val := gjson.Get(r.Body.String(), "mode")
		assert.Equal(t, http.StatusOK, r.Code)
		assert.Equal(t, "user", val.String())
	})
	t.Run("Unauthorized", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		GetClientConfig(router)
		r := AuthenticatedRequest(app, "GET", "/api/v1/config", "")
		val := gjson.Get(r.Body.String(), "mode")
		assert.Equal(t, http.StatusOK, r.Code)
		assert.Equal(t, "public", val.String())
	})
	t.Run("FrontendDisabled", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		conf.Options().DisableFrontend = true
		GetClientConfig(router)
		r := PerformRequest(app, "GET", "/api/v1/config")
		assert.Equal(t, http.StatusUnauthorized, r.Code)
		conf.Options().DisableFrontend = false
	})
	t.Run("PortalJWT", func(t *testing.T) {
		fx := newPortalJWTFixture(t, "client-config-handler")

		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		GetClientConfig(router)

		spec := fx.defaultClaimsSpec()
		spec.Scope = []string{acl.ResourceCluster.String(), acl.ResourceConfig.String()}

		token := fx.issue(t, spec)

		req, _ := http.NewRequest(http.MethodGet, "/api/v1/config", nil)
		req.RemoteAddr = "10.10.0.5:1234"
		header.SetAuthorization(req, token)
		req.Header.Set(header.UserAgent, "PhotoPrism Portal/1.0")

		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "user", gjson.Get(w.Body.String(), "mode").String())
	})
}
