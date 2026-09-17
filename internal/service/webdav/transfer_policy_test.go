package webdav

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/webdav"

	"github.com/photoprism/photoprism/pkg/fs"
)

// TestSkipSyncPath retains traversal handling while excluding hidden transfer paths.
func TestSkipSyncPath(t *testing.T) {
	for _, name := range append(fs.ReservedPathNames(), ".git/photo.jpg", ".config/photo.jpg", "folder/.photoprism/photo.jpg", ".ssh/../photo.jpg", "a/../.env.txt", ".hidden/photo.jpg", "photo.jpg.rclonelink", "nested/LINK.RCLONELINK/photo.jpg", `a\.HG\photo.jpg`) {
		assert.True(t, SkipSyncPath(name), name)
	}
	for _, name := range []string{"", "/", "photo.jpg", "photos/photo.jpg", "photo.rclonelink.jpg", "../outside.jpg", "/sub/../../outside.jpg", "%2egit/photo.jpg"} {
		assert.False(t, SkipSyncPath(name), name)
	}
}

// TestClient_TransferPolicy checks final transfer guards without contacting excluded paths.
func TestClient_TransferPolicy(t *testing.T) {
	remote := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(remote, "source.txt"), []byte("download-control"), fs.ModeFile))

	var requests atomic.Int64

	handler := &webdav.Handler{Prefix: "/.config", FileSystem: webdav.Dir(remote), LockSystem: webdav.NewMemLS()}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { requests.Add(1); handler.ServeHTTP(w, r) }))

	t.Cleanup(server.Close)

	client, err := NewClient(server.URL+"/.config/", "", "", TimeoutLow, "")

	require.NoError(t, err)

	local := filepath.Join(t.TempDir(), ".config", "originals")

	require.NoError(t, fs.MkdirAll(local))

	src := filepath.Join(local, "source.txt")

	require.NoError(t, os.WriteFile(src, []byte("upload-control"), fs.ModeFile))
	require.NoError(t, client.Upload(src, "uploaded.txt"))

	data, err := os.ReadFile(filepath.Join(remote, "uploaded.txt")) //nolint:gosec // Test reads its temporary transfer output.

	require.NoError(t, err)
	assert.Equal(t, "upload-control", string(data))

	dest := filepath.Join(local, "downloaded.txt")

	require.NoError(t, client.Download("source.txt", dest, false))

	data, err = os.ReadFile(dest) //nolint:gosec // Test reads its temporary transfer output.

	require.NoError(t, err)
	assert.Equal(t, "download-control", string(data))
	require.NoError(t, client.Mkdir("created"))
	require.NoError(t, client.MkdirAll("created/nested"))
	assert.DirExists(t, filepath.Join(remote, "created/nested"))
	require.NoError(t, client.Delete("uploaded.txt"))
	assert.NoFileExists(t, filepath.Join(remote, "uploaded.txt"))

	for _, name := range append(fs.ReservedPathNames(), "nested/_netrc", ".git/photo.jpg", "nested/.ssh/photo.jpg", ".config/photo.jpg", ".photoprism/photo.jpg", ".hidden/photo.jpg", "a/../.env.txt") {
		before := requests.Load()
		assert.ErrorIs(t, client.Upload(src, name), ErrSkipPath)
		assert.ErrorIs(t, client.Download(name, filepath.Join(local, "blocked.txt"), false), ErrSkipPath)
		assert.ErrorIs(t, client.MkdirAll(name), ErrSkipPath)
		assert.ErrorIs(t, client.Mkdir(name), ErrSkipPath)
		assert.ErrorIs(t, client.Delete(name), ErrSkipPath)

		files, err := client.Files(name, true)

		require.NoError(t, err)
		assert.Empty(t, files)

		dirs, err := client.Directories(name, true, 0)

		require.NoError(t, err)
		assert.Empty(t, dirs)
		assert.Equal(t, before, requests.Load(), name)
	}

	assert.NoFileExists(t, filepath.Join(local, "blocked.txt"))
}

// TestClient_DownloadDirTargets preserves operator-managed destination mappings.
func TestClient_DownloadDirTargets(t *testing.T) {
	remote := t.TempDir()
	require.NoError(t, fs.MkdirAll(filepath.Join(remote, "alias")))
	require.NoError(t, os.WriteFile(filepath.Join(remote, "alias/file.txt"), []byte("excluded-control"), fs.ModeFile))
	require.NoError(t, os.WriteFile(filepath.Join(remote, "ordinary.txt"), []byte("ordinary-control"), fs.ModeFile))
	handler := &webdav.Handler{FileSystem: webdav.Dir(remote), LockSystem: webdav.NewMemLS()}
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	client, err := NewClient(server.URL+"/", "", "", TimeoutLow, "")
	require.NoError(t, err)
	local := t.TempDir()
	require.NoError(t, fs.MkdirAll(filepath.Join(local, "operator-storage")))
	require.NoError(t, os.Symlink(filepath.Join(local, "operator-storage"), filepath.Join(local, "alias")))
	assert.Empty(t, client.DownloadDir("/", local, true, false))
	assert.FileExists(t, filepath.Join(local, "operator-storage/file.txt"))
	data, err := os.ReadFile(filepath.Join(local, "ordinary.txt")) //nolint:gosec // Test reads its temporary transfer output.
	require.NoError(t, err)
	assert.Equal(t, "ordinary-control", string(data))
}
