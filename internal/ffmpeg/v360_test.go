package ffmpeg

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/ffmpeg/encode"
	"github.com/photoprism/photoprism/pkg/fs"
)

func TestV360DualFisheyeToEquirect(t *testing.T) {
	assert.Equal(t, "v360=input=dfisheye:output=e:ih_fov=204:iv_fov=204", V360DualFisheyeToEquirect(204))
	assert.Equal(t, "v360=input=dfisheye:output=e:ih_fov=190:iv_fov=190", V360DualFisheyeToEquirect(190))
}

func TestDewarpDualFisheyeToJpegCmd(t *testing.T) {
	opt := &encode.Options{Bin: "/usr/bin/ffmpeg"}

	srcName := fs.Abs("./testdata/dualfisheye.insp")
	destName := fs.Abs("./testdata/dualfisheye.jpg")

	cmd := DewarpDualFisheyeToJpegCmd(srcName, destName, 190, opt)

	cmdStr := cmd.String()
	cmdStr = strings.Replace(cmdStr, srcName, "SRC", 1)
	cmdStr = strings.Replace(cmdStr, destName, "DEST", 1)

	assert.Equal(t, "/usr/bin/ffmpeg -hide_banner -loglevel error -y -i SRC -vf "+V360DualFisheyeToEquirect(190)+" -frames:v 1 DEST", cmdStr)
}

// Negative: ffmpeg binary is missing; command execution should error immediately.
func TestDewarpDualFisheyeToJpegCmd_MissingBinary(t *testing.T) {
	opt := &encode.Options{Bin: "/path/does/not/exist/ffmpeg"}
	srcName := fs.Abs("./testdata/dualfisheye.insp")
	destName := filepath.Join(t.TempDir(), "frame.jpg")
	cmd := DewarpDualFisheyeToJpegCmd(srcName, destName, 204, opt)
	err := cmd.Run()
	assert.Error(t, err)
}
