package fs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFileType(t *testing.T) {
	t.Run("jpeg", func(t *testing.T) {
		result := GetFileType("testdata/test.jpg")
		assert.Equal(t, TypeJpeg, result)
	})

	t.Run("raw", func(t *testing.T) {
		result := GetFileType("testdata/test (jpg).CR2")
		assert.Equal(t, TypeRaw, result)
	})

	t.Run("empty", func(t *testing.T) {
		result := GetFileType("")
		assert.Equal(t, TypeOther, result)
	})
}

func TestFileType_Find(t *testing.T) {
	t.Run("find jpg", func(t *testing.T) {
		result := TypeJpeg.Find("testdata/test.xmp", false)
		assert.Equal(t, "testdata/test.jpg", result)
	})

	t.Run("upper ext", func(t *testing.T) {
		result := TypeJpeg.Find("testdata/test.XMP", false)
		assert.Equal(t, "testdata/test.jpg", result)
	})

	t.Run("with sequence", func(t *testing.T) {
		result := TypeJpeg.Find("testdata/test (2).xmp", false)
		assert.Equal(t, "", result)
	})

	t.Run("strip sequence", func(t *testing.T) {
		result := TypeJpeg.Find("testdata/test (2).xmp", true)
		assert.Equal(t, "testdata/test.jpg", result)
	})
	t.Run("prefixUpper", func(t *testing.T) {
		result := TypeJpeg.Find("testdata/catyellow.xmp", true)
		assert.Equal(t, "testdata/CATYELLOW.jpg", result)
	})
	t.Run("prefixLower", func(t *testing.T) {
		result := TypeJpeg.Find("testdata/CHAMELEON_LIME.xmp", true)
		assert.Equal(t, "testdata/chameleon_lime.jpg", result)
	})
}
