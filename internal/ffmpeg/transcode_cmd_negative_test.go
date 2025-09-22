package ffmpeg

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/ffmpeg/encode"
	"github.com/photoprism/photoprism/pkg/fs"
)

// Negative: destination directory is unwritable; running ffmpeg should fail.
func TestTranscodeCmd_UnwritableDest(t *testing.T) {
	ffmpegBin := "/usr/bin/ffmpeg"
	opt := encode.NewVideoOptions(ffmpegBin, encode.SoftwareAvc, 640, encode.DefaultQuality, encode.PresetFast, "", "0:v:0", "0:a:0?")
	srcName := fs.Abs("./testdata/25fps.vp9")
	dir := t.TempDir()
	unwritable := filepath.Join(dir, "nope")
	if err := os.MkdirAll(unwritable, 0o555); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(unwritable, 0o755)
	destName := filepath.Join(unwritable, "out.mp4")

	cmd, _, err := TranscodeCmd(srcName, destName, opt)
	if err != nil {
		t.Fatal(err)
	}
	err = cmd.Run()
	assert.Error(t, err)
	assert.NoFileExists(t, destName)
}

// Negative: missing ffmpeg binary should cause execution error.
func TestTranscodeCmd_MissingBinary(t *testing.T) {
	opt := encode.NewVideoOptions("/path/does/not/exist/ffmpeg", encode.SoftwareAvc, 640, encode.DefaultQuality, encode.PresetFast, "", "0:v:0", "0:a:0?")
	srcName := fs.Abs("./testdata/25fps.vp9")
	destName := filepath.Join(t.TempDir(), "out.mp4")
	cmd, _, err := TranscodeCmd(srcName, destName, opt)
	if err != nil {
		t.Fatal(err)
	}
	err = cmd.Run()
	assert.Error(t, err)
}
