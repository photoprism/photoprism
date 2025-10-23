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

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/service/cluster"
	reg "github.com/photoprism/photoprism/internal/service/cluster/registry"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestClusterGetTheme(t *testing.T) {
	t.Run("FeatureDisabled", func(t *testing.T) {
		app, router, conf := NewApiTest()
		// Ensure portal feature flag is disabled.
		conf.Options().NodeRole = cluster.RoleInstance
		ClusterGetTheme(router)

		r := PerformRequest(app, http.MethodGet, "/api/v1/cluster/theme")
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
	t.Run("NotFound", func(t *testing.T) {
		app, router, conf := NewApiTest()
		// Enable portal feature flag for this endpoint.
		conf.Options().NodeRole = cluster.RolePortal
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
		conf.Options().NodeRole = cluster.RolePortal
		ClusterGetTheme(router)

		tempTheme, err := os.MkdirTemp("", "pp-theme-*")
		assert.NoError(t, err)
		defer func() { _ = os.RemoveAll(tempTheme) }()
		conf.SetThemePath(tempTheme)

		assert.NoError(t, os.MkdirAll(filepath.Join(tempTheme, "sub"), fs.ModeDir))
		// Visible files
		assert.NoError(t, os.WriteFile(filepath.Join(tempTheme, "app.js"), []byte("console.log('ok')\n"), fs.ModeFile))
		assert.NoError(t, os.WriteFile(filepath.Join(tempTheme, "style.css"), []byte("body{}\n"), fs.ModeFile))
		assert.NoError(t, os.WriteFile(filepath.Join(tempTheme, fs.VersionTxtFile), []byte(" 1.0.0\n"), fs.ModeFile))
		assert.NoError(t, os.WriteFile(filepath.Join(tempTheme, "sub", "visible.txt"), []byte("ok\n"), fs.ModeFile))
		// Hidden file
		assert.NoError(t, os.WriteFile(filepath.Join(tempTheme, ".hidden.txt"), []byte("secret\n"), fs.ModeFile))
		// Hidden directory
		assert.NoError(t, os.MkdirAll(filepath.Join(tempTheme, ".git"), fs.ModeDir))
		assert.NoError(t, os.WriteFile(filepath.Join(tempTheme, ".git", "HEAD"), []byte("ref: refs/heads/main\n"), fs.ModeFile))
		// Hidden directory pattern "_.folder"
		assert.NoError(t, os.MkdirAll(filepath.Join(tempTheme, "_.folder"), fs.ModeDir))
		assert.NoError(t, os.WriteFile(filepath.Join(tempTheme, "_.folder", "secret.txt"), []byte("hidden\n"), fs.ModeFile))
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
		conf.Options().NodeRole = cluster.RolePortal
		ClusterGetTheme(router)

		// Create an empty temporary theme directory (no includable files).
		tempTheme, err := os.MkdirTemp("", "pp-theme-empty-*")
		assert.NoError(t, err)
		defer func() { _ = os.RemoveAll(tempTheme) }()
		conf.SetThemePath(tempTheme)

		// Hidden-only content and no app.js should yield 404.
		assert.NoError(t, os.MkdirAll(filepath.Join(tempTheme, ".hidden-dir"), fs.ModeDir))
		assert.NoError(t, os.WriteFile(filepath.Join(tempTheme, ".hidden-dir", "file.txt"), []byte("secret\n"), fs.ModeFile))
		assert.NoError(t, os.WriteFile(filepath.Join(tempTheme, ".hidden"), []byte("secret\n"), fs.ModeFile))

		req := httptest.NewRequest(http.MethodGet, "/api/v1/cluster/theme", nil)
		req.Header.Set("Accept", "application/json")
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
	t.Run("CIDRAllowWithoutAuth", func(t *testing.T) {
		app, router, conf := NewApiTest()
		// Enable portal role and set CIDR to loopback/10.0.0.0/8 for test.
		conf.Options().NodeRole = cluster.RolePortal
		conf.Options().ClusterCIDR = "10.0.0.0/8"
		ClusterGetTheme(router)

		tempTheme, err := os.MkdirTemp("", "pp-theme-cidr-*")
		assert.NoError(t, err)
		defer func() { _ = os.RemoveAll(tempTheme) }()
		conf.SetThemePath(tempTheme)
		assert.NoError(t, os.WriteFile(filepath.Join(tempTheme, "app.js"), []byte("console.log('ok')\n"), fs.ModeFile))
		assert.NoError(t, os.WriteFile(filepath.Join(tempTheme, "style.css"), []byte("body{}\n"), fs.ModeFile))

		req := httptest.NewRequest(http.MethodGet, "/api/v1/cluster/theme", nil)
		// Simulate request from 10.1.2.3
		req.RemoteAddr = "10.1.2.3:12345"
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, header.ContentTypeZip, w.Header().Get(header.ContentType))
	})
	t.Run("UpdateThemeVersion", func(t *testing.T) {
		app, _, conf := NewApiTest()
		_ = app // unused
		conf.Options().NodeRole = cluster.RolePortal

		regy, err := reg.NewClientRegistryWithConfig(conf)
		assert.NoError(t, err)

		node := &reg.Node{Node: cluster.Node{Name: "pp-node-01", Role: "instance", UUID: rnd.UUIDv7()}}
		assert.NoError(t, regy.Put(node))

		client := entity.FindClientByUID(node.ClientID)
		sess := entity.NewSession(-1, -1)
		sess.SetClient(client)
		sess.RefID = "sess-test"

		updateNodeThemeVersion(conf, sess, " theme-v1 \n", "127.0.0.1", sess.RefID)

		stored, err := regy.Get(node.UUID)
		assert.NoError(t, err)
		assert.Equal(t, "theme-v1", stored.Theme)
	})
}
