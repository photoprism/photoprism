package workers

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/pkg/fs"
)

func TestRunPurgeArchives(t *testing.T) {
	c := config.NewMinimalTestConfig(t.TempDir())

	// resetZipDir empties the archive directory. Config.TempPath() is cached process-wide, so every test
	// config resolves to the same directory and subtests would otherwise see each other's archives.
	resetZipDir := func(t *testing.T) string {
		t.Helper()
		zipPath := filepath.Join(c.TempPath(), fs.ZipDir)
		if err := os.RemoveAll(zipPath); err != nil {
			t.Fatal(err)
		}
		return zipPath
	}

	// newArchive creates an archive in zipPath, backdated by age.
	newArchive := func(t *testing.T, zipPath, name string, age time.Duration) string {
		t.Helper()
		if err := os.MkdirAll(zipPath, fs.ModeDir); err != nil {
			t.Fatal(err)
		}
		fileName := filepath.Join(zipPath, name)
		if err := os.WriteFile(fileName, []byte("test"), fs.ModeFile); err != nil {
			t.Fatal(err)
		}
		modTime := time.Now().Add(-age)
		if err := os.Chtimes(fileName, modTime, modTime); err != nil {
			t.Fatal(err)
		}
		return fileName
	}

	t.Cleanup(func() {
		mutex.TempArchives.Store(true)
	})
	t.Run("RemovesExpiredArchive", func(t *testing.T) {
		zipPath := resetZipDir(t)
		expired := newArchive(t, zipPath, "photoprism-download-20260727-094439-zihqtuw4.zip", 2*time.Hour)
		fresh := newArchive(t, zipPath, "photoprism-download-20260727-101500-abcdefgh.zip", time.Minute)
		mutex.TempArchives.Store(true)
		RunPurgeArchives(c)
		assert.False(t, fs.FileExists(expired))
		assert.True(t, fs.FileExists(fresh))
		// A fresh archive is still present, so the next run must scan again.
		assert.True(t, mutex.TempArchives.Load())
	})
	t.Run("ClearsFlagWhenNothingRemains", func(t *testing.T) {
		zipPath := resetZipDir(t)
		expired := newArchive(t, zipPath, "photoprism-download-20260727-094439-zihqtuw4.zip", 2*time.Hour)
		mutex.TempArchives.Store(true)
		RunPurgeArchives(c)
		assert.False(t, fs.FileExists(expired))
		assert.False(t, mutex.TempArchives.Load())
	})
	t.Run("SkipsScanWhenFlagCleared", func(t *testing.T) {
		zipPath := resetZipDir(t)
		expired := newArchive(t, zipPath, "photoprism-download-20260727-094439-zihqtuw4.zip", 2*time.Hour)
		mutex.TempArchives.Store(false)
		RunPurgeArchives(c)
		// Untouched: a cleared flag must prevent any filesystem access.
		assert.True(t, fs.FileExists(expired))
	})
	t.Run("MissingDir", func(t *testing.T) {
		resetZipDir(t)
		mutex.TempArchives.Store(true)
		assert.NotPanics(t, func() { RunPurgeArchives(c) })
		assert.False(t, mutex.TempArchives.Load())
	})
}
