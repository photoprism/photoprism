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
	"github.com/photoprism/photoprism/internal/config/pwa"
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
		assert.Equal(t, "<a href=\""+conf.FrontendUri("/")+"\">Temporary Redirect</a>.\n\n", w.Body.String())
	})
	t.Run("HeadRoot", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("HEAD", "/", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, 307, w.Code)
	})
}

func TestStaticRoutesWebOverlay(t *testing.T) {
	t.Run("RootRedirectWithoutIndex", func(t *testing.T) {
		conf := config.NewMinimalTestConfig(t.TempDir())
		require.NoError(t, os.MkdirAll(conf.WebStoragePath(), fs.ModeDir))

		r := gin.New()
		r.LoadHTMLFiles(conf.TemplateFiles()...)
		registerStaticRoutes(r, conf)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(header.MethodGet, "/", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
		assert.Equal(t, conf.LoginUri(), w.Header().Get(header.Location))

		w = httptest.NewRecorder()
		req = httptest.NewRequest(header.MethodHead, "/", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
		assert.Equal(t, conf.LoginUri(), w.Header().Get(header.Location))
	})
	t.Run("ServeOverlayFileAndDirectoryIndex", func(t *testing.T) {
		conf := config.NewMinimalTestConfig(t.TempDir())
		webDir := conf.WebStoragePath()
		require.NoError(t, os.MkdirAll(filepath.Join(webDir, "docs"), fs.ModeDir))
		require.NoError(t, os.WriteFile(filepath.Join(webDir, "hello.txt"), []byte("hello"), fs.ModeFile))
		require.NoError(t, os.WriteFile(filepath.Join(webDir, "docs", IndexHtml), []byte("docs"), fs.ModeFile))

		r := gin.New()
		r.LoadHTMLFiles(conf.TemplateFiles()...)
		registerStaticRoutes(r, conf)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(header.MethodGet, "/hello.txt", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "hello", w.Body.String())

		w = httptest.NewRecorder()
		req = httptest.NewRequest(header.MethodGet, "/docs", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "docs", w.Body.String())

		w = httptest.NewRecorder()
		req = httptest.NewRequest(header.MethodHead, "/hello.txt", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Empty(t, w.Body.String())
	})
	t.Run("BasePathOverlayMapping", func(t *testing.T) {
		conf := config.NewMinimalTestConfig(t.TempDir())
		conf.Options().SiteUrl = "https://example.com/i/acme/"
		webDir := conf.WebStoragePath()
		require.NoError(t, os.MkdirAll(filepath.Join(webDir, "assets"), fs.ModeDir))
		require.NoError(t, os.WriteFile(filepath.Join(webDir, "assets", "app.js"), []byte("asset"), fs.ModeFile))

		r := gin.New()
		r.LoadHTMLFiles(conf.TemplateFiles()...)
		registerStaticRoutes(r, conf)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(header.MethodGet, conf.BaseUri("/assets/app.js"), nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "asset", w.Body.String())

		w = httptest.NewRecorder()
		req = httptest.NewRequest(header.MethodGet, "/assets/app.js", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
	t.Run("HiddenAndSpecialPathsBlocked", func(t *testing.T) {
		conf := config.NewMinimalTestConfig(t.TempDir())
		webDir := conf.WebStoragePath()
		require.NoError(t, os.MkdirAll(filepath.Join(webDir, "foo"), fs.ModeDir))
		require.NoError(t, os.MkdirAll(filepath.Join(webDir, "__MACOSX"), fs.ModeDir))
		require.NoError(t, os.WriteFile(filepath.Join(webDir, "env"), []byte("public-env"), fs.ModeFile))
		require.NoError(t, os.WriteFile(filepath.Join(webDir, "@secrets.txt"), []byte("secret"), fs.ModeFile))
		require.NoError(t, os.WriteFile(filepath.Join(webDir, "__MACOSX", "test.txt"), []byte("meta"), fs.ModeFile))
		require.NoError(t, os.WriteFile(filepath.Join(webDir, "foo", ".env"), []byte("hidden"), fs.ModeFile))
		require.NoError(t, os.WriteFile(filepath.Join(webDir, "public.txt"), []byte("ok"), fs.ModeFile))

		r := gin.New()
		r.LoadHTMLFiles(conf.TemplateFiles()...)
		registerStaticRoutes(r, conf)

		blocked := []string{
			"/.env",
			"/.htaccess",
			"/foo/.env",
			"/@secrets.txt",
			"/__MACOSX/test.txt",
		}

		for _, filePath := range blocked {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(header.MethodGet, filePath, nil)
			r.ServeHTTP(w, req)
			assert.Equal(t, http.StatusNotFound, w.Code, filePath)
		}

		w := httptest.NewRecorder()
		req := httptest.NewRequest(header.MethodGet, "/public.txt", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "ok", w.Body.String())
	})
	t.Run("SensitiveNamesBlocked", func(t *testing.T) {
		conf := config.NewMinimalTestConfig(t.TempDir())
		webDir := conf.WebStoragePath()
		require.NoError(t, os.MkdirAll(filepath.Join(webDir, "node", "secrets"), fs.ModeDir))
		require.NoError(t, os.MkdirAll(filepath.Join(webDir, "config", "portal"), fs.ModeDir))
		require.NoError(t, os.MkdirAll(filepath.Join(webDir, "config", "certificates"), fs.ModeDir))
		require.NoError(t, os.MkdirAll(filepath.Join(webDir, "tls"), fs.ModeDir))
		require.NoError(t, os.MkdirAll(filepath.Join(webDir, "db"), fs.ModeDir))
		require.NoError(t, os.MkdirAll(filepath.Join(webDir, "docs"), fs.ModeDir))
		require.NoError(t, os.WriteFile(filepath.Join(webDir, "options.yml"), []byte("blocked"), fs.ModeFile))
		require.NoError(t, os.WriteFile(filepath.Join(webDir, "Options.YML"), []byte("blocked"), fs.ModeFile))
		require.NoError(t, os.WriteFile(filepath.Join(webDir, "config.yaml"), []byte("blocked"), fs.ModeFile))
		require.NoError(t, os.WriteFile(filepath.Join(webDir, "id_rsa"), []byte("blocked"), fs.ModeFile))
		require.NoError(t, os.WriteFile(filepath.Join(webDir, "auth.json"), []byte("blocked"), fs.ModeFile))
		require.NoError(t, os.WriteFile(filepath.Join(webDir, "join_token"), []byte("blocked"), fs.ModeFile))
		require.NoError(t, os.WriteFile(filepath.Join(webDir, "tls", "server.pem"), []byte("blocked"), fs.ModeFile))
		require.NoError(t, os.WriteFile(filepath.Join(webDir, "db", "dump.sql"), []byte("blocked"), fs.ModeFile))
		require.NoError(t, os.WriteFile(filepath.Join(webDir, "docs", "public.toml"), []byte("blocked"), fs.ModeFile))
		require.NoError(t, os.WriteFile(filepath.Join(webDir, "docs", "client_secret"), []byte("blocked"), fs.ModeFile))
		require.NoError(t, os.WriteFile(filepath.Join(webDir, "node", "secrets", "token.txt"), []byte("blocked"), fs.ModeFile))
		require.NoError(t, os.WriteFile(filepath.Join(webDir, "config", "portal", "options.yml"), []byte("blocked"), fs.ModeFile))
		require.NoError(t, os.WriteFile(filepath.Join(webDir, "config", "certificates", "fullchain.pem"), []byte("blocked"), fs.ModeFile))
		require.NoError(t, os.WriteFile(filepath.Join(webDir, "docs", "public.txt"), []byte("ok"), fs.ModeFile))

		r := gin.New()
		r.LoadHTMLFiles(conf.TemplateFiles()...)
		registerStaticRoutes(r, conf)

		blocked := []string{
			"/options.yml",
			"/Options.YML",
			"/config.yaml",
			"/id_rsa",
			"/auth.json",
			"/join_token",
			"/tls/server.pem",
			"/db/dump.sql",
			"/docs/public.toml",
			"/docs/client_secret",
			"/node/secrets/token.txt",
			"/config/portal/options.yml",
			"/config/certificates/fullchain.pem",
		}

		for _, filePath := range blocked {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(header.MethodGet, filePath, nil)
			r.ServeHTTP(w, req)
			assert.Equal(t, http.StatusNotFound, w.Code, filePath)
		}

		w := httptest.NewRecorder()
		req := httptest.NewRequest(header.MethodGet, "/docs/public.txt", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "ok", w.Body.String())
	})
	t.Run("SymlinkEscapeBlocked", func(t *testing.T) {
		conf := config.NewMinimalTestConfig(t.TempDir())
		webDir := conf.WebStoragePath()
		outsideDir := filepath.Join(t.TempDir(), "outside")
		require.NoError(t, os.MkdirAll(webDir, fs.ModeDir))
		require.NoError(t, os.MkdirAll(outsideDir, fs.ModeDir))
		require.NoError(t, os.WriteFile(filepath.Join(webDir, "public.txt"), []byte("ok"), fs.ModeFile))
		require.NoError(t, os.WriteFile(filepath.Join(outsideDir, "secret.txt"), []byte("secret"), fs.ModeFile))

		if err := os.Symlink(filepath.Join(outsideDir, "secret.txt"), filepath.Join(webDir, "leak.txt")); err != nil {
			t.Skipf("symlink setup failed: %v", err)
		}

		r := gin.New()
		r.LoadHTMLFiles(conf.TemplateFiles()...)
		registerStaticRoutes(r, conf)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(header.MethodGet, "/public.txt", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "ok", w.Body.String())

		w = httptest.NewRecorder()
		req = httptest.NewRequest(header.MethodGet, "/leak.txt", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
	t.Run("EncodedAndAmbiguousPathsBlocked", func(t *testing.T) {
		conf := config.NewMinimalTestConfig(t.TempDir())
		webDir := conf.WebStoragePath()
		require.NoError(t, os.MkdirAll(filepath.Join(webDir, "docs"), fs.ModeDir))
		require.NoError(t, os.WriteFile(filepath.Join(webDir, "env"), []byte("public-env"), fs.ModeFile))
		require.NoError(t, os.WriteFile(filepath.Join(webDir, "docs", IndexHtml), []byte("docs"), fs.ModeFile))

		r := gin.New()
		r.LoadHTMLFiles(conf.TemplateFiles()...)
		registerStaticRoutes(r, conf)

		blocked := []string{
			"/%2eenv",
			"/docs//index.html",
			"/docs/../env",
		}

		for _, filePath := range blocked {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(header.MethodGet, filePath, nil)
			r.ServeHTTP(w, req)
			assert.Equal(t, http.StatusNotFound, w.Code, filePath)
		}
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
		req, _ := http.NewRequest("GET", conf.FrontendUri("/"), nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
		assert.NotEmpty(t, w.Body)
	})
	t.Run("HeadLibrary", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("HEAD", conf.FrontendUri("/"), nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
		assert.NotEmpty(t, w.Body)
	})
	t.Run("GetLibraryBrowse", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", conf.FrontendUri("/browse"), nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
		assert.NotEmpty(t, w.Body)
	})
	t.Run("HeadLibraryBrowse", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("HEAD", conf.FrontendUri("/browse"), nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
	})
	t.Run("GetManifest", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/"+fs.ManifestJsonFile, nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
		assert.NotEmpty(t, w.Body.String())
		assert.Equal(t, header.ContentTypeManifest, w.Header().Get(header.ContentType))
		manifest := w.Body.String()
		t.Logf("PWA Manifest: %s", manifest)
		assert.True(t, strings.Contains(manifest, `"scope": "/",`))
		assert.True(t, strings.Contains(manifest, `"start_url": "`+pwa.StartUrl(conf.BaseUri("/"), conf.FrontendUri(``))+`",`))
		assert.True(t, strings.Contains(manifest, `"url": "library/browse"`))
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
		swConf.Options().SiteUrl = "https://portal.example.com/i/acme/"

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

func TestWebAppManifestRouteWithBasePath(t *testing.T) {
	config.FlushCache()
	t.Cleanup(config.FlushCache)

	r := gin.New()
	conf := config.NewMinimalTestConfig(t.TempDir())
	conf.Options().SiteUrl = "https://app.localssl.dev/instance/pro-1/"

	registerWebAppRoutes(r, conf)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(header.MethodGet, conf.BaseUri("/"+fs.ManifestJsonFile), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, header.ContentTypeManifest, w.Header().Get(header.ContentType))
	assert.Contains(t, w.Body.String(), `"scope": "/instance/pro-1/",`)
	assert.Contains(t, w.Body.String(), `"start_url": "./library",`)
	assert.Contains(t, w.Body.String(), `"url": "library/browse"`)
}
