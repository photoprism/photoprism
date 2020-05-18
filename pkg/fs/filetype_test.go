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
	t.Run("name upper", func(t *testing.T) {
		result := TypeJpeg.Find("testdata/CATYELLOW.xmp", true)
		assert.Equal(t, "testdata/CATYELLOW.jpg", result)
	})
	t.Run("name lower", func(t *testing.T) {
		result := TypeJpeg.Find("testdata/chameleon_lime.xmp", true)
		assert.Equal(t, "testdata/chameleon_lime.jpg", result)
	})
}

func TestFileType_FindHidden(t *testing.T) {
	hiddenPath := ".photoprism"

	t.Run("find xmp", func(t *testing.T) {
		result := TypeXMP.FindSub("testdata/test.jpg", hiddenPath, false)
		assert.Equal(t, "testdata/.photoprism/test.xmp", result)
	})

	t.Run("find xmp upper ext", func(t *testing.T) {
		result := TypeXMP.FindSub("testdata/test.PNG", hiddenPath, false)
		assert.Equal(t, "testdata/.photoprism/test.xmp", result)
	})

	t.Run("find xmp without sequence", func(t *testing.T) {
		result := TypeXMP.FindSub("testdata/test (2).jpg", hiddenPath, false)
		assert.Equal(t, "", result)
	})

	t.Run("find xmp with sequence", func(t *testing.T) {
		result := TypeXMP.FindSub("testdata/test (2).jpg", hiddenPath, true)
		assert.Equal(t, "testdata/.photoprism/test.xmp", result)
	})

	t.Run("find jpg", func(t *testing.T) {
		result := TypeJpeg.FindSub("testdata/test.xmp", hiddenPath, false)
		assert.Equal(t, "testdata/test.jpg", result)
	})

	t.Run("upper ext", func(t *testing.T) {
		result := TypeJpeg.FindSub("testdata/test.XMP", hiddenPath, false)
		assert.Equal(t, "testdata/test.jpg", result)
	})

	t.Run("with sequence", func(t *testing.T) {
		result := TypeJpeg.FindSub("testdata/test (2).xmp", hiddenPath, false)
		assert.Equal(t, "", result)
	})

	t.Run("strip sequence", func(t *testing.T) {
		result := TypeJpeg.FindSub("testdata/test (2).xmp", hiddenPath, true)
		assert.Equal(t, "testdata/test.jpg", result)
	})
	t.Run("name upper", func(t *testing.T) {
		result := TypeJpeg.FindSub("testdata/CATYELLOW.xmp", hiddenPath, true)
		assert.Equal(t, "testdata/CATYELLOW.jpg", result)
	})
	t.Run("name lower", func(t *testing.T) {
		result := TypeJpeg.FindSub("testdata/chameleon_lime.xmp", hiddenPath, true)
		assert.Equal(t, "testdata/chameleon_lime.jpg", result)
	})
}
