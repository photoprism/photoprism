package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/service/cluster"
	"github.com/photoprism/photoprism/pkg/http/header"
)

func TestClusterHealth(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, router, conf := NewApiTest()
		enablePortalAPIs(t, conf)
		ClusterHealth(router)

		r := PerformRequest(app, http.MethodGet, "/api/v1/cluster/health")
		assert.Equal(t, http.StatusOK, r.Code)
		assert.Equal(t, header.CacheControlNoStore, r.Header().Get(header.CacheControl))
		assert.Equal(t, "", r.Header().Get(header.AccessControlAllowOrigin))
	})
	t.Run("FeatureDisabled", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.Options().NodeRole = cluster.RoleInstance
		ClusterHealth(router)

		r := PerformRequest(app, http.MethodGet, "/api/v1/cluster/health")
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
	t.Run("ClusterCIDRDenied", func(t *testing.T) {
		app, router, conf := NewApiTest()
		enablePortalAPIs(t, conf)
		conf.Options().ClusterCIDR = "192.0.2.0/24"
		ClusterHealth(router)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/cluster/health", nil)
		req.RemoteAddr = "198.51.100.9:12345"
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
	t.Run("ClusterCIDRList", func(t *testing.T) {
		app, router, conf := NewApiTest()
		enablePortalAPIs(t, conf)
		prevClusterCIDR := conf.Options().ClusterCIDR
		t.Cleanup(func() { conf.Options().ClusterCIDR = prevClusterCIDR })
		ClusterHealth(router)

		// request returns the status of a health check from the specified address.
		request := func(remoteAddr string) int {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/cluster/health", nil)
			req.RemoteAddr = remoteAddr
			w := httptest.NewRecorder()
			app.ServeHTTP(w, req)
			return w.Code
		}

		conf.Options().ClusterCIDR = "198.51.100.0/24, 192.0.2.0/24"
		assert.Equal(t, http.StatusOK, request("192.0.2.42:12345"))
		assert.Equal(t, http.StatusUnauthorized, request("203.0.113.9:12345"))
		conf.Options().ClusterCIDR = "192.0.2.0/24,garbage"
		assert.Equal(t, http.StatusUnauthorized, request("192.0.2.42:12345"))
	})
	t.Run("ClusterCIDRAllowed", func(t *testing.T) {
		app, router, conf := NewApiTest()
		enablePortalAPIs(t, conf)
		conf.Options().ClusterCIDR = "192.0.2.0/24"
		ClusterHealth(router)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/cluster/health", nil)
		req.RemoteAddr = "192.0.2.42:12345"
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
