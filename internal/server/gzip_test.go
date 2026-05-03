package server

import (
	"bytes"
	stdgzip "compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/http/proxy"
)

func TestShouldExcludeGzipExt(t *testing.T) {
	t.Run("ExcludesCompressedFormats", func(t *testing.T) {
		excludedExts := []string{".png", ".gif", ".jpeg", ".jpg", ".webp", ".mp3", ".mp4", ".zip", ".gz"}
		for _, ext := range excludedExts {
			assert.True(t, ShouldExcludeGzipExt(ext), "extension %s should be excluded", ext)
		}
	})
	t.Run("ExcludesCaseInsensitive", func(t *testing.T) {
		assert.True(t, ShouldExcludeGzipExt(".PNG"))
		assert.True(t, ShouldExcludeGzipExt(".Jpg"))
		assert.True(t, ShouldExcludeGzipExt(".GZ"))
	})
	t.Run("DoesNotExcludeTextFormats", func(t *testing.T) {
		allowedExts := []string{".html", ".css", ".js", ".json", ".xml", ".txt", ".svg"}
		for _, ext := range allowedExts {
			assert.False(t, ShouldExcludeGzipExt(ext), "extension %s should not be excluded", ext)
		}
	})
	t.Run("EmptyExtension", func(t *testing.T) {
		assert.False(t, ShouldExcludeGzipExt(""))
	})
}

func TestMatchesPrefixExclusion(t *testing.T) {
	prefixes := []string{"/api/v1/dl", "/health", "/livez"}

	t.Run("MatchesPrefix", func(t *testing.T) {
		assert.True(t, matchesPrefixExclusion("/api/v1/dl/file", prefixes))
		assert.True(t, matchesPrefixExclusion("/health", prefixes))
		assert.True(t, matchesPrefixExclusion("/healthz", prefixes))
		assert.True(t, matchesPrefixExclusion("/livez", prefixes))
	})
	t.Run("DoesNotMatchNonPrefixes", func(t *testing.T) {
		assert.False(t, matchesPrefixExclusion("/api/v1/photos", prefixes))
		assert.False(t, matchesPrefixExclusion("/other", prefixes))
	})
	t.Run("EmptyPrefixes", func(t *testing.T) {
		assert.False(t, matchesPrefixExclusion("/anything", nil))
		assert.False(t, matchesPrefixExclusion("/anything", []string{}))
	})
	t.Run("EmptyPath", func(t *testing.T) {
		assert.False(t, matchesPrefixExclusion("", prefixes))
	})
}

func TestMatchesFallbackExclusion(t *testing.T) {
	clusterTheme := "/api/v1/cluster/theme"
	photoDl := "/api/v1/photos/"
	share := "/s/"

	t.Run("MatchesClusterTheme", func(t *testing.T) {
		assert.True(t, matchesFallbackExclusion(clusterTheme, clusterTheme, photoDl, share))
	})
	t.Run("MatchesPhotoDl", func(t *testing.T) {
		assert.True(t, matchesFallbackExclusion("/api/v1/photos/abc123/dl", clusterTheme, photoDl, share))
	})
	t.Run("MatchesSharePreview", func(t *testing.T) {
		assert.True(t, matchesFallbackExclusion("/s/token/shared/preview", clusterTheme, photoDl, share))
	})
	t.Run("DoesNotMatchOtherPaths", func(t *testing.T) {
		assert.False(t, matchesFallbackExclusion("/api/v1/photos/abc123", clusterTheme, photoDl, share))
		assert.False(t, matchesFallbackExclusion("/s/token/shared/view", clusterTheme, photoDl, share))
		assert.False(t, matchesFallbackExclusion("/library/something/preview", clusterTheme, photoDl, share))
	})
}

func TestGzipMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Enable gzip for this test router.
	conf := config.TestConfig()
	conf.Options().HttpCompression = "gzip"

	r := gin.New()
	r.Use(gzip.Gzip(
		gzip.DefaultCompression,
		gzip.WithCustomShouldCompressFn(NewGzipShouldCompressFn(conf)),
	))

	r.GET("/ok", func(c *gin.Context) {
		c.String(http.StatusOK, "hello world")
	})

	excludedPath := conf.BaseUri(config.ApiUri + "/dl/test")
	r.GET(excludedPath, func(c *gin.Context) {
		c.String(http.StatusOK, "download")
	})

	livezPath := conf.BaseUri("/livez")
	r.GET(livezPath, func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	healthzPath := conf.BaseUri("/healthz")
	r.GET(healthzPath, func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	readyzPath := conf.BaseUri("/readyz")
	r.GET(readyzPath, func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	imagePath := "/file.jpg"
	r.GET(imagePath, func(c *gin.Context) {
		c.String(http.StatusOK, "image")
	})

	photoDlRoute := conf.BaseUri(config.ApiUri + "/photos/:uid/dl")
	r.GET(photoDlRoute, func(c *gin.Context) {
		c.String(http.StatusOK, "photo")
	})

	clusterThemeRoute := conf.BaseUri(config.ApiUri + "/cluster/theme")
	r.GET(clusterThemeRoute, func(c *gin.Context) {
		c.String(http.StatusOK, "theme")
	})

	sharePreviewRoute := conf.BaseUri("/s/:token/:shared/preview")
	r.GET(sharePreviewRoute, func(c *gin.Context) {
		c.String(http.StatusOK, "preview")
	})

	doRequest := func(path string, acceptGzip bool) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", path, nil)
		if acceptGzip {
			req.Header.Set("Accept-Encoding", "gzip")
		}
		r.ServeHTTP(w, req)
		return w
	}

	t.Run("CompressesSuccessfulResponse", func(t *testing.T) {
		w := doRequest("/ok", true)

		require.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "gzip", w.Header().Get("Content-Encoding"))

		zr, err := stdgzip.NewReader(bytes.NewReader(w.Body.Bytes()))
		require.NoError(t, err)
		defer func() {
			require.NoError(t, zr.Close())
		}()

		b, err := io.ReadAll(zr)
		require.NoError(t, err)
		assert.Equal(t, "hello world", string(b))
	})
	t.Run("DoesNotCompressExcludedPrefixes", func(t *testing.T) {
		w := doRequest(excludedPath, true)

		require.Equal(t, http.StatusOK, w.Code)
		assert.Empty(t, w.Header().Get("Content-Encoding"))
		assert.Equal(t, "download", w.Body.String())
	})
	t.Run("DoesNotCompressExcludedExtensions", func(t *testing.T) {
		w := doRequest(imagePath, true)

		require.Equal(t, http.StatusOK, w.Code)
		assert.Empty(t, w.Header().Get("Content-Encoding"))
		assert.Equal(t, "image", w.Body.String())
	})
	t.Run("DoesNotCompressHealthEndpoints", func(t *testing.T) {
		for _, path := range []string{livezPath, healthzPath, readyzPath} {
			w := doRequest(path, true)

			require.Equal(t, http.StatusOK, w.Code, path)
			assert.Empty(t, w.Header().Get("Content-Encoding"), path)
			assert.Equal(t, "ok", w.Body.String(), path)
		}
	})
	t.Run("DoesNotCompressPhotoOriginalDownload", func(t *testing.T) {
		w := doRequest(conf.BaseUri(config.ApiUri+"/photos/abc/dl"), true)

		require.Equal(t, http.StatusOK, w.Code)
		assert.Empty(t, w.Header().Get("Content-Encoding"))
		assert.Equal(t, "photo", w.Body.String())
	})
	t.Run("DoesNotCompressClusterThemeDownload", func(t *testing.T) {
		w := doRequest(clusterThemeRoute, true)

		require.Equal(t, http.StatusOK, w.Code)
		assert.Empty(t, w.Header().Get("Content-Encoding"))
		assert.Equal(t, "theme", w.Body.String())
	})
	t.Run("DoesNotCompressSharePreview", func(t *testing.T) {
		w := doRequest(conf.BaseUri("/s/tok/shared/preview"), true)

		require.Equal(t, http.StatusOK, w.Code)
		assert.Empty(t, w.Header().Get("Content-Encoding"))
		assert.Equal(t, "preview", w.Body.String())
	})
	t.Run("DoesNotCompressPortalProxyPrefix", func(t *testing.T) {
		proxyPath := conf.BaseUri(proxy.PathPrefix + "test/ok")
		r.GET(proxyPath, func(c *gin.Context) {
			c.String(http.StatusOK, "proxy")
		})

		w := doRequest(proxyPath, true)

		require.Equal(t, http.StatusOK, w.Code)
		assert.Empty(t, w.Header().Get("Content-Encoding"))
		assert.Equal(t, "proxy", w.Body.String())
	})
	t.Run("DoesNotCompressWithoutAcceptEncoding", func(t *testing.T) {
		w := doRequest("/ok", false)

		require.Equal(t, http.StatusOK, w.Code)
		assert.Empty(t, w.Header().Get("Content-Encoding"))
		assert.Equal(t, "hello world", w.Body.String())
	})
	t.Run("DoesNotCompressNotFound", func(t *testing.T) {
		w := doRequest("/missing", true)

		require.Equal(t, http.StatusNotFound, w.Code)
		assert.Empty(t, w.Header().Get("Content-Encoding"))
		assert.Contains(t, w.Body.String(), "404")
	})
}
