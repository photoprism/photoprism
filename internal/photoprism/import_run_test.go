package photoprism

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/fs/disk"
	"github.com/photoprism/photoprism/pkg/log/status"
)

// importRunHook runs a function once when the import logs a message with the given prefix.
type importRunHook struct {
	prefix string
	fire   func()
	fired  bool
}

// Levels selects the log levels the hook observes.
func (h *importRunHook) Levels() []logrus.Level { return logrus.AllLevels }

// Fire runs the function on the first matching message.
func (h *importRunHook) Fire(entry *logrus.Entry) error {
	if !h.fired && strings.HasPrefix(entry.Message, h.prefix) {
		h.fired = true
		h.fire()
	}
	return nil
}

// TestImport_Run verifies that Run reports an import that did not run or refused files.
func TestImport_Run(t *testing.T) {
	cfg := config.NewMinimalTestConfig(t.TempDir())
	convert := NewConvert(cfg)
	imp := NewImport(cfg, NewIndex(cfg, convert, NewFiles(), NewPhotos()), convert)
	storageCheck := config.DisableStorageCheck.Load()
	t.Cleanup(func() { config.DisableStorageCheck.Store(storageCheck) })

	// stage returns a folder with files the import cannot move.
	stage := func(t *testing.T, names ...string) string {
		dir := t.TempDir()
		for _, name := range append([]string{"notes.txt"}, names...) {
			require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(name), fs.ModeFile))
		}
		return dir
	}

	// The storage check is disabled unless a subtest tests it, so the host's free space does not matter.
	config.DisableStorageCheck.Store(true)

	t.Run("FacesLocked", func(t *testing.T) {
		lock, err := mutex.AcquireFileLock(cfg.FacesLockFile(), "faces migration")
		require.NoError(t, err)
		t.Cleanup(lock.Release)
		dir := stage(t)
		done, err := imp.Run(ImportOptionsUpload(dir, ""))
		assert.ErrorIs(t, err, ErrImportBusy)
		assert.Empty(t, done)
		assert.Empty(t, imp.Start(ImportOptionsUpload(dir, "")))
		lock.Release()
	})
	t.Run("IndexRunning", func(t *testing.T) {
		require.NoError(t, mutex.IndexWorker.Start())
		t.Cleanup(mutex.IndexWorker.Stop)
		dir := stage(t)
		opt := ImportOptionsUpload(dir, "")
		opt.NonBlocking = false
		done, err := imp.Run(opt)
		assert.ErrorIs(t, err, ErrImportBusy)
		assert.Empty(t, done)
		opt.NonBlocking = true
		done, err = imp.Run(opt)
		assert.NoError(t, err)
		assert.NotEmpty(t, done)
	})
	t.Run("Canceled", func(t *testing.T) {
		require.NoError(t, mutex.IndexWorker.Start())
		t.Cleanup(mutex.IndexWorker.Stop)
		mutex.IndexWorker.Cancel()
		done, err := imp.Run(ImportOptionsUpload(stage(t), ""))
		assert.ErrorIs(t, err, status.ErrCanceled)
		assert.Empty(t, done)
	})
	t.Run("StorageStopWhileIndexing", func(t *testing.T) {
		lowPct, lowBytes := disk.StorageLowPct, disk.StorageLowBytes
		t.Cleanup(func() {
			disk.StorageLowPct, disk.StorageLowBytes = lowPct, lowBytes
			config.DisableStorageCheck.Store(true)
			disk.FlushFree()
		})
		disk.StorageLowPct, disk.StorageLowBytes = disk.DefaultStorageLowPct, disk.DefaultStorageLowBytes
		config.DisableStorageCheck.Store(false)
		disk.FlushFree()
		if cfg.InsufficientStorage() || config.DisableStorageCheck.Load() {
			t.Skip("storage check is unavailable or reports low storage on this host")
		}
		logger, ok := log.(*logrus.Logger)
		require.True(t, ok)
		hooks := logger.ReplaceHooks(make(logrus.LevelHooks))
		t.Cleanup(func() { logger.ReplaceHooks(hooks) })
		hook := &importRunHook{prefix: "import: ignored", fire: func() { disk.SetFree(cfg.StoragePath(), disk.MB, 1000*disk.MB) }}
		logger.AddHook(hook)
		require.NoError(t, mutex.IndexWorker.Start())
		t.Cleanup(mutex.IndexWorker.Stop)
		dir := stage(t, "a.txt", "b.txt", "c.txt", "d.txt")
		require.NoError(t, os.WriteFile(filepath.Join(dir, fs.PPIgnoreFilename), []byte("b.txt\n"), fs.ModeFile))
		_, err := imp.Run(ImportOptionsUpload(dir, ""))
		require.True(t, hook.fired)
		assert.ErrorIs(t, err, status.ErrInsufficientStorage)
	})
	t.Run("NotFound", func(t *testing.T) {
		_, err := imp.Run(ImportOptionsUpload(filepath.Join(t.TempDir(), "missing"), ""))
		assert.Error(t, err)
	})
	t.Run("NoMediaFiles", func(t *testing.T) {
		done, err := imp.Run(ImportOptionsUpload(stage(t), ""))
		assert.NoError(t, err)
		assert.Zero(t, done.Processed())
	})
}
