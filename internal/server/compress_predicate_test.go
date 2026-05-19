package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/http/proxy"
)

func TestNewShouldCompressFn(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("NilConfigDeclines", func(t *testing.T) {
		fn := NewShouldCompressFn(nil)
		assert.False(t, fn(nil))
	})

	conf := config.TestConfig()
	fn := NewShouldCompressFn(conf)

	// Build a router so c.FullPath() resolves for the dynamic route exclusions.
	r := gin.New()
	check := func(c *gin.Context) {
		if fn(c) {
			c.Status(http.StatusOK)
		} else {
			c.Status(http.StatusNoContent)
		}
	}
	r.GET("/anything", check)
	r.GET("/file.jpg", check)
	r.GET(conf.BaseUri("/livez"), check)
	r.GET(conf.BaseUri("/health"), check)
	r.GET(conf.BaseUri("/readyz"), check)
	r.GET(conf.BaseUri(config.ApiUri+"/dl/test"), check)
	r.GET(conf.BaseUri(config.ApiUri+"/photos/:uid/dl"), check)
	r.GET(conf.BaseUri(config.ApiUri+"/cluster/theme"), check)
	r.GET(conf.BaseUri("/s/:token/:shared/preview"), check)
	r.GET(conf.BaseUri(proxy.PathPrefix+"test/ok"), check)
	r.GET(conf.BaseUri(config.StaticUri+"/build/app.js"), check)
	r.GET(conf.BaseUri(config.CustomStaticUri+"/foo.css"), check)

	doRequest := func(path string) int {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		r.ServeHTTP(w, req)
		return w.Code
	}

	t.Run("CompressesGenericPath", func(t *testing.T) {
		assert.Equal(t, http.StatusOK, doRequest("/anything"))
	})
	t.Run("ExcludesAlreadyCompressedExtensions", func(t *testing.T) {
		assert.Equal(t, http.StatusNoContent, doRequest("/file.jpg"))
	})
	t.Run("ExcludesHealthPrefixes", func(t *testing.T) {
		for _, p := range []string{"/livez", "/health", "/readyz"} {
			assert.Equalf(t, http.StatusNoContent, doRequest(conf.BaseUri(p)), "path=%s", p)
		}
	})
	t.Run("ExcludesApiDlPrefix", func(t *testing.T) {
		assert.Equal(t, http.StatusNoContent, doRequest(conf.BaseUri(config.ApiUri+"/dl/test")))
	})
	t.Run("ExcludesPhotoOriginalDownload", func(t *testing.T) {
		assert.Equal(t, http.StatusNoContent, doRequest(conf.BaseUri(config.ApiUri+"/photos/abc/dl")))
	})
	t.Run("ExcludesClusterTheme", func(t *testing.T) {
		assert.Equal(t, http.StatusNoContent, doRequest(conf.BaseUri(config.ApiUri+"/cluster/theme")))
	})
	t.Run("ExcludesSharePreview", func(t *testing.T) {
		assert.Equal(t, http.StatusNoContent, doRequest(conf.BaseUri("/s/tok/shared/preview")))
	})
	t.Run("ExcludesPortalProxy", func(t *testing.T) {
		assert.Equal(t, http.StatusNoContent, doRequest(conf.BaseUri(proxy.PathPrefix+"test/ok")))
	})
	t.Run("ExcludesBundledStatic", func(t *testing.T) {
		// /static/* is served by PrecompressedStatic, which serves precompressed
		// siblings inline; the runtime middleware must not double-encode.
		assert.Equal(t, http.StatusNoContent, doRequest(conf.BaseUri(config.StaticUri+"/build/app.js")))
	})
	t.Run("ExcludesCustomStatic", func(t *testing.T) {
		// Custom-extension static assets share the same precompressed pipeline.
		assert.Equal(t, http.StatusNoContent, doRequest(conf.BaseUri(config.CustomStaticUri+"/foo.css")))
	})

	t.Run("DoesNotCheckAcceptEncoding", func(t *testing.T) {
		// Predicate is encoding-agnostic — middleware owns the Accept-Encoding negotiation.
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/anything", nil)
		// No Accept-Encoding header, but the predicate should still return true.
		r.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)
	})
}
