package photoprism

import (
	"bytes"
	"crypto/rand"
	"image"
	"image/jpeg"
	iofs "io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
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
	// testJpeg writes a JPEG with random pixels, so it never duplicates a file the library holds.
	testJpeg := func(t *testing.T, fileName string) {
		img := image.NewRGBA(image.Rect(0, 0, 24, 16))
		_, err := rand.Read(img.Pix)
		require.NoError(t, err)
		var buf bytes.Buffer
		require.NoError(t, jpeg.Encode(&buf, img, nil))
		require.NoError(t, os.WriteFile(fileName, buf.Bytes(), fs.ModeFile))
	}
	t.Run("MoveFailed", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping test in short mode.")
		} else if os.Geteuid() == 0 {
			t.Skip("requires filesystem permission enforcement")
		}
		originals := cfg.Options().OriginalsPath
		t.Cleanup(func() { cfg.Options().OriginalsPath = originals })
		cfg.Options().OriginalsPath = t.TempDir()
		require.NoError(t, os.Chmod(cfg.OriginalsPath(), 0o500))       //nolint:gosec // Test makes a folder read-only.
		t.Cleanup(func() { _ = os.Chmod(cfg.OriginalsPath(), 0o700) }) //nolint:gosec // Test restores write access.
		dir := stage(t)
		file := filepath.Join(dir, "upload.jpg")
		testJpeg(t, file)
		for _, move := range []bool{true, false} {
			opt := ImportOptionsUpload(dir, "")
			opt.Move = move
			done, err := imp.Run(opt)
			assert.ErrorIs(t, err, ErrImportIncomplete, "move %t", move)
			assert.Equal(t, 1, done.Processed())
			assert.FileExists(t, file)
		}

		// Each run counts its own failures, so a later run that imports the file succeeds.
		require.NoError(t, os.Chmod(cfg.OriginalsPath(), 0o700)) //nolint:gosec // Test restores write access.
		_, err := imp.Run(ImportOptionsUpload(dir, ""))
		assert.NoError(t, err)
		assert.NoFileExists(t, file)
	})
	t.Run("SourceNotRemoved", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping test in short mode.")
		} else if os.Geteuid() == 0 {
			t.Skip("requires filesystem permission enforcement")
		}
		originals := cfg.Options().OriginalsPath
		t.Cleanup(func() { cfg.Options().OriginalsPath = originals })
		cfg.Options().OriginalsPath = t.TempDir()
		dir := stage(t)
		file := filepath.Join(dir, "upload.jpg")
		testJpeg(t, file)
		hash := fs.Hash(file)
		require.NoError(t, os.Chmod(dir, 0o500))       //nolint:gosec // Test makes a folder read-only.
		t.Cleanup(func() { _ = os.Chmod(dir, 0o700) }) //nolint:gosec // Test restores write access.
		_, err := imp.Run(ImportOptionsUpload(dir, ""))
		assert.NoError(t, err)
		assert.FileExists(t, file)
		imported := false
		require.NoError(t, filepath.WalkDir(cfg.OriginalsPath(), func(name string, entry iofs.DirEntry, err error) error {
			if err == nil && !entry.IsDir() && fs.Hash(name) == hash {
				imported = true
			}
			return err
		}))
		assert.True(t, imported, "the content must be in the originals folder")
		indexed, err := entity.FirstFileByHash(hash)
		require.NoError(t, err, "the imported file must be indexed")
		t.Cleanup(func() {
			entity.UnscopedDb().Unscoped().Delete(&entity.File{}, "photo_id = ?", indexed.PhotoID)
			entity.UnscopedDb().Unscoped().Delete(&entity.Details{}, "photo_id = ?", indexed.PhotoID)
			entity.UnscopedDb().Unscoped().Delete(&entity.Photo{}, "id = ?", indexed.PhotoID)
		})
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
