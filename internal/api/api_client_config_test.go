package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
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
	t.Run("ForbiddenFromCDN", func(t *testing.T) {
		app, router, _ := NewApiTest()
		GetClientConfig(router)

		req, _ := http.NewRequest(http.MethodGet, "/api/v1/config", nil)
		req.Header.Set(header.CdnHost, "edge.example")
		req.Header.Set(header.Accept, "application/json")

		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

// TestUpdateClientConfig pins the contract that PreviewToken and
// DownloadToken are stripped from the global "config.updated" broadcast
// so a per-session value cannot overwrite another session's tokens. The
// session-bound copy still carries the live tokens via ClientSession;
// see internal/config/client_config.go and the docs/comments on
// UpdateClientConfig.
func TestUpdateClientConfig(t *testing.T) {
	t.Run("OmitsTokensFromBroadcast", func(t *testing.T) {
		_, _, conf := NewApiTest()

		// The base config carries no tokens of its own, as ClientSession is the only
		// place a session token is assigned.
		direct := conf.ClientUser(false)
		require.Empty(t, direct.PreviewToken)
		require.Empty(t, direct.DownloadToken)

		// Prove the strip is still load-bearing rather than passing on an empty default: a config
		// that does carry tokens must lose them before the broadcast.
		seeded := conf.ClientUser(false)
		seeded.PreviewToken = "preview-token"
		seeded.DownloadToken = "download-token"
		stripBroadcastTokens(seeded)
		require.Empty(t, seeded.PreviewToken)
		require.Empty(t, seeded.DownloadToken)

		sub := event.Subscribe("config.updated")
		defer event.Unsubscribe(sub)

		UpdateClientConfig()

		select {
		case msg := <-sub.Receiver:
			assert.Equal(t, "config.updated", msg.Topic())

			payload, ok := msg.Fields["config"].(*config.ClientConfig)
			require.True(t, ok, "config field must be *config.ClientConfig, got %T", msg.Fields["config"])

			assert.Empty(t, payload.PreviewToken, "PreviewToken must be stripped from the global config.updated broadcast")
			assert.Empty(t, payload.DownloadToken, "DownloadToken must be stripped from the global config.updated broadcast")
		case <-time.After(5 * time.Second):
			t.Fatal("timed out waiting for config.updated event")
		}
	})
}
