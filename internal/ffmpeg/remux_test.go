package ffmpeg

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/ffmpeg/encode"
	"github.com/photoprism/photoprism/pkg/fs"
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
	if err := os.WriteFile(dest, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(dest)
	// Should be a no-op and return nil (dest exists, no force)
	err := RemuxFile(src, dest, opt)
	assert.NoError(t, err)
	assert.FileExists(t, dest)
}

func TestRemuxFile_TempExists_NoForce_Error(t *testing.T) {
	opt := encode.NewRemuxOptions("/usr/bin/ffmpeg", fs.VideoMp4, false)
	dir := fs.Abs("./testdata")
	// Use a copy to avoid modifying the original during test
	src := filepath.Join(dir, "30fps.remux-temp.mov")
	orig := filepath.Join(dir, "30fps.mov")
	dest := filepath.Join(dir, "30fps.remux-temp.mp4")
	temp := filepath.Join(dir, ".30fps.remux-temp.mp4")
	// Cleanup
	_ = os.Remove(src)
	_ = os.Remove(dest)
	_ = os.Remove(temp)
	defer func() { _ = os.Remove(src); _ = os.Remove(dest); _ = os.Remove(temp) }()
	// Prepare src and temp conflict
	if err := fs.Copy(orig, src, true); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(temp, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := RemuxFile(src, dest, opt)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "temp file")
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
