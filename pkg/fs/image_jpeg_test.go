package fs

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
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
		data := []byte{0x12, 0xFF, 0x00, 0x34, 0xFF, 0xD0, 0x56, 0xFF, 0xFF, 0xD1, 0x78, 0xFF, 0xD7, 0x9A}
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
		// Data appended after EOI, such as the video of a motion photo, is not read.
		trailer := []byte{0x00, 0x00, 0x00, 0x18, 'f', 't', 'y', 'p', 'm', 'p', '4', '2'}
		stream := append(append(jpegScans(1, nil), trailer...), jpegScans(MaxJpegScans+1, nil)[2:]...)
		assert.NoError(t, CheckJpegScans(bytes.NewReader(stream)))
	})
	t.Run("SecondStartOfImage", func(t *testing.T) {
		// A frame without EOI ends at the next SOI, e.g. in a Motion JPEG stream.
		first := jpegScans(1, nil)
		second := append([]byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10}, make([]byte, 70000)...)
		stream := append(append(first[:len(first)-2], second...), jpegScans(MaxJpegScans+1, nil)[2:]...)
		assert.NoError(t, CheckJpegScans(bytes.NewReader(stream)))
	})
	t.Run("SegmentPayload", func(t *testing.T) {
		// The markers of an embedded thumbnail are part of its segment and not read.
		thumbnail := jpegScans(1, nil)
		app1 := append([]byte{0xFF, 0xE1, byte((len(thumbnail) + 2) >> 8), byte(len(thumbnail) + 2)}, thumbnail...) //nolint:gosec // G115: the segment length is small
		stream := append(append([]byte{0xFF, 0xD8}, app1...), jpegScans(MaxJpegScans+1, nil)[2:]...)
		assert.ErrorIs(t, CheckJpegScans(bytes.NewReader(stream)), ErrImageTooComplex)
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

// TestJpegFrameConfig verifies that the frame size is read from the JPEG frame header.
func TestJpegFrameConfig(t *testing.T) {
	t.Run("Arithmetic", func(t *testing.T) {
		data, err := os.ReadFile(filepath.Join("testdata", "arithmetic.jpg"))
		require.NoError(t, err)
		cfg, err := jpegFrameConfig(bytes.NewReader(data))
		require.NoError(t, err)
		assert.Equal(t, 64, cfg.Width)
		assert.Equal(t, 48, cfg.Height)
		assert.Equal(t, color.YCbCrModel, cfg.ColorModel)
	})
	t.Run("MatchesDecoder", func(t *testing.T) {
		for _, name := range []string{"test.jpg", "progressive.jpg", "CATYELLOW.jpg"} {
			data, err := os.ReadFile(filepath.Join("testdata", name)) //nolint:gosec // G304: test fixture path
			require.NoError(t, err)
			expected, err := jpeg.DecodeConfig(bytes.NewReader(data))
			require.NoError(t, err)
			cfg, err := jpegFrameConfig(bytes.NewReader(data))
			require.NoError(t, err, name)
			assert.Equal(t, expected.Width, cfg.Width, name)
			assert.Equal(t, expected.Height, cfg.Height, name)
		}
	})
	t.Run("Gray", func(t *testing.T) {
		stream := []byte{0xFF, 0xD8, 0xFF, 0xFF, 0xC0, 0x00, 0x0B, 0x08, 0x00, 0x30, 0x00, 0x40, 0x01, 0x01, 0x11, 0x00}
		cfg, err := jpegFrameConfig(bytes.NewReader(stream))
		require.NoError(t, err)
		assert.Equal(t, image.Config{Width: 64, Height: 48, ColorModel: color.GrayModel}, cfg)
	})
	t.Run("TablesBeforeFrame", func(t *testing.T) {
		// DHT and DAC carry no frame size, even though their marker codes are in the SOF range.
		stream := []byte{0xFF, 0xD8, 0xFF, 0xC4, 0x00, 0x08, 0x00, 0x10, 0x00, 0x20, 0x00, 0x40, 0xFF, 0xCC, 0x00, 0x04, 0x00, 0x10,
			0xFF, 0xC9, 0x00, 0x0B, 0x08, 0x00, 0x30, 0x00, 0x40, 0x01, 0x01, 0x11, 0x00}
		cfg, err := jpegFrameConfig(bytes.NewReader(stream))
		require.NoError(t, err)
		assert.Equal(t, 64, cfg.Width)
		assert.Equal(t, 48, cfg.Height)
	})
	t.Run("ExtraneousBytes", func(t *testing.T) {
		// Bytes between segments and a stuffed zero are skipped like libjpeg's next_marker does.
		stream := []byte{0xFF, 0xD8, 0x12, 0x34, 0xFF, 0x00, 0xFF, 0xC0, 0x00, 0x0B, 0x08, 0x00, 0x30, 0x00, 0x40, 0x01, 0x01, 0x11, 0x00}
		cfg, err := jpegFrameConfig(bytes.NewReader(stream))
		require.NoError(t, err)
		assert.Equal(t, 64, cfg.Width)
		assert.Equal(t, 48, cfg.Height)
	})
	t.Run("Invalid", func(t *testing.T) {
		for name, stream := range map[string][]byte{
			"NotJpeg":     []byte("\x89PNG\r\n\x1a\n"),
			"NoFrame":     jpegScans(1, nil),
			"ZeroHeight":  {0xFF, 0xD8, 0xFF, 0xC0, 0x00, 0x0B, 0x08, 0x00, 0x00, 0x00, 0x40, 0x01, 0x01, 0x11, 0x00},
			"ShortHeader": {0xFF, 0xD8, 0xFF, 0xC0, 0x00, 0x05, 0x08, 0x00, 0x30},
			"BadLength":   {0xFF, 0xD8, 0xFF, 0xC0, 0x00, 0x0C, 0x08, 0x00, 0x30, 0x00, 0x40, 0x01, 0x01, 0x11, 0x00, 0x00},
			"Components":  {0xFF, 0xD8, 0xFF, 0xC0, 0x00, 0x17, 0x08, 0x00, 0x30, 0x00, 0x40, 0x05},
			"Truncated":   {0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A},
			"NoMarker":    {0xFF, 0xD8, 0x12, 0x34},
			"StuffedOnly": {0xFF, 0xD8, 0xFF, 0x00, 0xFF, 0x00},
		} {
			_, err := jpegFrameConfig(bytes.NewReader(stream))
			assert.Error(t, err, name)
		}
	})
}

// TestDecodeImageConfigFileArithmetic verifies that the size of a JPEG the Go decoder cannot read is
// taken from its frame header, while decoding it still fails.
func TestDecodeImageConfigFileArithmetic(t *testing.T) {
	fileName := filepath.Join("testdata", "arithmetic.jpg")
	cfg, format, err := DecodeImageConfigFile(fileName)
	require.NoError(t, err)
	assert.Equal(t, "jpeg", format)
	assert.Equal(t, 64, cfg.Width)
	assert.Equal(t, 48, cfg.Height)
	_, _, err = DecodeImageFile(fileName)
	assert.Error(t, err)
}
