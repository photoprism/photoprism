package ffmpeg

import (
	"path/filepath"
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

	assert.Equal(t, "/usr/bin/ffmpeg -hide_banner -loglevel error -y -strict -2 -hwaccel none -err_detect ignore_err -ss 00:00:00.000 -i SRC -ss 00:00:00.001 -vf setparams=range=tv:color_primaries=bt709:color_trc=bt709:colorspace=bt709,scale=trunc(iw/2)*2:trunc(ih/2)*2,setsar=1,format=yuvj422p -frames:v 1 DEST", cmdStr)

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

	assert.Equal(t, "/usr/bin/ffmpeg -hide_banner -loglevel error -y -strict -2 -hwaccel none -err_detect ignore_err -ss 00:00:00.000 -i SRC -ss 00:00:00.001 -vf setparams=range=tv:color_primaries=bt709:color_trc=bt709:colorspace=bt709,scale=trunc(iw/2)*2:trunc(ih/2)*2,setsar=1,format=yuvj422p -frames:v 1 DEST", cmdStr)

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

	assert.Equal(t, "/usr/bin/ffmpeg -hide_banner -loglevel error -y -strict -2 -hwaccel none -err_detect ignore_err -ss 00:00:00.000 -i SRC -ss 00:00:00.001 -vf scale=trunc(iw/2)*2:trunc(ih/2)*2,setsar=1 -frames:v 1 DEST", cmdStr)

	RunCommandTest(t, "png", srcName, destName, cmd, true)
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
