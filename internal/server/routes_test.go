package server

import (
	"net/http"
	"net/http/httptest"
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

	t.Run("GetHome", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, 307, w.Code)
		assert.Equal(t, "<a href=\"/library/browse\">Temporary Redirect</a>.\n\n", w.Body.String())
	})
	t.Run("HeadHome", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("HEAD", "/", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, 307, w.Code)
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
	t.Run("GetLibrary", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/library/", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
		assert.NotEmpty(t, w.Body)
	})
	t.Run("GetLibrary", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/library/", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
		assert.NotEmpty(t, w.Body)
	})
	t.Run("GetLibraryBrowse", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/library/browse", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
		assert.NotEmpty(t, w.Body)
	})
	t.Run("HeadLibraryBrowse", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("HEAD", "/library/browse", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
	})
}
