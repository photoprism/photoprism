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
	assert.Equal(t, "v360=input=dfisheye:output=e:ih_fov=204:iv_fov=204", V360DualFisheyeToEquirect(204, 0))
	assert.Equal(t, "v360=input=dfisheye:output=e:ih_fov=190:iv_fov=190:roll=180", V360DualFisheyeToEquirect(190, 180))
}

func TestV360FisheyeToEquirect(t *testing.T) {
	assert.Equal(t, "v360=input=fisheye:output=e:ih_fov=204:iv_fov=204", V360FisheyeToEquirect(204, 0))
	assert.Equal(t, "v360=input=fisheye:output=e:ih_fov=190:iv_fov=190:roll=90", V360FisheyeToEquirect(190, 90))
}

func TestDewarpDualFisheyeToJpegCmd(t *testing.T) {
	opt := &encode.Options{Bin: "/usr/bin/ffmpeg"}

	srcName := fs.Abs("./testdata/dualfisheye.insp")
	destName := fs.Abs("./testdata/dualfisheye.jpg")

	cmd := DewarpDualFisheyeToJpegCmd(srcName, destName, 190, 180, opt)

	cmdStr := cmd.String()
	cmdStr = strings.Replace(cmdStr, srcName, "SRC", 1)
	cmdStr = strings.Replace(cmdStr, destName, "DEST", 1)

	assert.Equal(t, "/usr/bin/ffmpeg -hide_banner -loglevel error -y -i SRC -vf "+V360DualFisheyeToEquirect(190, 180)+" -frames:v 1 DEST", cmdStr)
}

// Negative: ffmpeg binary is missing; command execution should error immediately.
func TestDewarpDualFisheyeToJpegCmd_MissingBinary(t *testing.T) {
	opt := &encode.Options{Bin: "/path/does/not/exist/ffmpeg"}
	srcName := fs.Abs("./testdata/dualfisheye.insp")
	destName := filepath.Join(t.TempDir(), "frame.jpg")
	cmd := DewarpDualFisheyeToJpegCmd(srcName, destName, 204, 0, opt)
	err := cmd.Run()
	assert.Error(t, err)
}

// TestDewarpDualFisheyePairToJpegCmd verifies canonical lens ordering and filter composition.
func TestDewarpDualFisheyePairToJpegCmd(t *testing.T) {
	opt := &encode.Options{Bin: "/usr/bin/ffmpeg"}
	cmd := DewarpDualFisheyePairToJpegCmd("LEFT", "RIGHT", "DEST", 204, 180, opt)
	cmdStr := cmd.String()

	assert.Contains(t, cmdStr, "-i LEFT -i RIGHT")
	assert.Contains(t, cmdStr, "[0:v:0][1:v:0]hstack=inputs=2:shortest=1,"+V360DualFisheyeToEquirect(204, 180)+"[v]")
	assert.Contains(t, cmdStr, "-map [v] -frames:v 1 DEST")
}

// TestDewarpStackedDualFisheyeToJpegCmd verifies vertical lens splitting and horizontal stacking.
func TestDewarpStackedDualFisheyeToJpegCmd(t *testing.T) {
	opt := &encode.Options{Bin: "/usr/bin/ffmpeg", SizeLimit: 15360}
	cmd := DewarpStackedDualFisheyeToJpegCmd("SOURCE", "DEST", 204, 180, opt)
	cmdStr := cmd.String()

	assert.Contains(t, cmdStr, "[0:v:0]crop=iw:ih/2:0:0[top]")
	assert.Contains(t, cmdStr, "[0:v:0]crop=iw:ih/2:0:ih/2[bottom]")
	assert.Contains(t, cmdStr, "[top][bottom]hstack=inputs=2:shortest=1,v360=input=dfisheye:output=e")
	assert.Contains(t, cmdStr, "roll=180")
	assert.Contains(t, cmdStr, "min(15360, iw)")
	assert.Contains(t, cmdStr, "-map [v] -frames:v 1 DEST")
}

// TestDewarpDualFisheyePairToAvcCmd verifies video, audio, and metadata mapping for paired lenses.
func TestDewarpDualFisheyePairToAvcCmd(t *testing.T) {
	opt := encode.NewVideoOptions("/usr/bin/ffmpeg", encode.SoftwareAvc, 1920, 23, "fast", "", "", "")
	opt.V360 = V360DualFisheyeToEquirect(204, 180)
	cmd := DewarpDualFisheyePairToAvcCmd("LEFT", "RIGHT", "DEST", opt)
	cmdStr := cmd.String()

	assert.Contains(t, cmdStr, "-i LEFT -i RIGHT")
	assert.Contains(t, cmdStr, "[0:v:0][1:v:0]hstack=inputs=2:shortest=1,v360=input=dfisheye:output=e")
	assert.Contains(t, cmdStr, "-map [v] -map 0:a:0?")
	assert.Contains(t, cmdStr, "-c:v libx264")
	assert.Contains(t, cmdStr, "-map_metadata 0 -shortest DEST")
}
