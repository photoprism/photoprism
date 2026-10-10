package ffmpeg

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

	assert.Equal(t, "/usr/bin/ffmpeg -hide_banner -loglevel error -y -f jpeg_pipe -i SRC -vf "+V360DualFisheyeToEquirect(190, 180)+" -frames:v 1 -update 1 DEST", cmdStr)
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

	assert.Contains(t, cmdStr, "-f mov -i LEFT -f mov -i RIGHT")
	assert.Contains(t, cmdStr, "[0:v:0][1:v:0]hstack=inputs=2:shortest=1,"+V360DualFisheyeToEquirect(204, 180)+"[v]")
	assert.Contains(t, cmdStr, "-map [v] -frames:v 1 -update 1 DEST")
}

// TestDewarpStackedDualFisheyeToJpegCmd verifies vertical lens splitting and horizontal stacking.
func TestDewarpStackedDualFisheyeToJpegCmd(t *testing.T) {
	opt := &encode.Options{Bin: "/usr/bin/ffmpeg", SizeLimit: 15360}
	cmd := DewarpStackedDualFisheyeToJpegCmd("SOURCE.jpg", "DEST", 204, 180, opt)
	cmdStr := cmd.String()

	assert.Contains(t, cmdStr, "-y -f jpeg_pipe -i SOURCE.jpg -filter_complex")

	assert.Contains(t, cmdStr, "[0:v:0]crop=iw:ih/2:0:0[top]")
	assert.Contains(t, cmdStr, "[0:v:0]crop=iw:ih/2:0:ih/2[bottom]")
	assert.Contains(t, cmdStr, "[top][bottom]hstack=inputs=2:shortest=1,v360=input=dfisheye:output=e")
	assert.Contains(t, cmdStr, "roll=180")
	assert.Contains(t, cmdStr, "min(15360, iw)")
	assert.Contains(t, cmdStr, "-map [v] -frames:v 1 -update 1 DEST")
}

// TestDewarpDualFisheyePairToAvcCmd verifies video, audio, and metadata mapping for paired lenses.
func TestDewarpDualFisheyePairToAvcCmd(t *testing.T) {
	opt := encode.NewVideoOptions("/usr/bin/ffmpeg", encode.SoftwareAvc, 1920, 23, "fast", "", "", "")
	opt.V360 = V360DualFisheyeToEquirect(204, 180)
	cmd := DewarpDualFisheyePairToAvcCmd("LEFT", "RIGHT", "DEST", opt)
	cmdStr := cmd.String()

	assert.Contains(t, cmdStr, "-f mov -i LEFT -f mov -i RIGHT")
	assert.Contains(t, cmdStr, "[0:v:0][1:v:0]hstack=inputs=2:shortest=1,v360=input=dfisheye:output=e")
	assert.Contains(t, cmdStr, "-map [v] -map 0:a:0?")
	assert.Contains(t, cmdStr, "-c:v libx264")
	assert.Contains(t, cmdStr, "-map_metadata 0 -shortest DEST")
}

// TestDewarpDualStreamToJpegCmd verifies that both lens streams of one input are stacked before dewarping.
func TestDewarpDualStreamToJpegCmd(t *testing.T) {
	opt := &encode.Options{Bin: "/usr/bin/ffmpeg", SizeLimit: 15360}
	cmd := DewarpDualStreamToJpegCmd("SOURCE", "DEST", 204, 0, opt)
	cmdStr := cmd.String()

	assert.Contains(t, cmdStr, "-y -f mov -i SOURCE -filter_complex")
	assert.NotContains(t, cmdStr, "-i SOURCE -i")
	assert.Contains(t, cmdStr, "[0:v:1][0:v:0]hstack=inputs=2:shortest=1,v360=input=dfisheye:output=e")
	assert.Contains(t, cmdStr, "min(15360, iw)")
	assert.Contains(t, cmdStr, "-map [v] -frames:v 1 -update 1 DEST")
}

// TestDewarpDualStreamToAvcCmd verifies video, audio, and metadata mapping for two streams in one input.
func TestDewarpDualStreamToAvcCmd(t *testing.T) {
	opt := encode.NewVideoOptions("/usr/bin/ffmpeg", encode.SoftwareAvc, 1920, 23, "fast", "", "", "")
	opt.V360 = V360DualFisheyeToEquirect(204, 0)
	cmd := DewarpDualStreamToAvcCmd("SOURCE", "DEST", opt)
	cmdStr := cmd.String()

	assert.Contains(t, cmdStr, "-strict -2 -f mov -i SOURCE -filter_complex")
	assert.Contains(t, cmdStr, "[0:v:1][0:v:0]hstack=inputs=2:shortest=1,v360=input=dfisheye:output=e")
	assert.Contains(t, cmdStr, "-map [v] -map 0:a:0?")
	assert.Contains(t, cmdStr, "-map_metadata 0 -shortest DEST")
}

func TestLensInputArgs(t *testing.T) {
	t.Run("Pair", func(t *testing.T) {
		assert.Equal(t, []string{"-f", "mov", "-i", "LEFT", "-f", "mov", "-i", "RIGHT"}, lensInputArgs([]string{"LEFT", "RIGHT"}))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Empty(t, lensInputArgs(nil))
	})
}

// TestDewarpDualFisheyePairToJpegCmd_InputFormat runs the command with lens files whose content is not an
// MP4 container, and with valid lens files.
func TestDewarpDualFisheyePairToJpegCmd_InputFormat(t *testing.T) {
	bin, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skip("ffmpeg not found")
	}

	dir := t.TempDir()
	opt := &encode.Options{Bin: bin}

	// writeLens encodes a short square test clip with the given container format and a built-in encoder.
	writeLens := func(t *testing.T, fileName, format string) {
		t.Helper()
		// #nosec G204 -- arguments are test constants.
		out, runErr := exec.Command(bin, "-v", "error", "-y", "-f", "lavfi", "-i", "testsrc=size=64x64:rate=5", "-t", "1",
			"-c:v", "mpeg4", "-f", format, fileName).CombinedOutput()
		require.NoError(t, runErr, string(out))
	}

	left := filepath.Join(dir, "left.insv")
	writeLens(t, left, "mp4")

	for name, content := range map[string]func(t *testing.T, fileName string){
		"Text": func(t *testing.T, fileName string) {
			require.NoError(t, os.WriteFile(fileName, []byte("not a video\n"), 0o600))
		},
		"Matroska":        func(t *testing.T, fileName string) { writeLens(t, fileName, "matroska") },
		"TransportStream": func(t *testing.T, fileName string) { writeLens(t, fileName, "mpegts") },
	} {
		t.Run(name, func(t *testing.T) {
			right := filepath.Join(dir, name+".insv")
			content(t, right)
			jpegName := filepath.Join(dir, name+".jpg")
			assert.Error(t, DewarpDualFisheyePairToJpegCmd(left, right, jpegName, 204, 0, opt).Run())
			assert.NoFileExists(t, jpegName)
		})
	}
	t.Run("Valid", func(t *testing.T) {
		right := filepath.Join(dir, "right.insv")
		writeLens(t, right, "mp4")
		jpegName := filepath.Join(dir, "valid.jpg")
		out, runErr := DewarpDualFisheyePairToJpegCmd(left, right, jpegName, 204, 0, opt).CombinedOutput()
		require.NoError(t, runErr, string(out))
		assert.FileExists(t, jpegName)
	})
}
