package photoprism

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/fs"
)

// TestExportValidationCache verifies shared reads, failure retries, expiry, and bounded retention.
func TestExportValidationCache(t *testing.T) {
	cache := &exportValidationCache{}
	var reads atomic.Int32
	var wg sync.WaitGroup
	start := make(chan struct{})
	release := make(chan struct{})
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			assert.NoError(t, cache.verify("same", func() error { reads.Add(1); <-release; return nil }))
		}()
	}
	close(start)
	require.Eventually(t, func() bool { return reads.Load() > 0 }, time.Second, time.Millisecond)
	time.Sleep(10 * time.Millisecond)
	close(release)
	wg.Wait()
	assert.EqualValues(t, 1, reads.Load())
	t.Run("FailuresAreRetried", func(t *testing.T) {
		assert.ErrorIs(t, cache.verify("failure", func() error { return ErrExportSourceChanged }), ErrExportSourceChanged)
		assert.NoError(t, cache.verify("failure", func() error { return nil }))
	})
	t.Run("ExpiredSuccessIsRevalidated", func(t *testing.T) {
		cache.entries["same"].expires = time.Now().Add(-time.Second)
		assert.NoError(t, cache.verify("same", func() error { reads.Add(1); return nil }))
		assert.EqualValues(t, 2, reads.Load())
	})
	t.Run("BoundedRetention", func(t *testing.T) {
		for i := range 256 {
			require.NoError(t, cache.verify(fmt.Sprint(i), func() error { return nil }))
		}
		assert.LessOrEqual(t, len(cache.entries), 128)
	})
	t.Run("HeavyReadsDoNotOccupyProbeSlots", func(t *testing.T) {
		for range cap(exportValidationSlots) {
			exportValidationSlots <- struct{}{}
		}
		defer func() {
			for range cap(exportValidationSlots) {
				<-exportValidationSlots
			}
		}()
		assert.ErrorIs(t, cache.verify("cold", func() error { t.Fatal("must not run"); return nil }), ErrExportBusy)
		assert.Empty(t, exportProbeSlots)
		w := exportTestConfig(t)
		f := exportTestVideo(t, w, "independent.mp4", "libx264")
		_, err := w.ProbeExport(f)
		require.NoError(t, err)
	})
	t.Run("PanicReleasesWaitersAndCapacity", func(t *testing.T) {
		assert.Panics(t, func() { _ = cache.verify("panic", func() error { panic("test") }) })
		assert.Empty(t, exportValidationSlots)
		assert.NoError(t, cache.verify("panic", func() error { return nil }))
	})
}

// TestExportValidationVersion rejects a mutation with a restored mtime after caching.
func TestExportValidationVersion(t *testing.T) {
	w := exportTestConfig(t)
	f := exportTestVideo(t, w, "version.mp4", "libx265")
	_, err := w.PrepareExport(f, "avc")
	require.NoError(t, err)
	name := filepath.Join(w.conf.OriginalsPath(), f.FileName)
	before, err := os.Stat(name)
	require.NoError(t, err)
	data, err := os.ReadFile(name) //nolint:gosec // name is confined to the disposable originals directory
	require.NoError(t, err)
	data[len(data)-1] ^= 1
	require.NoError(t, os.WriteFile(name, data, fs.ModeFile)) //nolint:gosec // name is confined to the disposable originals directory
	require.NoError(t, os.Chtimes(name, before.ModTime(), before.ModTime()))
	_, _, _, err = w.ExportFile(f, "avc")
	assert.ErrorIs(t, err, ErrExportSourceChanged)
}

// TestExportWarmDownloads reuses verified sources while cold hash capacity is fully occupied.
func TestExportWarmDownloads(t *testing.T) {
	w := exportTestConfig(t)
	f := exportTestVideo(t, w, "warm.mp4", "libx265")
	stat, err := os.Stat(filepath.Join(w.conf.OriginalsPath(), f.FileName))
	require.NoError(t, err)
	if exportFileVersion(stat) == "" {
		t.Skip("reliable file versions unavailable")
	}
	_, err = w.PrepareExport(f, "avc")
	require.NoError(t, err)
	for range cap(exportValidationSlots) {
		exportValidationSlots <- struct{}{}
	}
	defer func() {
		for range cap(exportValidationSlots) {
			<-exportValidationSlots
		}
	}()
	var wg sync.WaitGroup
	for range 4 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			file, _, output, openErr := w.ExportFile(f, "avc")
			assert.NoError(t, openErr)
			if file != nil {
				assert.NoError(t, file.Close())
			}
			assert.True(t, output.Ready)
		}()
	}
	wg.Wait()
	_, err = w.ProbeExport(f)
	require.NoError(t, err)
}
