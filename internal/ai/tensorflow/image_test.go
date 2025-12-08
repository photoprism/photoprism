package tensorflow

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wamuir/graft/tensorflow"
)

var defaultImageInput = &PhotoInput{
	Height: 224,
	Width:  224,
	Shape:  DefaultPhotoInputShape(),
}

var examplesPath = filepath.Join(assetsPath, "examples")

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
		imageBuffer, err := os.ReadFile(filepath.Join(examplesPath, "cat_brown.jpg")) //nolint:gosec // reading bundled test fixture

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
		imageBuffer, err := os.ReadFile(filepath.Join(examplesPath, "Random.docx")) //nolint:gosec // reading bundled test fixture
		assert.Nil(t, err)
		result, err := ImageFromBytes(imageBuffer, defaultImageInput, nil)

		assert.Empty(t, result)
		assert.EqualError(t, err, "image: unknown format")
	})
}
