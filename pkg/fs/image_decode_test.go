package fs

import (
	"bytes"
	"image"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func malformedLittleEndianTIFFIfdOffset() []byte {
	return []byte{0x49, 0x49, 0x2a, 0x00, 0xff, 0xff, 0xff, 0xff}
}

func malformedBigEndianTIFFIfdOffset() []byte {
	return []byte{0x4d, 0x4d, 0x00, 0x2a, 0xff, 0xff, 0xff, 0xff}
}

func truncatedTIFFBody() []byte {
	return []byte{0x49, 0x49, 0x2a, 0x00, 0x08, 0x00, 0x00, 0x00, 0x00}
}

func TestDecodeImageConfigFile(t *testing.T) {
	t.Run("Png", func(t *testing.T) {
		fileName := filepath.Join(t.TempDir(), "pixel.png")
		buf := bytes.NewBuffer(nil)
		require.NoError(t, png.Encode(buf, image.NewNRGBA(image.Rect(0, 0, 1, 1))))
		require.NoError(t, os.WriteFile(fileName, buf.Bytes(), ModeFile))

		cfg, format, err := DecodeImageConfigFile(fileName)

		require.NoError(t, err)
		assert.Equal(t, "png", format)
		assert.Equal(t, 1, cfg.Width)
		assert.Equal(t, 1, cfg.Height)
	})
	t.Run("LayeredTiff", func(t *testing.T) {
		fileName := Abs("../../assets/samples/layered-16bit-small.tif")

		cfg, format, err := DecodeImageConfigFile(fileName)

		require.NoError(t, err)
		assert.Equal(t, "tiff", format)
		assert.Equal(t, 236, cfg.Width)
		assert.Equal(t, 158, cfg.Height)
	})
	t.Run("MalformedTiffIfdOffset", func(t *testing.T) {
		fileName := filepath.Join(t.TempDir(), "evil.tiff")
		payload := malformedLittleEndianTIFFIfdOffset()
		require.NoError(t, os.WriteFile(fileName, payload, ModeFile))

		_, _, err := DecodeImageConfigFile(fileName)

		require.Error(t, err)
	})
	t.Run("MalformedBigEndianTiffIfdOffset", func(t *testing.T) {
		fileName := filepath.Join(t.TempDir(), "evil-mm.tiff")
		require.NoError(t, os.WriteFile(fileName, malformedBigEndianTIFFIfdOffset(), ModeFile))

		_, _, err := DecodeImageConfigFile(fileName)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid TIFF: IFD offset")
	})
	t.Run("TruncatedTiffBody", func(t *testing.T) {
		fileName := filepath.Join(t.TempDir(), "truncated.tiff")
		require.NoError(t, os.WriteFile(fileName, truncatedTIFFBody(), ModeFile))

		_, _, err := DecodeImageConfigFile(fileName)

		require.Error(t, err)
		assert.ErrorIs(t, err, io.EOF)
	})
}

func TestDecodeImageFile(t *testing.T) {
	t.Run("Png", func(t *testing.T) {
		fileName := filepath.Join(t.TempDir(), "pixel.png")
		buf := bytes.NewBuffer(nil)
		require.NoError(t, png.Encode(buf, image.NewNRGBA(image.Rect(0, 0, 2, 3))))
		require.NoError(t, os.WriteFile(fileName, buf.Bytes(), ModeFile))

		img, format, err := DecodeImageFile(fileName)

		require.NoError(t, err)
		assert.Equal(t, "png", format)
		assert.Equal(t, 2, img.Bounds().Dx())
		assert.Equal(t, 3, img.Bounds().Dy())
	})
	t.Run("LayeredTiff", func(t *testing.T) {
		fileName := Abs("../../assets/samples/layered-16bit-small.tif")

		img, format, err := DecodeImageFile(fileName)

		require.NoError(t, err)
		assert.Equal(t, "tiff", format)
		assert.Equal(t, 236, img.Bounds().Dx())
		assert.Equal(t, 158, img.Bounds().Dy())
	})
	t.Run("MalformedTiffIfdOffset", func(t *testing.T) {
		fileName := filepath.Join(t.TempDir(), "evil.tiff")
		payload := malformedLittleEndianTIFFIfdOffset()
		require.NoError(t, os.WriteFile(fileName, payload, ModeFile))

		_, _, err := DecodeImageFile(fileName)

		require.Error(t, err)
	})
	t.Run("MalformedBigEndianTiffIfdOffset", func(t *testing.T) {
		fileName := filepath.Join(t.TempDir(), "evil-mm.tiff")
		require.NoError(t, os.WriteFile(fileName, malformedBigEndianTIFFIfdOffset(), ModeFile))

		_, _, err := DecodeImageFile(fileName)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid TIFF: IFD offset")
	})
	t.Run("TruncatedTiffBody", func(t *testing.T) {
		fileName := filepath.Join(t.TempDir(), "truncated.tiff")
		require.NoError(t, os.WriteFile(fileName, truncatedTIFFBody(), ModeFile))

		_, _, err := DecodeImageFile(fileName)

		require.Error(t, err)
		assert.ErrorIs(t, err, io.EOF)
	})
}

func TestDecodeImageData(t *testing.T) {
	t.Run("MalformedBigEndianTiffIfdOffset", func(t *testing.T) {
		_, _, err := DecodeImageData(malformedBigEndianTIFFIfdOffset())

		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid TIFF: IFD offset")
	})
	t.Run("TruncatedTiffBody", func(t *testing.T) {
		_, _, err := DecodeImageData(truncatedTIFFBody())

		require.Error(t, err)
		assert.ErrorIs(t, err, io.EOF)
	})
}

func TestDecodeImageConfigData(t *testing.T) {
	t.Run("MalformedBigEndianTiffIfdOffset", func(t *testing.T) {
		_, _, err := DecodeImageConfigData(malformedBigEndianTIFFIfdOffset())

		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid TIFF: IFD offset")
	})
	t.Run("TruncatedTiffBody", func(t *testing.T) {
		_, _, err := DecodeImageConfigData(truncatedTIFFBody())

		require.Error(t, err)
		assert.ErrorIs(t, err, io.EOF)
	})
}

func TestValidateTIFFOffset(t *testing.T) {
	t.Run("RejectsLittleEndianOffsetBeyondSize", func(t *testing.T) {
		err := validateTIFFOffset(malformedLittleEndianTIFFIfdOffset(), 8)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid TIFF: IFD offset")
	})
	t.Run("RejectsBigEndianOffsetBeyondSize", func(t *testing.T) {
		err := validateTIFFOffset(malformedBigEndianTIFFIfdOffset(), 8)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid TIFF: IFD offset")
	})
	t.Run("AcceptsTruncatedBodyOffsetWithinSize", func(t *testing.T) {
		err := validateTIFFOffset(truncatedTIFFBody(), int64(len(truncatedTIFFBody())))
		require.NoError(t, err)
	})
}
