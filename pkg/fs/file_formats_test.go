package fs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFileFormat(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		result := GetFileFormat("")
		assert.Equal(t, FormatOther, result)
	})
	t.Run("JPEG", func(t *testing.T) {
		result := GetFileFormat("testdata/test.jpg")
		assert.Equal(t, FormatJpeg, result)
	})
	t.Run("RawCRw", func(t *testing.T) {
		result := GetFileFormat("testdata/test (jpg).crw")
		assert.Equal(t, FormatRaw, result)
	})
	t.Run("RawCR2", func(t *testing.T) {
		result := GetFileFormat("testdata/test (jpg).CR2")
		assert.Equal(t, FormatRaw, result)
	})
	t.Run("MP4", func(t *testing.T) {
		assert.Equal(t, FileFormat("mp4"), GetFileFormat("file.mp"))
	})
}
