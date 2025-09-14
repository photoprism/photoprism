package api

import (
	"archive/zip"
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/service/cluster"
	"github.com/photoprism/photoprism/pkg/service/http/header"
)

func TestClusterGetTheme(t *testing.T) {
	t.Run("FeatureDisabled", func(t *testing.T) {
		app, router, conf := NewApiTest()
		// Ensure portal feature flag is disabled.
		conf.Options().NodeType = cluster.Instance
		ClusterGetTheme(router)

		r := PerformRequest(app, http.MethodGet, "/api/v1/cluster/theme")
		assert.Equal(t, http.StatusForbidden, r.Code)
	})

	t.Run("NotFound", func(t *testing.T) {
		app, router, conf := NewApiTest()
		// Enable portal feature flag for this endpoint.
		conf.Options().NodeType = cluster.Portal
		ClusterGetTheme(router)

		missing := filepath.Join(os.TempDir(), "photoprism-test-missing-theme")
		_ = os.RemoveAll(missing)
		conf.SetThemePath(missing)
		assert.False(t, fs.PathExists(conf.ThemePath()))

		req := httptest.NewRequest(http.MethodGet, "/api/v1/cluster/theme", nil)
		req.Header.Set("Accept", "application/json")
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Success", func(t *testing.T) {
		app, router, conf := NewApiTest()
		// Enable portal feature flag for this endpoint.
		conf.Options().NodeType = cluster.Portal
		ClusterGetTheme(router)

		tempTheme, err := os.MkdirTemp("", "pp-theme-*")
		assert.NoError(t, err)
		defer func() { _ = os.RemoveAll(tempTheme) }()
		conf.SetThemePath(tempTheme)

		assert.NoError(t, os.MkdirAll(filepath.Join(tempTheme, "sub"), 0o755))
		// Visible files
		assert.NoError(t, os.WriteFile(filepath.Join(tempTheme, "style.css"), []byte("body{}\n"), 0o644))
		assert.NoError(t, os.WriteFile(filepath.Join(tempTheme, "sub", "visible.txt"), []byte("ok\n"), 0o644))
		// Hidden file
		assert.NoError(t, os.WriteFile(filepath.Join(tempTheme, ".hidden.txt"), []byte("secret\n"), 0o644))
		// Hidden directory
		assert.NoError(t, os.MkdirAll(filepath.Join(tempTheme, ".git"), 0o755))
		assert.NoError(t, os.WriteFile(filepath.Join(tempTheme, ".git", "HEAD"), []byte("ref: refs/heads/main\n"), 0o644))
		// Hidden directory pattern "_.folder"
		assert.NoError(t, os.MkdirAll(filepath.Join(tempTheme, "_.folder"), 0o755))
		assert.NoError(t, os.WriteFile(filepath.Join(tempTheme, "_.folder", "secret.txt"), []byte("hidden\n"), 0o644))
		// Symlink (should be skipped); best-effort
		_ = os.Symlink(filepath.Join(tempTheme, "style.css"), filepath.Join(tempTheme, "link.css"))

		r := PerformRequest(app, http.MethodGet, "/api/v1/cluster/theme")
		assert.Equal(t, http.StatusOK, r.Code)

		// Verify headers
		assert.Equal(t, header.ContentTypeZip, r.Header().Get(header.ContentType))
		assert.Contains(t, r.Header().Get(header.ContentDisposition), "attachment; filename=theme.zip")

		// Verify zip contents
		body := r.Body.Bytes()
		zr, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
		assert.NoError(t, err)

		names := make([]string, 0, len(zr.File))
		for _, f := range zr.File {
			names = append(names, f.Name)
		}

		// Included
		assert.Contains(t, names, "style.css")
		// Subdirectories are not included for security reasons
		assert.NotContains(t, names, "sub/visible.txt")

		// Excluded (hidden files/dirs and symlinks)
		assert.NotContains(t, names, ".hidden.txt")
		assert.NotContains(t, names, ".git/HEAD")
		assert.NotContains(t, names, "_.folder/secret.txt")
		assert.NotContains(t, names, "link.css")
	})

	t.Run("Empty", func(t *testing.T) {
		app, router, conf := NewApiTest()
		// Enable portal feature flag for this endpoint.
		conf.Options().NodeType = cluster.Portal
		ClusterGetTheme(router)

		// Create an empty temporary theme directory (no includable files).
		tempTheme, err := os.MkdirTemp("", "pp-theme-empty-*")
		assert.NoError(t, err)
		defer func() { _ = os.RemoveAll(tempTheme) }()
		conf.SetThemePath(tempTheme)

		// Hidden-only content to ensure exclusion yields empty archive.
		assert.NoError(t, os.MkdirAll(filepath.Join(tempTheme, ".hidden-dir"), 0o755))
		assert.NoError(t, os.WriteFile(filepath.Join(tempTheme, ".hidden-dir", "file.txt"), []byte("secret\n"), 0o644))
		assert.NoError(t, os.WriteFile(filepath.Join(tempTheme, ".hidden"), []byte("secret\n"), 0o644))

		r := PerformRequest(app, http.MethodGet, "/api/v1/cluster/theme")
		assert.Equal(t, http.StatusOK, r.Code)

		// Verify headers
		assert.Equal(t, header.ContentTypeZip, r.Header().Get(header.ContentType))
		assert.Contains(t, r.Header().Get(header.ContentDisposition), "attachment; filename=theme.zip")

		// Verify zip is valid and empty (no files included)
		body := r.Body.Bytes()
		zr, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
		assert.NoError(t, err)
		assert.Equal(t, 0, len(zr.File))
	})
}
