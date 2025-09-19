package api

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/service/cluster"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/service/http/header"
)

func TestClusterPermissions(t *testing.T) {
	t.Run("UnauthorizedWhenPublicDisabled", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.Options().NodeRole = cluster.RolePortal

		// Disable public mode so Auth requires a session.
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		ClusterSummary(router)

		r := PerformRequest(app, http.MethodGet, "/api/v1/cluster")
		assert.Equal(t, http.StatusUnauthorized, r.Code)
	})

	t.Run("ForbiddenFromCDN", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.Options().NodeRole = cluster.RolePortal

		ClusterListNodes(router)

		req, _ := http.NewRequest(http.MethodGet, "/api/v1/cluster/nodes", nil)
		// Mark as CDN request, which Auth() forbids.
		req.Header.Set("Cdn-Host", "edge.example")
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("AdminCanAccess", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.Options().NodeRole = cluster.RolePortal
		ClusterSummary(router)
		token := AuthenticateAdmin(app, router)
		r := AuthenticatedRequest(app, http.MethodGet, "/api/v1/cluster", token)
		assert.Equal(t, http.StatusOK, r.Code)
	})

	// Note: most fixture users have admin role; client-scope test below covers non-admin denial.

	t.Run("ClientInsufficientScope", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.Options().NodeRole = cluster.RolePortal
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		// Acquire client token with metrics scope (does not include cluster).
		OAuthToken(router)

		data := url.Values{
			"grant_type":    {authn.GrantClientCredentials.String()},
			"client_id":     {"cs5cpu17n6gj2qo5"},
			"client_secret": {"xcCbOrw6I0vcoXzhnOmXhjpVSyFq0l0e"},
			"scope":         {"metrics"},
		}
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/oauth/token", strings.NewReader(data.Encode()))
		req.Header.Add(header.ContentType, header.ContentTypeForm)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		token := gjson.Get(w.Body.String(), "access_token").String()

		ClusterSummary(router)
		r := AuthenticatedRequest(app, http.MethodGet, "/api/v1/cluster", token)
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
}
