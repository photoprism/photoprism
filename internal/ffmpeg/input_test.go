package ffmpeg

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ffmpeg/encode"
)

// TestExtractJpegImageCmd_InputFormat extracts a still from files of the supported containers and from
// files whose content is not a supported container.
func TestExtractJpegImageCmd_InputFormat(t *testing.T) {
	bin, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skip("ffmpeg not found")
	}

	dir := t.TempDir()
	opt := &encode.Options{Bin: bin, SeekOffset: "00:00:00.000", TimeOffset: "00:00:00.000"}

	// writeClip encodes a short test clip with the given container format and video codec.
	writeClip := func(t *testing.T, fileName, format, codec string) {
		t.Helper()
		// #nosec G204 -- arguments are test constants.
		out, runErr := exec.Command(bin, "-v", "error", "-y", "-f", "lavfi", "-i", "testsrc=size=64x48:rate=5", "-t", "1",
			"-c:v", codec, "-f", format, fileName).CombinedOutput()
		require.NoError(t, runErr, string(out))
	}

	for name, clip := range map[string][3]string{
		"Mp4":             {"clip.mp4", "mp4", "mpeg4"},
		"Matroska":        {"clip.mkv", "matroska", "mpeg4"},
		"Avi":             {"clip.avi", "avi", "mpeg4"},
		"TransportStream": {"clip.mts", "mpegts", "mpeg2video"},
		"ProgramStream":   {"clip.mpg", "mpeg", "mpeg2video"},
		"Gif":             {"clip.gif", "gif", "gif"},
		"MultipartJpeg":   {"clip.mjpg", "mpjpeg", "mjpeg"},
	} {
		t.Run(name, func(t *testing.T) {
			fileName := filepath.Join(dir, name+"-"+clip[0])
			writeClip(t, fileName, clip[1], clip[2])
			jpegName := fileName + ".jpg"
			out, runErr := ExtractJpegImageCmd(fileName, jpegName, opt).CombinedOutput()
			require.NoError(t, runErr, string(out))
			assert.FileExists(t, jpegName)
		})
	}
	t.Run("Text", func(t *testing.T) {
		fileName := filepath.Join(dir, "text.mp4")
		require.NoError(t, os.WriteFile(fileName, []byte("not a video\n"), 0o600))
		jpegName := fileName + ".jpg"
		assert.Error(t, ExtractJpegImageCmd(fileName, jpegName, opt).Run())
		assert.NoFileExists(t, jpegName)
	})
	t.Run("Jpeg", func(t *testing.T) {
		fileName := filepath.Join(dir, "image.mp4")
		// #nosec G204 -- arguments are test constants.
		out, runErr := exec.Command(bin, "-v", "error", "-y", "-f", "lavfi", "-i", "testsrc=size=64x48", "-frames:v", "1",
			"-f", "image2", "-update", "1", fileName).CombinedOutput()
		require.NoError(t, runErr, string(out))
		jpegName := fileName + ".jpg"
		assert.Error(t, ExtractJpegImageCmd(fileName, jpegName, opt).Run())
		assert.NoFileExists(t, jpegName)
	})
}

// TestExtractJpegImageCmd_LiteralNames extracts stills from files whose names contain a percent sign.
func TestExtractJpegImageCmd_LiteralNames(t *testing.T) {
	bin, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skip("ffmpeg not found")
	}

	dir := t.TempDir()
	opt := &encode.Options{Bin: bin, SeekOffset: "00:00:00.000", TimeOffset: "00:00:00.000"}
	fileName := filepath.Join(dir, "v%d.mp4")

	// #nosec G204 -- arguments are test constants.
	out, err := exec.Command(bin, "-v", "error", "-y", "-f", "lavfi", "-i", "testsrc=size=64x48:rate=5", "-t", "1",
		"-c:v", "mpeg4", "-f", "mp4", fileName).CombinedOutput()
	require.NoError(t, err, string(out))

	t.Run("Jpeg", func(t *testing.T) {
		out, err = ExtractJpegImageCmd(fileName, fileName+".jpg", opt).CombinedOutput()
		require.NoError(t, err, string(out))
		assert.FileExists(t, fileName+".jpg")
		assert.NoFileExists(t, filepath.Join(dir, "v1.mp4.jpg"))
	})
	t.Run("Png", func(t *testing.T) {
		out, err = ExtractPngImageCmd(fileName, fileName+".png", opt).CombinedOutput()
		require.NoError(t, err, string(out))
		assert.FileExists(t, fileName+".png")
		assert.NoFileExists(t, filepath.Join(dir, "v1.mp4.png"))
	})
}
