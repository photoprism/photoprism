package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/http/header"
)

func TestStaticRoutes(t *testing.T) {
	// Create router.
	r := gin.Default()

	// Get test config.
	conf := config.TestConfig()

	// Find and load templates.
	r.LoadHTMLFiles(conf.TemplateFiles()...)

	// Register routes.
	registerStaticRoutes(r, conf)

	t.Run("GetRoot", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, 307, w.Code)
		assert.Equal(t, "<a href=\"/library/\">Temporary Redirect</a>.\n\n", w.Body.String())
	})
	t.Run("HeadRoot", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("HEAD", "/", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, 307, w.Code)
	})
}

func TestWebAppRoutes(t *testing.T) {
	// Create router.
	r := gin.Default()

	// Get test config.
	conf := config.TestConfig()

	// Find and load templates.
	r.LoadHTMLFiles(conf.TemplateFiles()...)

	// Register user interface routes.
	registerWebAppRoutes(r, conf)

	// Bootstrapping.
	t.Run("GetLibrary", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", conf.LibraryUri("/"), nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
		assert.NotEmpty(t, w.Body)
	})
	t.Run("HeadLibrary", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("HEAD", conf.LibraryUri("/"), nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
		assert.NotEmpty(t, w.Body)
	})
	t.Run("GetLibraryBrowse", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", conf.LibraryUri("/browse"), nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
		assert.NotEmpty(t, w.Body)
	})
	t.Run("HeadLibraryBrowse", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("HEAD", conf.LibraryUri("/browse"), nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
	})
	t.Run("GetManifest", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/"+fs.ManifestJsonFile, nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
		assert.NotEmpty(t, w.Body.String())
		manifest := w.Body.String()
		t.Logf("PWA Manifest: %s", manifest)
		assert.True(t, strings.Contains(manifest, `"scope": "/",`))
		assert.True(t, strings.Contains(manifest, `"start_url": "/library/",`))
		assert.True(t, strings.Contains(manifest, "/static/icons/logo/128.png"))
	})
	t.Run("GetServiceWorker", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/"+fs.SwJsFile, nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
		assert.NotEmpty(t, w.Body)
		assert.Equal(t, header.CacheControlNoStore, w.Header().Get(header.CacheControl))
		assert.Contains(t, w.Header().Get(header.ContentType), "javascript")
	})
	t.Run("HeadServiceWorker", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("HEAD", "/"+fs.SwJsFile, nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
		assert.Empty(t, w.Body)
		assert.Equal(t, header.CacheControlNoStore, w.Header().Get(header.CacheControl))
		assert.Equal(t, header.ContentTypeJavaScript, w.Header().Get(header.ContentType))
	})
	t.Run("GetServiceWorkerScopeCleanup", func(t *testing.T) {
		scopeCleanupFile := conf.StaticBuildFile(fs.SwScopeCleanupJsFile)
		require.NoError(t, os.MkdirAll(filepath.Dir(scopeCleanupFile), fs.ModeDir))
		require.NoError(t, os.WriteFile(scopeCleanupFile, []byte(`self.addEventListener("activate", () => {});`), fs.ModeFile))
		require.FileExists(t, scopeCleanupFile)
		t.Cleanup(func() { _ = os.Remove(scopeCleanupFile) })

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/"+fs.SwScopeCleanupJsFile, nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
		assert.NotEmpty(t, w.Body.String())
		assert.Equal(t, header.CacheControlNoStore, w.Header().Get(header.CacheControl))
		assert.Contains(t, w.Header().Get(header.ContentType), "javascript")
	})
	t.Run("HeadServiceWorkerScopeCleanup", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("HEAD", "/"+fs.SwScopeCleanupJsFile, nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
		assert.Empty(t, w.Body.String())
		assert.Equal(t, header.CacheControlNoStore, w.Header().Get(header.CacheControl))
		assert.Equal(t, header.ContentTypeJavaScript, w.Header().Get(header.ContentType))
	})
	t.Run("ServiceWorkerFallbackAndBaseUri", func(t *testing.T) {
		swConf := config.NewMinimalTestConfig(t.TempDir())
		swConf.Options().AssetsPath = t.TempDir()
		swConf.Options().SiteUrl = "https://portal.example.com/p/acme/"

		swRouter := gin.New()
		registerWebAppRoutes(swRouter, swConf)

		type getCase struct {
			name     string
			path     string
			expected string
		}

		getCases := []getCase{
			{name: "GetServiceWorkerRootFallback", path: "/" + fs.SwJsFile, expected: string(fallbackServiceWorker)},
			{name: "GetServiceWorkerBaseUriFallback", path: swConf.BaseUri("/" + fs.SwJsFile), expected: string(fallbackServiceWorker)},
			{name: "GetScopeCleanupRootFallback", path: "/" + fs.SwScopeCleanupJsFile, expected: string(fallbackScopeCleanupScript)},
			{name: "GetScopeCleanupBaseUriFallback", path: swConf.BaseUri("/" + fs.SwScopeCleanupJsFile), expected: string(fallbackScopeCleanupScript)},
		}

		for _, tc := range getCases {
			t.Run(tc.name, func(t *testing.T) {
				w := httptest.NewRecorder()
				req := httptest.NewRequest(header.MethodGet, tc.path, nil)
				swRouter.ServeHTTP(w, req)
				assert.Equal(t, http.StatusOK, w.Code)
				assert.Equal(t, tc.expected, w.Body.String())
				assert.Equal(t, header.CacheControlNoStore, w.Header().Get(header.CacheControl))
				assert.Equal(t, header.ContentTypeJavaScript, w.Header().Get(header.ContentType))
			})
		}

		headCases := []struct {
			name string
			path string
		}{
			{name: "HeadServiceWorkerRootFallback", path: "/" + fs.SwJsFile},
			{name: "HeadServiceWorkerBaseUriFallback", path: swConf.BaseUri("/" + fs.SwJsFile)},
			{name: "HeadScopeCleanupRootFallback", path: "/" + fs.SwScopeCleanupJsFile},
			{name: "HeadScopeCleanupBaseUriFallback", path: swConf.BaseUri("/" + fs.SwScopeCleanupJsFile)},
		}

		for _, tc := range headCases {
			t.Run(tc.name, func(t *testing.T) {
				w := httptest.NewRecorder()
				req := httptest.NewRequest(header.MethodHead, tc.path, nil)
				swRouter.ServeHTTP(w, req)
				assert.Equal(t, http.StatusOK, w.Code)
				assert.Empty(t, w.Body.String())
				assert.Equal(t, header.CacheControlNoStore, w.Header().Get(header.CacheControl))
				assert.Equal(t, header.ContentTypeJavaScript, w.Header().Get(header.ContentType))
			})
		}
	})
	t.Run("GetWorkboxHelperRoot", func(t *testing.T) {
		workboxFile := conf.StaticBuildFile("workbox-123abc.js")
		require.NoError(t, os.MkdirAll(filepath.Dir(workboxFile), fs.ModeDir))
		require.NoError(t, os.WriteFile(workboxFile, []byte(`console.log("workbox");`), fs.ModeFile))
		require.FileExists(t, workboxFile)
		t.Cleanup(func() { _ = os.Remove(workboxFile) })

		h := newWorkboxHandler(conf)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/workbox-123abc.js", nil)
		c.Params = gin.Params{gin.Param{Key: "hash", Value: "123abc.js"}}

		h(c)
		assert.Equal(t, 200, w.Code)
		assert.NotEmpty(t, w.Body)
	})
	t.Run("GetWorkboxHelperBaseUri", func(t *testing.T) {
		workboxPath := conf.BaseUri("/workbox-123abc.js")
		if workboxPath == "/workbox-123abc.js" {
			return
		}

		workboxFile := conf.StaticBuildFile("workbox-123abc.js")
		require.NoError(t, os.MkdirAll(filepath.Dir(workboxFile), fs.ModeDir))
		require.NoError(t, os.WriteFile(workboxFile, []byte(`console.log("workbox");`), fs.ModeFile))
		require.FileExists(t, workboxFile)
		t.Cleanup(func() { _ = os.Remove(workboxFile) })

		h := newWorkboxHandler(conf)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", workboxPath, nil)
		c.Params = gin.Params{gin.Param{Key: "hash", Value: "123abc.js"}}

		h(c)
		assert.Equal(t, 200, w.Code)
		assert.NotEmpty(t, w.Body)
	})
}
