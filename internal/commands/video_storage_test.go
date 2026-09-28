package commands

import (
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/fs"
)

func TestVideoExistingDir(t *testing.T) {
	t.Run("Existing", func(t *testing.T) {
		dir := t.TempDir()
		assert.Equal(t, dir, videoExistingDir(dir))
	})
	t.Run("Missing", func(t *testing.T) {
		dir := t.TempDir()
		assert.Equal(t, dir, videoExistingDir(filepath.Join(dir, "sidecar", "2026", "09")))
	})
	t.Run("File", func(t *testing.T) {
		dir := t.TempDir()
		fileName := filepath.Join(dir, "clip.mp4")
		require.NoError(t, os.WriteFile(fileName, []byte("clip"), fs.ModeFile))
		assert.Equal(t, dir, videoExistingDir(filepath.Join(fileName, "sub")))
	})
	t.Run("Root", func(t *testing.T) {
		assert.Equal(t, "/", videoExistingDir("/"))
	})
}

func TestVideoCheckFreeSpace(t *testing.T) {
	t.Run("MissingDir", func(t *testing.T) {
		dir := t.TempDir()
		dest := filepath.Join(dir, "sidecar", "2026", "clip.avi.avc")

		// The free space of the nearest existing folder applies, and no folder is created.
		require.NoError(t, videoCheckFreeSpace([]videoOutputPlan{{Destination: dest, SizeBytes: 1}}))
		assert.NoDirExists(t, filepath.Join(dir, "sidecar"))
	})
	t.Run("Insufficient", func(t *testing.T) {
		dir := t.TempDir()
		dest := filepath.Join(dir, "sidecar", "clip.avi.avc")

		err := videoCheckFreeSpace([]videoOutputPlan{{Destination: dest, SizeBytes: math.MaxInt64}})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient free space in "+dir)
	})
	t.Run("NoDestination", func(t *testing.T) {
		assert.NoError(t, videoCheckFreeSpace([]videoOutputPlan{{SizeBytes: math.MaxInt64}}))
	})
}

func TestVideoCreateStageFile(t *testing.T) {
	t.Run("CreateDir", func(t *testing.T) {
		dest := filepath.Join(t.TempDir(), "sidecar", "2026", "clip.mp4")

		staged, err := videoCreateStageFile(dest, true)
		require.NoError(t, err)
		assert.Equal(t, filepath.Dir(dest), filepath.Dir(staged))
		assert.FileExists(t, staged)
		assert.NoFileExists(t, dest)
	})
	t.Run("ExistingDir", func(t *testing.T) {
		dest := filepath.Join(t.TempDir(), "missing", "clip.mp4")

		// Without createDir, the folder must already exist.
		_, err := videoCreateStageFile(dest, false)
		assert.Error(t, err)
		assert.NoDirExists(t, filepath.Dir(dest))
	})
}

func TestVideoCreatesSidecarDir(t *testing.T) {
	conf := config.NewMinimalTestConfig(t.TempDir())

	t.Run("Absolute", func(t *testing.T) {
		require.True(t, conf.SidecarPathIsAbs())
		assert.True(t, videoCreatesSidecarDir(conf, true))
	})
	t.Run("Originals", func(t *testing.T) {
		assert.False(t, videoCreatesSidecarDir(conf, false))
	})
	t.Run("Relative", func(t *testing.T) {
		saved := conf.Options().SidecarPath
		t.Cleanup(func() { conf.Options().SidecarPath = saved })
		conf.Options().SidecarPath = ".photoprism"
		require.False(t, conf.SidecarPathIsAbs())

		// A relative sidecar path would be resolved against the working directory.
		assert.False(t, videoCreatesSidecarDir(conf, true))
	})
	t.Run("NoConfig", func(t *testing.T) {
		assert.False(t, videoCreatesSidecarDir(nil, true))
	})
}
