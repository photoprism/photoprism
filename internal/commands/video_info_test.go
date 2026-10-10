package commands

import (
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVideoRunFFprobe(t *testing.T) {
	ffmpegBin, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skip("ffmpeg not found")
	}

	ffprobeBin, err := exec.LookPath("ffprobe")
	if err != nil {
		t.Skip("ffprobe not found")
	}

	dir := t.TempDir()

	t.Run("MotionJpeg", func(t *testing.T) {
		fileName := filepath.Join(dir, "clip.mjpeg")
		// #nosec G204 -- arguments are test constants.
		out, runErr := exec.Command(ffmpegBin, "-v", "error", "-y", "-f", "lavfi", "-i", "testsrc=size=320x240:rate=5",
			"-frames:v", "1", "-c:v", "mjpeg", "-f", "mjpeg", fileName).CombinedOutput()
		require.NoError(t, runErr, string(out))

		parsed, raw, probeErr := videoRunFFprobe(ffprobeBin, fileName)
		require.NoError(t, probeErr)
		assert.NotNil(t, parsed)
		assert.Contains(t, raw, `"codec_name": "mjpeg"`)
	})
	t.Run("JpegAsMp4", func(t *testing.T) {
		fileName := filepath.Join(dir, "image.mp4")
		// #nosec G204 -- arguments are test constants.
		out, runErr := exec.Command(ffmpegBin, "-v", "error", "-y", "-f", "lavfi", "-i", "testsrc=size=320x240",
			"-frames:v", "1", "-f", "image2", "-update", "1", fileName).CombinedOutput()
		require.NoError(t, runErr, string(out))

		_, _, probeErr := videoRunFFprobe(ffprobeBin, fileName)
		assert.Error(t, probeErr)
	})
}
