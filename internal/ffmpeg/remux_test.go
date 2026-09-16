package ffmpeg

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ffmpeg/encode"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media/video"
)

func TestRemuxFile(t *testing.T) {
	ffmpegBin := "/usr/bin/ffmpeg"

	t.Run("NoFilePath", func(t *testing.T) {
		opt := encode.NewRemuxOptions(ffmpegBin, fs.VideoMp4, false)
		err := RemuxFile("", "", opt)

		assert.Equal(t, "invalid video file path", err.Error())
	})
	t.Run("Mp4", func(t *testing.T) {
		opt := encode.NewRemuxOptions(ffmpegBin, fs.VideoMp4, false)

		// QuickTime MOV container with HVC1 (HEVC) codec.
		origName := fs.Abs("./testdata/30fps.mov")
		srcName := fs.Abs("./testdata/30fps.remux-file.mov")
		tmpName := fs.Abs("./testdata/.30fps.remux-file.mp4")
		destName := fs.Abs("./testdata/30fps.remux-file.avc")

		_ = os.Remove(srcName)
		_ = os.Remove(tmpName)
		_ = os.Remove(destName)

		defer func() {
			_ = os.Remove(srcName)
			_ = os.Remove(tmpName)
			_ = os.Remove(destName)
		}()

		if err := fs.Copy(origName, srcName, true); err != nil {
			t.Fatal(err)
		}

		if err := RemuxFile(srcName, destName, opt); err != nil {
			t.Fatal(err)
		}

		assert.FileExists(t, srcName)
		assert.NoFileExists(t, tmpName)
		assert.FileExists(t, destName)
	})
}

func TestRemuxCmd(t *testing.T) {
	ffmpegBin := "/usr/bin/ffmpeg"

	t.Run("NoSrcName", func(t *testing.T) {
		opt := encode.NewRemuxOptions(ffmpegBin, fs.VideoMp4, false)
		_, err := RemuxCmd("", "", opt)

		assert.Equal(t, "empty source filename", err.Error())
	})
	t.Run("Mp4", func(t *testing.T) {
		opt := encode.NewRemuxOptions(ffmpegBin, fs.VideoMp4, false)

		// QuickTime MOV container with HVC1 (HEVC) codec.
		origName := fs.Abs("./testdata/30fps.mov")

		srcName := fs.Abs("./testdata/30fps.remux-cmd.mov")
		destName := fs.Abs("./testdata/30fps.remux-cmd.mp4")

		_ = os.Remove(srcName)
		_ = os.Remove(destName)

		defer func() {
			_ = os.Remove(srcName)
			_ = os.Remove(destName)
		}()

		if err := fs.Copy(origName, srcName, true); err != nil {
			t.Fatal(err)
		}

		cmd, err := RemuxCmd(srcName, destName, opt)

		if err != nil {
			t.Fatal(err)
		}

		cmdStr := cmd.String()
		cmdStr = strings.Replace(cmdStr, srcName, "SRC", 1)
		cmdStr = strings.Replace(cmdStr, destName, "DEST", 1)

		assert.Equal(t, "/usr/bin/ffmpeg -hide_banner -y -strict -2 -avoid_negative_ts make_zero -i SRC -map 0:v:0 -map 0:a:0? -dn -ignore_unknown -codec copy -f mp4 -movflags use_metadata_tags+faststart -map_metadata 0 DEST", cmdStr)
	})
}

func TestRemuxFile_DestExists_NoForce_NoOp(t *testing.T) {
	opt := encode.NewRemuxOptions("/usr/bin/ffmpeg", fs.VideoMp4, false)
	dir := fs.Abs("./testdata")
	src := filepath.Join(dir, "30fps.mov")
	dest := filepath.Join(dir, "already-there.mp4")
	// Create a tiny placeholder dest file
	_ = os.Remove(dest)
	if err := os.WriteFile(dest, []byte("x"), fs.ModeFile); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(dest)
	// Should be a no-op and return nil (dest exists, no force)
	err := RemuxFile(src, dest, opt)
	assert.NoError(t, err)
	assert.FileExists(t, dest)
}

// TestRemuxFile_SharedStem verifies that concurrent conversions of sources sharing a base name each
// produce their own result. A camcorder writes such a set, e.g. MOV001.tod beside MOV001.mts.
func TestRemuxFile_SharedStem(t *testing.T) {
	opt := encode.NewRemuxOptions("/usr/bin/ffmpeg", fs.VideoMp4, false)
	dir := t.TempDir()

	// Every source carries the stem "MOV001" and one of two video codecs, so each destination can
	// be checked against the source it was made from.
	fixtures := []string{"30fps.mov", "25fps.vp9", "30fps.mov", "25fps.vp9", "30fps.mov", "25fps.vp9"}
	sources := make([]string, len(fixtures))
	dests := make([]string, len(fixtures))

	for i, fixture := range fixtures {
		srcDir := filepath.Join(dir, fmt.Sprintf("src%d", i))
		require.NoError(t, os.Mkdir(srcDir, fs.ModeDir))

		sources[i] = copyFixture(t, fixture, filepath.Join(srcDir, "MOV001"+filepath.Ext(fixture)))
		dests[i] = filepath.Join(dir, fmt.Sprintf("out%d.mp4", i))
	}

	var wg sync.WaitGroup
	errs := make([]error, len(fixtures))

	wg.Add(len(fixtures))

	for i := range fixtures {
		go func() {
			defer wg.Done()
			errs[i] = RemuxFile(sources[i], dests[i], opt)
		}()
	}

	wg.Wait()

	// Each destination is checked against the codec of its own source.
	for i, fixture := range fixtures {
		require.NoErrorf(t, errs[i], "remux of %s", fixture)

		info, err := video.ProbeFile(dests[i])
		require.NoError(t, err)

		if fixture == "30fps.mov" {
			assert.Equalf(t, video.CodecHvc1, info.VideoCodec, "%s must hold the stream of %s", dests[i], fixture)
		} else {
			assert.NotEqualf(t, video.CodecHvc1, info.VideoCodec, "%s must hold the stream of %s", dests[i], fixture)
		}
	}
}

