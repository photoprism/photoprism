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
	webdavPath := conf.BaseUri(WebDAVOriginals)

	r.GET(proxyPath, func(c *gin.Context) {
		c.String(http.StatusOK, "proxy")
	})
	r.GET(regularPath, func(c *gin.Context) {
		c.String(http.StatusOK, "regular")
	})
	r.Handle(header.MethodPropfind, webdavPath, func(c *gin.Context) {
		c.String(http.StatusMultiStatus, "webdav")
	})

	doRequest := func(method, path string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(method, path, nil)
		r.ServeHTTP(w, req)
		return w
	}

	t.Run("SkipsHeadersForProxyPrefix", func(t *testing.T) {
		w := doRequest(header.MethodGet, proxyPath)

		require.Equal(t, http.StatusOK, w.Code)
		assert.Empty(t, w.Header().Get(header.ContentSecurityPolicy))
		assert.Empty(t, w.Header().Get(header.XFrameOptions))
		assert.Empty(t, w.Header().Get(header.RobotsTag))
	})
	t.Run("AddsHeadersForNonProxyPath", func(t *testing.T) {
		w := doRequest(header.MethodGet, regularPath)

		require.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, header.DefaultContentSecurityPolicy, w.Header().Get(header.ContentSecurityPolicy))
		assert.Equal(t, header.DefaultFrameOptions, w.Header().Get(header.XFrameOptions))
		assert.Equal(t, header.RobotsNone, w.Header().Get(header.RobotsTag))
	})
	t.Run("SkipsBrowserHeadersForWebDAVPath", func(t *testing.T) {
		w := doRequest(header.MethodPropfind, webdavPath)

		require.Equal(t, http.StatusMultiStatus, w.Code)
		assert.Empty(t, w.Header().Get(header.ContentSecurityPolicy))
		assert.Empty(t, w.Header().Get(header.XFrameOptions))
		assert.Equal(t, header.RobotsNone, w.Header().Get(header.RobotsTag))
	})
}

func TestSecurityMiddlewareRobotsTag(t *testing.T) {
	gin.SetMode(gin.TestMode)

	conf := config.TestConfig()

	origDemo := conf.Options().Demo
	origPublic := conf.Options().Public
	origAuthMode := conf.Options().AuthMode

	t.Cleanup(func() {
		conf.Options().Demo = origDemo
		conf.Options().Public = origPublic
		conf.Options().AuthMode = origAuthMode
	})

	r := gin.New()
	r.Use(Security(conf))

	landingPath := conf.FrontendUri("")
	browsePath := conf.FrontendUri("/browse")
	albumPath := conf.FrontendUri("/albums/as6sg6bxpogaaba8/view")

	for _, p := range []string{landingPath, landingPath + "/", browsePath, albumPath} {
		r.GET(p, func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
		})
	}

	robotsTag := func(path string) string {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(header.MethodGet, path, nil))
		require.Equal(t, http.StatusOK, w.Code)
		return w.Header().Get(header.RobotsTag)
	}

	t.Run("PublicDemoLandingPageIndexable", func(t *testing.T) {
		conf.Options().Demo = true

		assert.Empty(t, robotsTag(landingPath))
		assert.Empty(t, robotsTag(landingPath+"/"))
	})
	t.Run("PublicDemoOtherRoutesNoIndex", func(t *testing.T) {
		conf.Options().Demo = true

		assert.Equal(t, header.RobotsNoIndex, robotsTag(browsePath))
		assert.Equal(t, header.RobotsNoIndex, robotsTag(albumPath))
	})
	t.Run("NonDemoNeverIndexable", func(t *testing.T) {
		conf.Options().Demo = false
		conf.Options().Public = true

		assert.Equal(t, header.RobotsNone, robotsTag(landingPath+"/"))
		assert.Equal(t, header.RobotsNone, robotsTag(browsePath))
	})
	t.Run("NonPublicNeverIndexable", func(t *testing.T) {
		conf.Options().Demo = false
		conf.Options().Public = false
		conf.Options().AuthMode = config.AuthModePasswd

		assert.Equal(t, header.RobotsNone, robotsTag(landingPath+"/"))
		assert.Equal(t, header.RobotsNone, robotsTag(browsePath))
	})
}
