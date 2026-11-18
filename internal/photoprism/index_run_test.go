package photoprism

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// TestIndex_Run pins that a run held off by a face migration reports it, while Start stays quiet
// for the callers that only skip the run, and that any other run reports nothing.
func TestIndex_Run(t *testing.T) {
	cfg := config.TestConfig()
	ind := NewIndex(cfg, NewConvert(cfg), NewFiles(), NewPhotos())

	// A folder that does not exist ends the run before it walks anything.
	opt := NewIndexOptions("index-run-test-missing", false, false, true, false, true, cfg)

	// An empty folder is walked to the end without indexing a fixture, and keeps originals from
	// reading as empty, which would end the run early too.
	dir := filepath.Join(cfg.OriginalsPath(), "index-run-test-"+rnd.Base36(8))
	require.NoError(t, os.MkdirAll(dir, fs.ModeDir))
	t.Cleanup(func() {
		_ = os.Remove(dir)
		_ = entity.UnscopedDb().Where("path = ? AND root = ?", filepath.Base(dir), entity.RootOriginals).Delete(&entity.Folder{}).Error
	})
	emptyOpt := NewIndexOptions(filepath.Base(dir), false, false, true, false, true, cfg)

	t.Run("FacesLocked", func(t *testing.T) {
		lock, err := mutex.AcquireFileLock(cfg.FacesLockFile(), "faces migration")
		require.NoError(t, err)
		t.Cleanup(lock.Release)

		_, _, err = ind.Run(opt)
		assert.ErrorIs(t, err, ErrIndexBusy)
	})
	t.Run("MissingFolder", func(t *testing.T) {
		require.Empty(t, cfg.FacesLocked())

		_, _, err := ind.Run(opt)
		assert.NoError(t, err)
	})
	t.Run("IndexRunning", func(t *testing.T) {
		// Another index holding the worker is not a face migration, and Run does not report it.
		require.NoError(t, mutex.IndexWorker.Start())
		t.Cleanup(mutex.IndexWorker.Stop)

		lastRun, _ := ind.LastRun()

		_, _, err := ind.Run(emptyOpt)
		assert.NoError(t, err)

		after, _ := ind.LastRun()
		assert.Equal(t, lastRun, after, "the run ended at the worker check")
	})
	t.Run("Completed", func(t *testing.T) {
		_, updated, err := ind.Run(emptyOpt)
		assert.NoError(t, err)
		assert.Zero(t, updated)
		assert.False(t, ind.lastRun.IsZero(), "the run reached its end")

		// The folder the walk records is the one the cleanup removes.
		require.NotNil(t, entity.FindFolder(entity.RootOriginals, filepath.Base(dir)))
		require.NoError(t, entity.UnscopedDb().Where("path = ? AND root = ?", filepath.Base(dir), entity.RootOriginals).Delete(&entity.Folder{}).Error)
		assert.Nil(t, entity.FindFolder(entity.RootOriginals, filepath.Base(dir)))
	})
}
