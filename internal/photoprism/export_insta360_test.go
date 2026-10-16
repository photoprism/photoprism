package photoprism

import (
	"fmt"
	iofs "io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/ffmpeg/encode"
	"github.com/photoprism/photoprism/pkg/fs"
)

// exportInsta360Pair creates both lenses without a stitched sidecar.
func exportInsta360Pair(t *testing.T, w *Convert) (*entity.File, *entity.File) {
	t.Helper()
	files := make([]*entity.File, 0, 2)
	for _, name := range []string{insta360StackLeft, insta360StackRight} {
		writeInsta360StackMedia(t, w.conf, w.conf.OriginalsPath(), name)
		files = append(files, &entity.File{FileRoot: entity.RootOriginals, FileName: name, FileHash: fs.Hash(filepath.Join(w.conf.OriginalsPath(), name)), FileType: "insv", FileVideo: true, MediaType: entity.MediaVideo, FileWidth: 320, FileHeight: 320})
	}
	return files[0], files[1]
}

// TestExportInsta360Capture preserves stitched playback when Export prepares a missing sidecar.
func TestExportInsta360Capture(t *testing.T) {
	w := exportTestConfig(t)
	left, right := exportInsta360Pair(t, w)
	t.Run("ProbeOwnership", func(t *testing.T) {
		probe, err := w.ProbeExport(left)
		require.NoError(t, err)
		assert.Equal(t, "main", probe.CaptureRole)
		require.NotNil(t, probe.Avc)
		assert.True(t, probe.Avc.Preparable)
		probe, err = w.ProbeExport(right)
		require.NoError(t, err)
		assert.Equal(t, "secondary", probe.CaptureRole)
		assert.Nil(t, probe.Avc)
		_, err = w.PrepareExport(right, "avc")
		assert.ErrorIs(t, err, ErrExportUnavailable)
		_, _, _, err = w.ExportFile(right, "avc")
		assert.ErrorIs(t, err, ErrExportUnavailable)
	})
	t.Run("MissingSidecarIsStitched", func(t *testing.T) {
		output, err := w.PrepareExport(left, "avc")
		require.NoError(t, err)
		assert.Equal(t, 2*output.Height, output.Width)
		assert.Greater(t, output.Width, 320)
		name, downloaded := exportTestSidecar(t, w, left)
		assert.Equal(t, output, downloaded)
		before, err := os.Stat(name)
		require.NoError(t, err)
		media, err := NewMediaFile(filepath.Join(w.conf.OriginalsPath(), left.FileName))
		require.NoError(t, err)
		playback, err := w.ToAvc(media, encode.SoftwareAvc, false, false)
		require.NoError(t, err)
		assert.Equal(t, name, playback.FileName())
		_, err = w.PrepareExport(left, "avc")
		require.NoError(t, err)
		after, err := os.Stat(name)
		require.NoError(t, err)
		assert.True(t, os.SameFile(before, after))
		assert.Equal(t, before.ModTime(), after.ModTime())
	})
}

// TestExportInsta360LensChanged prevents publication if the second lens changes during conversion.
func TestExportInsta360LensChanged(t *testing.T) {
	w := exportTestConfig(t)
	left, right := exportInsta360Pair(t, w)
	started, _, unblock := blockingAvcCommand(t, w, false)
	result := make(chan error, 1)
	go func() { _, err := w.PrepareExport(left, "avc"); result <- err }()
	require.Eventually(t, func() bool { return fs.FileExists(started) }, 10*time.Second, time.Millisecond)
	name := filepath.Join(w.conf.OriginalsPath(), right.FileName)
	stat, err := os.Stat(name)
	require.NoError(t, err)
	require.NoError(t, os.Chtimes(name, stat.ModTime().Add(time.Second), stat.ModTime().Add(time.Second)))
	unblock()
	assert.ErrorIs(t, <-result, ErrExportSourceChanged)
	entries, err := os.ReadDir(w.conf.SidecarPath())
	require.NoError(t, err)
	assert.Empty(t, entries)
}

// TestExportInsta360Admission requires the companion to belong to the same indexed photo.
func TestExportInsta360Admission(t *testing.T) {
	for _, state := range []string{"Unindexed", "OtherPhoto", "SamePhoto", "StaleHash"} {
		t.Run(state, func(t *testing.T) {
			cfg := newInsta360StackConfig(t, "export-insta360-admission-"+state, false)
			require.NoError(t, cfg.CreateDirectories())
			w := NewConvert(cfg)
			left, right := exportInsta360Pair(t, w)
			photo := &entity.Photo{PhotoTitle: "Export Capture"}
			require.NoError(t, photo.Create())
			left.PhotoID, left.PhotoUID = photo.ID, photo.PhotoUID
			require.NoError(t, left.Create())
			if state != "Unindexed" {
				owner := photo
				if state == "OtherPhoto" {
					owner = &entity.Photo{PhotoTitle: "Other Capture"}
					require.NoError(t, owner.Create())
				}
				right.PhotoID, right.PhotoUID = owner.ID, owner.PhotoUID
				if state == "StaleHash" {
					right.FileHash = "1234567890123456789012345678901234567890"
				}
				require.NoError(t, right.Create())
			}
			before := exportCacheTree(t, cfg.CachePath())
			for range 2 {
				probe, err := w.ProbeExport(left)
				if state == "SamePhoto" || state == "StaleHash" {
					require.NoError(t, err)
					assert.Equal(t, "main", probe.CaptureRole)
					require.NotNil(t, probe.Avc)
				} else {
					assert.ErrorIs(t, err, ErrExportUnavailable)
				}
			}
			assert.Equal(t, before, exportCacheTree(t, cfg.CachePath()))
			if state == "StaleHash" {
				_, err := w.PrepareExport(left, "avc")
				assert.ErrorIs(t, err, ErrExportSourceChanged)
				entries, err := os.ReadDir(cfg.SidecarPath())
				require.NoError(t, err)
				assert.Empty(t, entries)
			}
		})
	}
}

