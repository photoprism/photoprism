//go:build unix

package photoprism

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/fs"
)

// TestExportSpecialFiles ensures named pipes cannot block sidecar or metadata reads.
func TestExportSpecialFiles(t *testing.T) {
	w := exportTestConfig(t)
	f := exportTestVideo(t, w, "fifo.mp4", "libx265")
	source, err := w.exportSource(f)
	require.NoError(t, err)
	name, err := w.avcName(source.media)
	require.NoError(t, err)
	require.NoError(t, fs.MkdirAll(filepath.Dir(name)))
	t.Run("OutputFIFO", func(t *testing.T) {
		require.NoError(t, syscall.Mkfifo(name, uint32(fs.ModeFile)))
		t.Cleanup(func() { _ = os.Remove(name) })
		_, _, _, err := w.ExportFile(f, "avc")
		assert.ErrorIs(t, err, ErrExportNotReady)
		_, err = w.PrepareExport(f, "avc")
		assert.ErrorIs(t, err, ErrExportUnavailable)
	})
	t.Run("OutputSymlink", func(t *testing.T) {
		other := exportTestVideo(t, w, "foreign.mp4", "libx264")
		require.NoError(t, os.Symlink(filepath.Join(w.conf.OriginalsPath(), other.FileName), name))
		t.Cleanup(func() { _ = os.Remove(name) })
		_, _, _, err := w.ExportFile(f, "avc")
		assert.ErrorIs(t, err, ErrExportNotReady)
		_, err = w.PrepareExport(f, "avc")
		assert.ErrorIs(t, err, ErrExportUnavailable)
	})
	t.Run("MetadataFIFO", func(t *testing.T) {
		dir, err := fs.CachePath(w.conf.CachePath(), f.FileHash, "json", true)
		require.NoError(t, err)
		require.NoError(t, syscall.Mkfifo(filepath.Join(dir, f.FileHash+"_exiftool.json"), uint32(fs.ModeFile)))
		assert.Empty(t, w.exportAppleKey(f))
	})
}
