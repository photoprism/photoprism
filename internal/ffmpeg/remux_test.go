package ffmpeg

import (
	"os"
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

		if err := fs.Copy(origName, srcName); err != nil {
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

		if err := fs.Copy(origName, srcName); err != nil {
			t.Fatal(err)
		}

		cmd, err := RemuxCmd(srcName, destName, opt)

		if err != nil {
			t.Fatal(err)
		}

		cmdStr := cmd.String()
		cmdStr = strings.Replace(cmdStr, srcName, "SRC", 1)
		cmdStr = strings.Replace(cmdStr, destName, "DEST", 1)

		assert.Equal(t, "/usr/bin/ffmpeg -hide_banner -y -strict -2 -avoid_negative_ts make_non_negative -i SRC -map 0:v:0 -map 0:a:0? -dn -ignore_unknown -codec copy -f mp4 -movflags use_metadata_tags+faststart -map_metadata 0 DEST", cmdStr)
	})
}
