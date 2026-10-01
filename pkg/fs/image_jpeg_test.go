package fs

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setMaxJpegScans changes MaxJpegScans for the duration of a test.
func setMaxJpegScans(t *testing.T, n int) {
	t.Helper()

	previous := MaxJpegScans
	MaxJpegScans = n
	t.Cleanup(func() { MaxJpegScans = previous })
}

// jpegScans returns a minimal JPEG stream with the given number of scans and entropy-coded data.
func jpegScans(scans int, data []byte) []byte {
	stream := []byte{0xFF, 0xD8}

	for range scans {
		stream = append(stream, 0xFF, 0xDA, 0x00, 0x02)
		stream = append(stream, data...)
	}

	return append(stream, 0xFF, 0xD9)
}

// TestCheckJpegScans verifies that the scans of a JPEG stream are counted without decoding it.
func TestCheckJpegScans(t *testing.T) {
	t.Run("Progressive", func(t *testing.T) {
		data, err := os.ReadFile(filepath.Join("testdata", "progressive.jpg"))
		require.NoError(t, err)
		require.Equal(t, 10, bytes.Count(data, []byte{0xFF, 0xDA}))
		assert.NoError(t, CheckJpegScans(bytes.NewReader(data)))
	})
	t.Run("Baseline", func(t *testing.T) {
		data, err := os.ReadFile(filepath.Join("testdata", "test.jpg"))
		require.NoError(t, err)
		assert.NoError(t, CheckJpegScans(bytes.NewReader(data)))
	})
	t.Run("AboveLimit", func(t *testing.T) {
		assert.ErrorIs(t, CheckJpegScans(bytes.NewReader(jpegScans(MaxJpegScans+1, nil))), ErrImageTooComplex)
		assert.NoError(t, CheckJpegScans(bytes.NewReader(jpegScans(MaxJpegScans, nil))))
	})
	t.Run("ProgressiveAboveLimit", func(t *testing.T) {
		setMaxJpegScans(t, 9)
		data, err := os.ReadFile(filepath.Join("testdata", "progressive.jpg"))
		require.NoError(t, err)
		assert.ErrorIs(t, CheckJpegScans(bytes.NewReader(data)), ErrImageTooComplex)
	})
	t.Run("EntropyCodedData", func(t *testing.T) {
		// Stuffed bytes, restart markers and fill bytes inside scan data are not counted as scans.
		data := []byte{0x12, 0xFF, 0x00, 0x34, 0xFF, 0xD0, 0x56, 0xFF, 0xFF, 0xD1, 0x78}
		setMaxJpegScans(t, 2)
		assert.NoError(t, CheckJpegScans(bytes.NewReader(jpegScans(2, data))))
		setMaxJpegScans(t, 1)
		assert.ErrorIs(t, CheckJpegScans(bytes.NewReader(jpegScans(2, data))), ErrImageTooComplex)
	})
	t.Run("NotJpeg", func(t *testing.T) {
		assert.NoError(t, CheckJpegScans(bytes.NewReader([]byte("\x89PNG\r\n\x1a\n"))))
		stream := jpegScans(MaxJpegScans+1, nil)
		assert.NoError(t, CheckJpegScans(bytes.NewReader(append([]byte{0x00, 0x00}, stream[2:]...))))
		assert.NoError(t, CheckJpegScans(bytes.NewReader(append([]byte{0xFF, 0xFB}, stream[2:]...))))
	})
	t.Run("AfterEndOfImage", func(t *testing.T) {
		stream := append(jpegScans(1, nil), jpegScans(MaxJpegScans+1, nil)[2:]...)
		assert.NoError(t, CheckJpegScans(bytes.NewReader(stream)))
	})
	t.Run("ShortSegmentLength", func(t *testing.T) {
		for _, segment := range [][]byte{{0xFF, 0xFE, 0x00, 0x00}, {0xFF, 0xE1, 0x00, 0x01}, {0xFF, 0xFE, 0x00, 0x02}} {
			stream := jpegScans(MaxJpegScans+1, nil)
			stream = append(append(append([]byte{}, stream[:2]...), segment...), stream[2:]...)
			assert.ErrorIs(t, CheckJpegScans(bytes.NewReader(stream)), ErrImageTooComplex)
		}
	})
	t.Run("Temporary", func(t *testing.T) {
		setMaxJpegScans(t, 1)
		stream := jpegScans(1, []byte{0xFF, 0x01, 0xFF, 0xDA, 0x00, 0x02})
		assert.ErrorIs(t, CheckJpegScans(bytes.NewReader(stream)), ErrImageTooComplex)
		stream = jpegScans(1, []byte{0xFF, 0x01, 0x12})
		assert.NoError(t, CheckJpegScans(bytes.NewReader(stream)))
	})
	t.Run("Truncated", func(t *testing.T) {
		assert.NoError(t, CheckJpegScans(bytes.NewReader([]byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00})))
	})
	t.Run("Disabled", func(t *testing.T) {
		setMaxJpegScans(t, 0)
		assert.NoError(t, CheckJpegScans(bytes.NewReader(jpegScans(1000, nil))))
	})
	t.Run("Nil", func(t *testing.T) {
		assert.NoError(t, CheckJpegScans(nil))
	})
}

// TestCheckJpegScansFile verifies the file variant of CheckJpegScans.
func TestCheckJpegScansFile(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		assert.NoError(t, CheckJpegScansFile(filepath.Join("testdata", "progressive.jpg")))
	})
	t.Run("AboveLimit", func(t *testing.T) {
		fileName := filepath.Join(t.TempDir(), "scans.jpg")
		require.NoError(t, os.WriteFile(fileName, jpegScans(MaxJpegScans+1, nil), ModeFile))
		assert.ErrorIs(t, CheckJpegScansFile(fileName), ErrImageTooComplex)
	})
	t.Run("Missing", func(t *testing.T) {
		assert.Error(t, CheckJpegScansFile(filepath.Join(t.TempDir(), "missing.jpg")))
	})
}

// TestDecodeImageFileJpegScans verifies that the decoder checks the scans before decoding.
func TestDecodeImageFileJpegScans(t *testing.T) {
	fileName := filepath.Join("testdata", "progressive.jpg")

	t.Run("WithinLimit", func(t *testing.T) {
		img, format, err := DecodeImageFile(fileName)
		require.NoError(t, err)
		assert.Equal(t, "jpeg", format)
		assert.Equal(t, 64, img.Bounds().Dx())
	})
	t.Run("AboveLimit", func(t *testing.T) {
		setMaxJpegScans(t, 9)
		_, _, err := DecodeImageFile(fileName)
		assert.True(t, errors.Is(err, ErrImageTooComplex))
	})
}
