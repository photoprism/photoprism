package ffmpeg

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/ffmpeg/encode"
	"github.com/photoprism/photoprism/pkg/fs"
)

// Negative: destination directory is unwritable, ffmpeg should fail to write.
func TestExtractImageCmd_UnwritableDest(t *testing.T) {
	opt := encode.NewPreviewImageOptions("/usr/bin/ffmpeg", time.Second*1)
	srcName := fs.Abs("./testdata/25fps.vp9")
	dir := t.TempDir()
	unwritable := filepath.Join(dir, "nope")
	if err := os.MkdirAll(unwritable, 0o555); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(unwritable, fs.ModeDir)

	destName := filepath.Join(unwritable, "frame.jpg")
	cmd := ExtractImageCmd(srcName, destName, opt)
	err := cmd.Run()
	assert.Error(t, err)
	assert.NoFileExists(t, destName)
}

// Negative: ffmpeg binary is missing; command execution should error immediately.
func TestExtractImageCmd_MissingBinary(t *testing.T) {
	opt := encode.NewPreviewImageOptions("/path/does/not/exist/ffmpeg", time.Second*1)
	srcName := fs.Abs("./testdata/25fps.vp9")
	destName := filepath.Join(t.TempDir(), "frame.jpg")
	cmd := ExtractImageCmd(srcName, destName, opt)
	err := cmd.Run()
	assert.Error(t, err)
}
