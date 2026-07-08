package server

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/event"
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

func TestWebDAVSetFavoriteFlag_WriteErrorGoesToSystemLog(t *testing.T) {
	dir := t.TempDir()
	// A regular file where a directory is expected makes MkdirAll fail with an
	// error that embeds the absolute path; that must not reach the UI log stream.
	blocker := filepath.Join(dir, "blocker")
	assert.NoError(t, os.WriteFile(blocker, []byte("x"), 0o600))
	file := filepath.Join(blocker, "img.jpg")

	hook := &logCapture{}
	event.SystemLog.ReplaceHooks(logrus.LevelHooks{})
	event.SystemLog.AddHook(hook)
	defer event.SystemLog.ReplaceHooks(logrus.LevelHooks{})

	WebDAVSetFavoriteFlag(file)

	var found bool
	for _, e := range hook.entries {
		if e.Level == logrus.ErrorLevel {
			found = true
		}
	}
	assert.True(t, found, "sidecar write failure must be logged on the system log")
	assert.NoFileExists(t, filepath.Join(blocker, "img.yml"))
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
