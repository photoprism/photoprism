package photoprism

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/fs/disk"
	"github.com/photoprism/photoprism/pkg/log/status"
)

// TestExportFailurePaths checks capacity, source replacement, and failed staging.
func TestExportFailurePaths(t *testing.T) {
	w := exportTestConfig(t)
	t.Run("ProbeCapacity", func(t *testing.T) {
		f := exportTestVideo(t, w, "capacity.mp4", "libx265")
		for i := 0; i < cap(exportProbeSlots); i++ {
			exportProbeSlots <- struct{}{}
		}
		defer func() {
			for i := 0; i < cap(exportProbeSlots); i++ {
				<-exportProbeSlots
			}
		}()
		_, err := w.ProbeExport(f)
		assert.ErrorIs(t, err, ErrExportBusy)
		_, _, _, err = w.ExportFile(f, "avc")
		assert.ErrorIs(t, err, ErrExportBusy)
	})
	t.Run("PreparationCapacity", func(t *testing.T) {
		f := exportTestVideo(t, w, "prepare-capacity.mp4", "libx265")
		for i := 0; i < cap(exportSlots); i++ {
			exportSlots <- struct{}{}
		}
		defer func() {
			for i := 0; i < cap(exportSlots); i++ {
				<-exportSlots
			}
		}()
		_, err := w.PrepareExport(f, "avc")
		assert.ErrorIs(t, err, ErrExportBusy)
	})
	t.Run("ChangedWhileWaiting", func(t *testing.T) {
		f := exportTestVideo(t, w, "changed.mp4", "libx265")
		blocked, key, unblock := blockAvcDestination(t, w, f)
		result := make(chan error, 1)
		go func() { _, err := w.PrepareExport(f, "avc"); result <- err }()
		require.Eventually(t, func() bool {
			avcConversions.mutex.Lock()
			defer avcConversions.mutex.Unlock()
			return avcConversions.calls[key] != blocked
		}, time.Second, time.Millisecond)
		replacement := exportTestVideo(t, w, "replacement.mp4", "libx264")
		require.NoError(t, os.Rename(filepath.Join(w.conf.OriginalsPath(), replacement.FileName), filepath.Join(w.conf.OriginalsPath(), f.FileName)))
		unblock()
		assert.ErrorIs(t, <-result, ErrExportSourceChanged)
		_, _, _, err := w.ExportFile(f, "avc")
		assert.ErrorIs(t, err, ErrExportSourceChanged)
	})
	t.Run("LowStorage", func(t *testing.T) {
		f := exportTestVideo(t, w, "low-storage.mp4", "libx265")
		disk.FlushFree()
		defer disk.FlushFree()
		old := config.DisableStorageCheck.Load()
		defer config.DisableStorageCheck.Store(old)
		config.DisableStorageCheck.Store(false)
		disk.SetFree(w.conf.StoragePath(), disk.MB, 1000*disk.MB)
		_, err := w.PrepareExport(f, "avc")
		assert.ErrorIs(t, err, status.ErrInsufficientStorage)
	})
	t.Run("UnwritableSidecar", func(t *testing.T) {
		f := exportTestVideo(t, w, "no-cache.mp4", "libx265")
		old := w.conf.Options().SidecarPath
		w.conf.Options().SidecarPath = filepath.Join(t.TempDir(), "file")
		defer func() { w.conf.Options().SidecarPath = old }()
		require.NoError(t, os.WriteFile(w.conf.SidecarPath(), []byte("not a directory"), fs.ModeFile))
		_, err := w.PrepareExport(f, "avc")
		assert.Error(t, err)
	})
	t.Run("NoSidecarBorrowing", func(t *testing.T) {
		f := exportTestVideo(t, w, "borrow.mp4", "libx265")
		foreign := exportTestVideo(t, w, "borrow.avc.mp4", "libx264")
		foreignName := filepath.Join(w.conf.OriginalsPath(), foreign.FileName)
		before, err := os.ReadFile(foreignName) //nolint:gosec // name is confined to the disposable originals directory
		require.NoError(t, err)
		_, err = w.PrepareExport(f, "avc")
		require.NoError(t, err)
		name, _ := exportTestSidecar(t, w, f)
		assert.NotEqual(t, foreignName, name)
		after, err := os.ReadFile(foreignName) //nolint:gosec // name is confined to the disposable originals directory
		require.NoError(t, err)
		assert.Equal(t, before, after)
	})
	t.Run("RejectsUnsupportedInput", func(t *testing.T) {
		name := filepath.Join(w.conf.OriginalsPath(), "unsupported.mp4")
		require.NoError(t, os.WriteFile(name, []byte("unsupported media"), fs.ModeFile))
		_, err := w.exportVideoInfo(name)
		assert.ErrorIs(t, err, ErrExportUnavailable)
	})
}
