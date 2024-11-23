package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
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
		assert.Equal(t, "<a href=\"/library/browse\">Temporary Redirect</a>.\n\n", w.Body.String())
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
		req, _ := http.NewRequest("GET", "/manifest.json", nil)
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
		req, _ := http.NewRequest("GET", "/sw.js", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
		assert.NotEmpty(t, w.Body)
	})
	t.Run("HeadServiceWorker", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("HEAD", "/sw.js", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
		assert.Empty(t, w.Body)
	})
}
