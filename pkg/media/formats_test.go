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
