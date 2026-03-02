package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/http/proxy"
)

func TestSecurityMiddlewareSkipsPortalProxy(t *testing.T) {
	gin.SetMode(gin.TestMode)

	conf := config.TestConfig()

	r := gin.New()
	r.Use(Security(conf))

	proxyPath := conf.BaseUri(proxy.PathPrefix + "test" + conf.FrontendUri("/login"))
	regularPath := conf.FrontendUri("/login")

	r.GET(proxyPath, func(c *gin.Context) {
		c.String(http.StatusOK, "proxy")
	})
	r.GET(regularPath, func(c *gin.Context) {
		c.String(http.StatusOK, "regular")
	})

	doRequest := func(path string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(header.MethodGet, path, nil)
		r.ServeHTTP(w, req)
		return w
	}

	t.Run("SkipsHeadersForProxyPrefix", func(t *testing.T) {
		w := doRequest(proxyPath)

		require.Equal(t, http.StatusOK, w.Code)
		assert.Empty(t, w.Header().Get(header.ContentSecurityPolicy))
		assert.Empty(t, w.Header().Get(header.XFrameOptions))
		assert.Empty(t, w.Header().Get(header.RobotsTag))
	})
	t.Run("AddsHeadersForNonProxyPath", func(t *testing.T) {
		w := doRequest(regularPath)

		require.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, header.DefaultContentSecurityPolicy, w.Header().Get(header.ContentSecurityPolicy))
		assert.Equal(t, header.DefaultFrameOptions, w.Header().Get(header.XFrameOptions))
		assert.Equal(t, header.RobotsNone, w.Header().Get(header.RobotsTag))
	})
}
