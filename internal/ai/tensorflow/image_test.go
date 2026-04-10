package tensorflow

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wamuir/graft/tensorflow"

	"github.com/photoprism/photoprism/pkg/fs"
)

var defaultImageInput = &PhotoInput{
	Height: 224,
	Width:  224,
	Shape:  DefaultPhotoInputShape(),
}

var samplesPath = filepath.Join(assetsPath, "samples")

func TestConvertValue(t *testing.T) {
	result := convertValue(uint32(98765432), &Interval{Start: -1, End: 1})
	assert.Equal(t, float32(3024.8982), result)
}

func TestConvertStdMean(t *testing.T) {
	mean := float32(1.0 / 127.5)
	stdDev := float32(-1.0)

	result := convertValue(uint32(98765432), &Interval{Mean: &mean, StdDev: &stdDev})
	assert.Equal(t, float32(3024.8982), result)
}

func TestImageFromBytes(t *testing.T) {
	t.Run("CatJpeg", func(t *testing.T) {
		imageBuffer, err := os.ReadFile(filepath.Join(samplesPath, "cat_brown.jpg")) //nolint:gosec // reading bundled test fixture

		if err != nil {
			t.Fatal(err)
		}

		result, err := ImageFromBytes(imageBuffer, defaultImageInput, nil)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, tensorflow.DataType(0x1), result.DataType())
		assert.Equal(t, int64(1), result.Shape()[0])
		assert.Equal(t, int64(224), result.Shape()[2])
	})
	t.Run("Document", func(t *testing.T) {
		imageBuffer, err := os.ReadFile(filepath.Join(samplesPath, "Random.docx")) //nolint:gosec // reading bundled test fixture
		assert.Nil(t, err)
		result, err := ImageFromBytes(imageBuffer, defaultImageInput, nil)

		assert.Empty(t, result)
		assert.EqualError(t, err, "unsupported image format")
	})
	t.Run("MalformedTiffIfdOffset", func(t *testing.T) {
		imageBuffer := []byte{0x49, 0x49, 0x2a, 0x00, 0xff, 0xff, 0xff, 0xff}

		result, err := ImageFromBytes(imageBuffer, defaultImageInput, nil)

		assert.Nil(t, result)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid TIFF: IFD offset")
	})
	t.Run("TruncatedTiffBody", func(t *testing.T) {
		imageBuffer := []byte{0x49, 0x49, 0x2a, 0x00, 0x08, 0x00, 0x00, 0x00, 0x00}

		result, err := ImageFromBytes(imageBuffer, defaultImageInput, nil)

		assert.Nil(t, result)
		require.Error(t, err)
		assert.ErrorIs(t, err, io.EOF)
	})
}

func TestOpenImage(t *testing.T) {
	t.Run("MalformedTiffIfdOffset", func(t *testing.T) {
		fileName := filepath.Join(t.TempDir(), "evil.tiff")
		require.NoError(t, os.WriteFile(fileName, []byte{0x49, 0x49, 0x2a, 0x00, 0xff, 0xff, 0xff, 0xff}, fs.ModeFile))

		result, err := OpenImage(fileName)

		assert.Nil(t, result)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid TIFF: IFD offset")
	})
}
