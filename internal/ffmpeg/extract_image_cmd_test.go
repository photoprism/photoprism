package ffmpeg

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/ffmpeg/encode"
	"github.com/photoprism/photoprism/pkg/fs"
)

func TestExtractImageCmd(t *testing.T) {
	opt := encode.NewPreviewImageOptions("/usr/bin/ffmpeg", time.Second*9)

	srcName := fs.Abs("./testdata/25fps.vp9")
	destName := fs.Abs("./testdata/25fps.jpg")

	cmd := ExtractImageCmd(srcName, destName, opt)

	cmdStr := cmd.String()
	cmdStr = strings.Replace(cmdStr, srcName, "SRC", 1)
	cmdStr = strings.Replace(cmdStr, destName, "DEST", 1)

	assert.Equal(t, "/usr/bin/ffmpeg -y -strict -2 -ss 00:00:03.000 -i SRC -vframes 1 DEST", cmdStr)

	RunCommandTest(t, "jpg", srcName, destName, cmd, true)
}

func TestExtractJpegImageCmd(t *testing.T) {
	opt := encode.NewPreviewImageOptions("/usr/bin/ffmpeg", time.Second*9)

	srcName := fs.Abs("./testdata/25fps.vp9")
	destName := fs.Abs("./testdata/25fps.jpeg")

	cmd := ExtractJpegImageCmd(srcName, destName, opt)

	cmdStr := cmd.String()
	cmdStr = strings.Replace(cmdStr, srcName, "SRC", 1)
	cmdStr = strings.Replace(cmdStr, destName, "DEST", 1)

	assert.Equal(t, "/usr/bin/ffmpeg -y -strict -2 -ss 00:00:03.000 -i SRC -vframes 1 DEST", cmdStr)

	RunCommandTest(t, "jpeg", srcName, destName, cmd, true)
}

func TestExtractPngImageCmd(t *testing.T) {
	opt := encode.NewPreviewImageOptions("/usr/bin/ffmpeg", time.Second*9)

	srcName := fs.Abs("./testdata/25fps.vp9")
	destName := fs.Abs("./testdata/25fps.png")

	cmd := ExtractPngImageCmd(srcName, destName, opt)

	cmdStr := cmd.String()
	cmdStr = strings.Replace(cmdStr, srcName, "SRC", 1)
	cmdStr = strings.Replace(cmdStr, destName, "DEST", 1)

	assert.Equal(t, "/usr/bin/ffmpeg -y -strict -2 -ss 00:00:03.000 -i SRC -vframes 1 DEST", cmdStr)

	RunCommandTest(t, "png", srcName, destName, cmd, true)
}
