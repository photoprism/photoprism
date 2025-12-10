package server

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWebDAVSetFavoriteFlag_CreatesYamlOnce(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "img.jpg")
	assert.NoError(t, os.WriteFile(file, []byte("x"), 0o600))
	// First call creates YAML
	WebDAVSetFavoriteFlag(file)
	// YAML is written next to file without the media extension (AbsPrefix)
	yml := filepath.Join(filepath.Dir(file), "img.yml")
	assert.FileExists(t, yml)
	// Write a marker and ensure second call doesn't overwrite content
	// #nosec G304 -- test reads file created in a temp directory.
	orig, _ := os.ReadFile(yml)
	WebDAVSetFavoriteFlag(file)
	// #nosec G304 -- test reads file created in a temp directory.
	now, _ := os.ReadFile(yml)
	assert.Equal(t, string(orig), string(now))
}

func TestWebDAVSetFileMtime_NoFuture(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "a.txt")
	assert.NoError(t, os.WriteFile(file, []byte("x"), 0o600))
	// Set a past mtime
	WebDAVSetFileMtime(file, 946684800) // 2000-01-01 UTC
	after, _ := os.Stat(file)
	// Compare seconds to avoid platform-specific rounding
	got := after.ModTime().Unix()
	assert.Equal(t, int64(946684800), got)
}