// exportCacheTree snapshots nested cache paths and versions to detect probe filesystem writes.
func exportCacheTree(t *testing.T, root string) []string {
	t.Helper()
	result := make([]string, 0)
	require.NoError(t, filepath.WalkDir(root, func(name string, entry iofs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		result = append(result, fmt.Sprintf("%s:%s:%d:%d", name, info.Mode(), info.Size(), info.ModTime().UnixNano()))
		return nil
	}))
	return result
}

// TestExportInsta360FrozenCapture rejects a mate that appears after source admission.
func TestExportInsta360FrozenCapture(t *testing.T) {
	w := exportTestConfig(t)
	left, right := exportInsta360Pair(t, w)
	rightName := filepath.Join(w.conf.OriginalsPath(), right.FileName)
	saved := filepath.Join(t.TempDir(), "right.insv")
	require.NoError(t, os.Rename(rightName, saved))
	source, err := w.exportSource(left)
	require.NoError(t, err)
	require.NoError(t, os.Rename(saved, rightName))
	assert.ErrorIs(t, source.version.check(source.media.FileName()), ErrExportSourceChanged)
	capture := FindInsta360Capture(source.media)
	require.NotNil(t, capture)
	assert.Nil(t, capture.Right)
}

// TestExportInsta360PlaybackLensChanged rejects a changed companion in playback-first conversion.
func TestExportInsta360PlaybackLensChanged(t *testing.T) {
	w := exportTestConfig(t)
	left, right := exportInsta360Pair(t, w)
	media, err := NewMediaFile(filepath.Join(w.conf.OriginalsPath(), left.FileName))
	require.NoError(t, err)
	started, _, unblock := blockingAvcCommand(t, w, false)
	result := make(chan error, 1)
	go func() { _, callErr := w.ToAvc(media, encode.SoftwareAvc, false, false); result <- callErr }()
	require.Eventually(t, func() bool { return fs.FileExists(started) }, 10*time.Second, time.Millisecond)
	name := filepath.Join(w.conf.OriginalsPath(), right.FileName)
	stat, err := os.Stat(name)
	require.NoError(t, err)
	require.NoError(t, os.Chtimes(name, stat.ModTime().Add(time.Second), stat.ModTime().Add(time.Second)))
	unblock()
	assert.ErrorIs(t, <-result, ErrExportSourceChanged)
	entries, err := os.ReadDir(w.conf.SidecarPath())
	require.NoError(t, err)
	assert.Empty(t, entries)
}

// TestExportInsta360QueuedMateAppearance keeps queued playback and exact conversions within their snapshot.
func TestExportInsta360QueuedMateAppearance(t *testing.T) {
	for _, exact := range []bool{false, true} {
		name := "Playback"
		if exact {
			name = "ExactExport"
		}
		t.Run(name, func(t *testing.T) {
			w := exportTestConfig(t)
			left, right := exportInsta360Pair(t, w)
			rightName := filepath.Join(w.conf.OriginalsPath(), right.FileName)
			saved := filepath.Join(t.TempDir(), "right.insv")
			require.NoError(t, os.Rename(rightName, saved))
			source, err := w.exportSource(left)
			require.NoError(t, err)
			blocked, key, unblock := blockAvcDestination(t, w, left)
			result := make(chan error, 1)
			go func() {
				var callErr error
				if exact {
					_, callErr = w.coordinatedAvc(source.media, encode.SoftwareAvc, false, false, true, source.version)
				} else {
					_, callErr = w.ToAvc(source.media, encode.SoftwareAvc, false, false)
				}
				result <- callErr
			}()
			require.Eventually(t, func() bool {
				avcConversions.mutex.Lock()
				defer avcConversions.mutex.Unlock()
				return avcConversions.calls[key] != blocked
			}, time.Second, time.Millisecond)
			require.NoError(t, os.Rename(saved, rightName))
			unblock()
			assert.ErrorIs(t, <-result, ErrExportSourceChanged)
			entries, err := os.ReadDir(w.conf.SidecarPath())
			require.NoError(t, err)
			assert.Empty(t, entries)
		})
	}
}
