package photoprism

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/fs"
)

const (
	insta360StackLeft  = "VID_20220625_140410_00_008.insv"
	insta360StackRight = "VID_20220625_140410_10_008.insv"
	insta360StackProxy = "LRV_20220625_140410_11_008.insv"
	insta360StackName  = "VID_20220625_140410_00_008"
)

// newInsta360StackConfig returns an isolated config with its own database and restores the
// previous config when the test ends.
func newInsta360StackConfig(t *testing.T, dbName string, stackSequences bool) *config.Config {
	t.Helper()

	cfg := config.NewMinimalTestConfigWithDb(dbName, filepath.Join(t.TempDir(), "storage"))
	cfg.Settings().Stack.Name = stackSequences

	if !cfg.FFmpegEnabled() {
		t.Skip("FFmpeg must be available to create synthetic capture files")
	}

	oldCfg := Config()
	SetConfig(cfg)
	t.Cleanup(func() {
		SetConfig(oldCfg)
		oldCfg.RegisterDb()
	})

	return cfg
}

// writeInsta360StackMedia writes a one-second square H.264 clip, or a JPEG for image names, with
// bytes that differ per file name.
func writeInsta360StackMedia(t *testing.T, cfg *config.Config, dir, name string) {
	t.Helper()

	require.NoError(t, fs.MkdirAll(dir))

	fileName := filepath.Join(dir, name)
	fileType := fs.FileType(name)
	image := fileType == fs.ImageJpeg || fileType == fs.ImageInsp

	args := []string{"-y", "-loglevel", "error", "-f", "lavfi", "-i", "testsrc=size=320x320:rate=30",
		"-t", "1", "-c:v", "libx264", "-pix_fmt", "yuv420p", "-metadata", "title=" + name, "-f", "mp4", fileName}

	if image {
		args = []string{"-y", "-loglevel", "error", "-f", "lavfi", "-i", "testsrc=size=320x320",
			"-frames:v", "1", "-f", "image2", "-c:v", "mjpeg", fileName}
	}

	// #nosec G204 -- arguments are test constants.
	out, err := exec.Command(cfg.FFmpegBin(), args...).CombinedOutput()
	require.NoError(t, err, strings.TrimSpace(string(out)))

	// Images get the name appended after the end marker, so files with different names never share a hash.
	if image {
		// #nosec G304 -- the destination directory and filename are controlled by the test.
		f, openErr := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, 0)
		require.NoError(t, openErr)
		_, writeErr := f.WriteString(name)
		require.NoError(t, writeErr)
		require.NoError(t, f.Close())
	}
}

// indexInsta360StackFolder indexes a folder below the originals path.
func indexInsta360StackFolder(cfg *config.Config, folder string, rescan, skipArchived bool) {
	indexInsta360StackOptions(cfg, NewIndexOptions(folder, rescan, true, true, false, skipArchived, cfg))
}

// indexInsta360StackOptions runs the indexer with the given options.
func indexInsta360StackOptions(cfg *config.Config, opt IndexOptions) {
	ind := NewIndex(cfg, NewConvert(cfg), NewFiles(), NewPhotos())
	ind.Start(opt)
}

// writeInsta360Photo writes a JPEG of the specified size under an .insp name, with bytes that differ per name.
func writeInsta360Photo(t *testing.T, cfg *config.Config, fileName, size string) {
	t.Helper()
	require.NoError(t, fs.MkdirAll(filepath.Dir(fileName)))

	// #nosec G204 -- arguments are test constants.
	out, err := exec.Command(cfg.FFmpegBin(), "-y", "-loglevel", "error", "-f", "lavfi", "-i", "testsrc=size="+size,
		"-frames:v", "1", "-f", "image2", "-c:v", "mjpeg", "-metadata", "comment="+filepath.Base(fileName), fileName).CombinedOutput()
	require.NoError(t, err, strings.TrimSpace(string(out)))
}
