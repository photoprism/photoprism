package media

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/fs"
)

func TestFileTypes(t *testing.T) {
	t.Run("Raw", func(t *testing.T) {
		// Raw must cover the Adobe Digital Negative type as well as generic sensor data.
		result := FileTypes(Raw)
		assert.Equal(t, []fs.Type{fs.ImageDng, fs.ImageRaw}, result)
	})
	t.Run("Sorted", func(t *testing.T) {
		result := FileTypes(Video)
		assert.Greater(t, len(result), 1)
		assert.IsIncreasing(t, result)
	})
	t.Run("Video", func(t *testing.T) {
		result := FileTypes(Video)
		assert.Contains(t, result, fs.VideoMp4)
		assert.Contains(t, result, fs.VideoMov)
		assert.NotContains(t, result, fs.ImageJpeg)
	})
	t.Run("Image", func(t *testing.T) {
		// Cineon is rendered by the generic ImageMagick path, so it must report as an
		// image for the converter to offer a command for it at all.
		result := FileTypes(Image)
		assert.Contains(t, result, fs.ImageCineon)
		assert.NotContains(t, result, fs.ImageRaw)
	})
	t.Run("Unknown", func(t *testing.T) {
		assert.Empty(t, FileTypes(Type("invalid")))
	})
}

func TestFileTypeStrings(t *testing.T) {
	t.Run("Raw", func(t *testing.T) {
		assert.Equal(t, []string{"dng", "raw"}, FileTypeStrings(Raw))
	})
	t.Run("Unknown", func(t *testing.T) {
		assert.Empty(t, FileTypeStrings(Type("invalid")))
	})
}

func TestImageTypesExceptJpeg(t *testing.T) {
	t.Run("Contents", func(t *testing.T) {
		result := ImageTypesExceptJpeg()

		assert.NotEmpty(t, result)
		assert.NotContains(t, result, fs.ImageJpeg.String())
		assert.Contains(t, result, fs.ImagePng.String())
		assert.Contains(t, result, fs.ImagePsd.String())
		assert.Contains(t, result, fs.ImageCineon.String())
		assert.NotContains(t, result, fs.VideoMp4.String())
		assert.NotContains(t, result, fs.ImageRaw.String())
	})
	t.Run("CoversEveryImageType", func(t *testing.T) {
		for _, fileType := range FileTypes(Image) {
			if fileType == fs.ImageJpeg {
				continue
			}
			assert.Containsf(t, ImageTypesExceptJpeg(), fileType.String(), "%s is missing", fileType)
		}
	})
	t.Run("AppendReallocates", func(t *testing.T) {
		// Every call hands out the same backing array, so spare capacity would let one
		// caller's append overwrite another's entry. Full capacity forces a copy instead.
		first := append(ImageTypesExceptJpeg(), "first")
		second := append(ImageTypesExceptJpeg(), "second")

		assert.Equal(t, "first", first[len(first)-1])
		assert.Equal(t, "second", second[len(second)-1])
	})
}
