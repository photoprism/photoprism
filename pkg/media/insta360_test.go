package media

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/fs"
)

// writeInsta360Fixture writes a temporary media file with the specified payload.
func writeInsta360Fixture(t *testing.T, payload []byte) string {
	t.Helper()

	fileName := filepath.Join(t.TempDir(), "camera.insp")
	require.NoError(t, os.WriteFile(fileName, payload, fs.ModeFile))

	return fileName
}

// TestInsta360CameraModelFile verifies bounded OneRS metadata detection and safe fallbacks.
func TestInsta360CameraModelFile(t *testing.T) {
	t.Run("OneRSImage", func(t *testing.T) {
		payload := append([]byte(insta360OneRSImageField), bytes.Repeat([]byte{0x7f}, 256)...)
		model, err := Insta360CameraModelFile(writeInsta360Fixture(t, payload))
		require.NoError(t, err)
		assert.Equal(t, insta360OneRSModel, model)
	})
	t.Run("OneRSVideo", func(t *testing.T) {
		payload := append(bytes.Repeat([]byte{0x7f}, 256), insta360OneRSVideoField...)
		model, err := Insta360CameraModelFile(writeInsta360Fixture(t, payload))
		require.NoError(t, err)
		assert.Equal(t, insta360OneRSModel, model)
	})
	t.Run("PlainTextRejected", func(t *testing.T) {
		model, err := Insta360CameraModelFile(writeInsta360Fixture(t, []byte(insta360OneRSModel)))
		require.NoError(t, err)
		assert.Empty(t, model)
	})
	t.Run("TruncatedFieldRejected", func(t *testing.T) {
		model, err := Insta360CameraModelFile(writeInsta360Fixture(t, append([]byte{0x12, 0x0e}, []byte("Insta360 One")...)))
		require.NoError(t, err)
		assert.Empty(t, model)
	})
	t.Run("OutsideScanLimitsRejected", func(t *testing.T) {
		padding := bytes.Repeat([]byte{0}, insta360MetadataScanLimit+1)
		payload := append(append(append([]byte{}, padding...), insta360OneRSVideoField...), padding...)
		model, err := Insta360CameraModelFile(writeInsta360Fixture(t, payload))
		require.NoError(t, err)
		assert.Empty(t, model)
	})
	t.Run("Unknown", func(t *testing.T) {
		model, err := Insta360CameraModelFile(writeInsta360Fixture(t, []byte{0x12, 0x0b, 'O', 't', 'h', 'e', 'r'}))
		require.NoError(t, err)
		assert.Empty(t, model)
	})
	t.Run("Empty", func(t *testing.T) {
		model, err := Insta360CameraModelFile(writeInsta360Fixture(t, nil))
		require.NoError(t, err)
		assert.Empty(t, model)
	})
	t.Run("MissingFilename", func(t *testing.T) {
		model, err := Insta360CameraModelFile("")
		assert.Error(t, err)
		assert.Empty(t, model)
	})
	t.Run("MissingFile", func(t *testing.T) {
		model, err := Insta360CameraModelFile(filepath.Join(t.TempDir(), "missing.insv"))
		assert.Error(t, err)
		assert.Empty(t, model)
	})
}

// TestParseInsta360VideoName verifies strict capture identity and role parsing.
func TestParseInsta360VideoName(t *testing.T) {
	t.Run("Left", func(t *testing.T) {
		name, ok := ParseInsta360VideoName("videos/VID_20220625_140410_00_008.insv")
		require.True(t, ok)
		assert.Equal(t, Insta360VideoLeft, name.Role)
		assert.Equal(t, filepath.Join("videos", "20220625_140410_008"), name.CaptureKey())
		assert.Equal(t, filepath.Join("videos", "VID_20220625_140410_00_008.insv"), name.FileName(Insta360VideoLeft))
		assert.Equal(t, filepath.Join("videos", "VID_20220625_140410_10_008.insv"), name.FileName(Insta360VideoRight))
		assert.Equal(t, filepath.Join("videos", "LRV_20220625_140410_11_008.insv"), name.FileName(Insta360VideoProxy))
	})
	t.Run("RightUppercaseExtension", func(t *testing.T) {
		name, ok := ParseInsta360VideoName("VID_20220625_140410_10_008.INSV")
		require.True(t, ok)
		assert.Equal(t, Insta360VideoRight, name.Role)
	})
	t.Run("Proxy", func(t *testing.T) {
		name, ok := ParseInsta360VideoName("LRV_20220625_140410_11_008.insv")
		require.True(t, ok)
		assert.Equal(t, Insta360VideoProxy, name.Role)
	})

	invalid := []string{
		"VID_20220625_140410_11_008.insv",
		"LRV_20220625_140410_00_008.insv",
		"VID_20220625_140410_01_008.insv",
		"VID_20220625_140410_00_08.insv",
		"VID_20220625_140410_00_008.mp4",
		"copy-VID_20220625_140410_00_008.insv",
	}

	for _, fileName := range invalid {
		t.Run(fileName, func(t *testing.T) {
			_, ok := ParseInsta360VideoName(fileName)
			assert.False(t, ok)
		})
	}
}