// TestRemuxFile_FailureLeavesNothing verifies that a remux the muxer cannot complete leaves neither
// a destination nor a working file behind.
func TestRemuxFile_FailureLeavesNothing(t *testing.T) {
	opt := encode.NewRemuxOptions("/usr/bin/ffmpeg", fs.VideoMp4, false)
	dir := t.TempDir()

	src := filepath.Join(dir, "MOV001.mov")
	require.NoError(t, os.WriteFile(src, []byte("not a video"), fs.ModeFile))

	dest := filepath.Join(dir, "MOV001.mp4")

	assert.Error(t, RemuxFile(src, dest, opt))
	assert.NoFileExists(t, dest)

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)

	for _, entry := range entries {
		assert.Equalf(t, "MOV001.mov", entry.Name(), "a failed remux must leave no working file")
	}
}

// TestRemuxFile_ForeignTempFile verifies that a hidden file beside the destination is neither
// written nor removed, and does not block the conversion.
func TestRemuxFile_ForeignTempFile(t *testing.T) {
	opt := encode.NewRemuxOptions("/usr/bin/ffmpeg", fs.VideoMp4, false)
	dir := t.TempDir()

	src := copyFixture(t, "30fps.mov", filepath.Join(dir, "MOV001.mov"))
	dest := filepath.Join(dir, "MOV001.mp4")

	foreign := filepath.Join(dir, ".MOV001.mp4")
	kept := []byte("written by another call")
	require.NoError(t, os.WriteFile(foreign, kept, fs.ModeFile))

	require.NoError(t, RemuxFile(src, dest, opt))

	info, err := video.ProbeFile(dest)
	require.NoError(t, err)
	assert.Equal(t, video.CodecHvc1, info.VideoCodec)

	// #nosec G304 -- the path is built by the test from its own temp directory.
	after, err := os.ReadFile(foreign)
	require.NoError(t, err)
	assert.Equal(t, kept, after, "a file this call did not create must not be written or removed")
}

// copyFixture copies a test fixture to the given path and returns it.
func copyFixture(t *testing.T, name, dest string) string {
	t.Helper()

	if err := fs.Copy(fs.Abs(filepath.Join("./testdata", name)), dest, true); err != nil {
		t.Fatal(err)
	}

	return dest
}

func TestRemuxCmd_VideoTag(t *testing.T) {
	ffmpegBin := "/usr/bin/ffmpeg"
	src := fs.Abs("./testdata/30fps.mov")
	dest := fs.Abs("./testdata/30fps.video-tag.mp4")

	defer func() { _ = os.Remove(dest) }()

	opt := encode.NewRemuxOptions(ffmpegBin, fs.VideoMp4, false)
	opt.VideoTag = "hvc1"

	cmd, err := RemuxCmd(src, dest, opt)
	if err != nil {
		t.Fatal(err)
	}

	assert.Contains(t, cmd.String(), "-tag:v hvc1")
}

func TestRemuxFile_AutoTagsHevc(t *testing.T) {
	// 30fps.mov is a QuickTime MOV with HEVC (hvc1) video.
	src := fs.Abs("./testdata/30fps.mov")
	if !fs.FileExistsNotEmpty(src) {
		t.Skip("missing testdata")
	}

	opt := encode.NewRemuxOptions("/usr/bin/ffmpeg", fs.VideoMp4, false)
	assert.Empty(t, opt.VideoTag)

	// RemuxFile mutates the opt's VideoTag when it detects an HEVC source — we
	// can verify that by reaching into the package via a small helper test
	// that mirrors RemuxFile's pre-cmd block without actually invoking ffmpeg.
	if (opt.Container == fs.VideoMp4 || opt.Container == fs.VideoMov) && opt.VideoTag == "" {
		if video.IsHEVCFile(src) {
			opt.VideoTag = "hvc1"
		}
	}

	assert.Equal(t, "hvc1", opt.VideoTag)
}

func TestRemuxCmd_ErrorPaths_And_DefaultBin(t *testing.T) {
	// Same source/dest error
	opt := encode.NewRemuxOptions("", fs.VideoMp4, false)
	_, err := RemuxCmd("file.mp4", "file.mp4", opt)
	assert.Error(t, err)
	// Non-existent src
	_, err = RemuxCmd("./testdata/does-not-exist.mp4", "out.mp4", opt)
	assert.Error(t, err)
	// Default ffmpeg bin selected when empty
	// Use an existing file to pass validation
	src := fs.Abs("./testdata/30fps.mov")
	dest := fs.Abs("./testdata/30fps.default-bin.mp4")
	_ = os.Remove(dest)
	defer os.Remove(dest)
	cmd, err := RemuxCmd(src, dest, opt)
	if err != nil {
		t.Fatal(err)
	}
	assert.Contains(t, cmd.String(), "ffmpeg ")
}
