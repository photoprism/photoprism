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

func TestFileType_FindFirst(t *testing.T) {
	dirs := []string{HiddenPath}

	t.Run("find xmp", func(t *testing.T) {
		result := TypeXMP.FindFirst("testdata/test.jpg", dirs, "", false)
		assert.Equal(t, "testdata/.photoprism/test.xmp", result)
	})

	t.Run("find xmp upper ext", func(t *testing.T) {
		result := TypeXMP.FindFirst("testdata/test.PNG", dirs, "", false)
		assert.Equal(t, "testdata/.photoprism/test.xmp", result)
	})

	t.Run("find xmp without sequence", func(t *testing.T) {
		result := TypeXMP.FindFirst("testdata/test (2).jpg", dirs, "", false)
		assert.Equal(t, "", result)
	})

	t.Run("find xmp with sequence", func(t *testing.T) {
		result := TypeXMP.FindFirst("testdata/test (2).jpg", dirs, "", true)
		assert.Equal(t, "testdata/.photoprism/test.xmp", result)
	})

	t.Run("find jpg", func(t *testing.T) {
		result := TypeJpeg.FindFirst("testdata/test.xmp", dirs, "", false)
		assert.Equal(t, "testdata/test.jpg", result)
	})

	t.Run("find jpg abs", func(t *testing.T) {
		result := TypeJpeg.FindFirst(Abs("testdata/test.xmp"), dirs, "", false)
		assert.Equal(t, Abs("testdata/test.jpg"), result)
	})

	t.Run("upper ext", func(t *testing.T) {
		result := TypeJpeg.FindFirst("testdata/test.XMP", dirs, "", false)
		assert.Equal(t, "testdata/test.jpg", result)
	})

	t.Run("with sequence", func(t *testing.T) {
		result := TypeJpeg.FindFirst("testdata/test (2).xmp", dirs, "", false)
		assert.Equal(t, "", result)
	})

	t.Run("strip sequence", func(t *testing.T) {
		result := TypeJpeg.FindFirst("testdata/test (2).xmp", dirs, "", true)
		assert.Equal(t, "testdata/test.jpg", result)
	})

	t.Run("name upper", func(t *testing.T) {
		result := TypeJpeg.FindFirst("testdata/CATYELLOW.xmp", dirs, "", true)
		assert.Equal(t, "testdata/CATYELLOW.jpg", result)
	})

	t.Run("name lower", func(t *testing.T) {
		result := TypeJpeg.FindFirst("testdata/chameleon_lime.xmp", dirs, "", true)
		assert.Equal(t, "testdata/chameleon_lime.jpg", result)
	})

	t.Run("example_bmp_notfound", func(t *testing.T) {
		result := TypeBitmap.FindFirst("testdata/example.00001.jpg", dirs, "", true)
		assert.Equal(t, "", result)
	})

	t.Run("example_bmp_found", func(t *testing.T) {
		result := TypeBitmap.FindFirst("testdata/example.00001.jpg", []string{"directory"}, "", true)
		assert.Equal(t, "testdata/directory/example.bmp", result)
	})

	t.Run("example_png_found", func(t *testing.T) {
		result := TypePng.FindFirst("testdata/example.00001.jpg", []string{"directory", "directory/subdirectory"}, "", true)
		assert.Equal(t, "testdata/directory/subdirectory/example.png", result)
	})

	t.Run("example_bmp_found", func(t *testing.T) {
		result := TypeBitmap.FindFirst(Abs("testdata/example.00001.jpg"), []string{"directory"}, Abs("testdata"), true)
		assert.Equal(t, Abs("testdata/directory/example.bmp"), result)
	})
}
