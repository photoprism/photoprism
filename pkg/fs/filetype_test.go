package fs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFileType(t *testing.T) {
	t.Run("jpeg", func(t *testing.T) {
		result := GetFileFormat("testdata/test.jpg")
		assert.Equal(t, FormatJpeg, result)
	})

	t.Run("raw", func(t *testing.T) {
		result := GetFileFormat("testdata/test (jpg).CR2")
		assert.Equal(t, FormatRaw, result)
	})

	t.Run("empty", func(t *testing.T) {
		result := GetFileFormat("")
		assert.Equal(t, FormatOther, result)
	})
}

func TestFileType_Find(t *testing.T) {
	t.Run("find jpg", func(t *testing.T) {
		result := FormatJpeg.Find("testdata/test.xmp", false)
		assert.Equal(t, "testdata/test.jpg", result)
	})

	t.Run("upper ext", func(t *testing.T) {
		result := FormatJpeg.Find("testdata/test.XMP", false)
		assert.Equal(t, "testdata/test.jpg", result)
	})

	t.Run("with sequence", func(t *testing.T) {
		result := FormatJpeg.Find("testdata/test (2).xmp", false)
		assert.Equal(t, "", result)
	})

	t.Run("strip sequence", func(t *testing.T) {
		result := FormatJpeg.Find("testdata/test (2).xmp", true)
		assert.Equal(t, "testdata/test.jpg", result)
	})

	t.Run("name upper", func(t *testing.T) {
		result := FormatJpeg.Find("testdata/CATYELLOW.xmp", true)
		assert.Equal(t, "testdata/CATYELLOW.jpg", result)
	})

	t.Run("name lower", func(t *testing.T) {
		result := FormatJpeg.Find("testdata/chameleon_lime.xmp", true)
		assert.Equal(t, "testdata/chameleon_lime.jpg", result)
	})
}

func TestFileType_FindFirst(t *testing.T) {
	dirs := []string{HiddenPath}

	t.Run("find xmp", func(t *testing.T) {
		result := FormatXMP.FindFirst("testdata/test.jpg", dirs, "", false)
		assert.Equal(t, "testdata/.photoprism/test.xmp", result)
	})

	t.Run("find xmp upper ext", func(t *testing.T) {
		result := FormatXMP.FindFirst("testdata/test.PNG", dirs, "", false)
		assert.Equal(t, "testdata/.photoprism/test.xmp", result)
	})

	t.Run("find xmp without sequence", func(t *testing.T) {
		result := FormatXMP.FindFirst("testdata/test (2).jpg", dirs, "", false)
		assert.Equal(t, "", result)
	})

	t.Run("find xmp with sequence", func(t *testing.T) {
		result := FormatXMP.FindFirst("testdata/test (2).jpg", dirs, "", true)
		assert.Equal(t, "testdata/.photoprism/test.xmp", result)
	})

	t.Run("find jpg", func(t *testing.T) {
		result := FormatJpeg.FindFirst("testdata/test.xmp", dirs, "", false)
		assert.Equal(t, "testdata/test.jpg", result)
	})

	t.Run("find jpg abs", func(t *testing.T) {
		result := FormatJpeg.FindFirst(Abs("testdata/test.xmp"), dirs, "", false)
		assert.Equal(t, Abs("testdata/test.jpg"), result)
	})

	t.Run("upper ext", func(t *testing.T) {
		result := FormatJpeg.FindFirst("testdata/test.XMP", dirs, "", false)
		assert.Equal(t, "testdata/test.jpg", result)
	})

	t.Run("with sequence", func(t *testing.T) {
		result := FormatJpeg.FindFirst("testdata/test (2).xmp", dirs, "", false)
		assert.Equal(t, "", result)
	})

	t.Run("strip sequence", func(t *testing.T) {
		result := FormatJpeg.FindFirst("testdata/test (2).xmp", dirs, "", true)
		assert.Equal(t, "testdata/test.jpg", result)
	})

	t.Run("name upper", func(t *testing.T) {
		result := FormatJpeg.FindFirst("testdata/CATYELLOW.xmp", dirs, "", true)
		assert.Equal(t, "testdata/CATYELLOW.jpg", result)
	})

	t.Run("name lower", func(t *testing.T) {
		result := FormatJpeg.FindFirst("testdata/chameleon_lime.xmp", dirs, "", true)
		assert.Equal(t, "testdata/chameleon_lime.jpg", result)
	})

	t.Run("example_bmp_notfound", func(t *testing.T) {
		result := FormatBitmap.FindFirst("testdata/example.00001.jpg", dirs, "", true)
		assert.Equal(t, "", result)
	})

	t.Run("example_bmp_found", func(t *testing.T) {
		result := FormatBitmap.FindFirst("testdata/example.00001.jpg", []string{"directory"}, "", true)
		assert.Equal(t, "testdata/directory/example.bmp", result)
	})

	t.Run("example_png_found", func(t *testing.T) {
		result := FormatPng.FindFirst("testdata/example.00001.jpg", []string{"directory", "directory/subdirectory"}, "", true)
		assert.Equal(t, "testdata/directory/subdirectory/example.png", result)
	})

	t.Run("example_bmp_found", func(t *testing.T) {
		result := FormatBitmap.FindFirst(Abs("testdata/example.00001.jpg"), []string{"directory"}, Abs("testdata"), true)
		assert.Equal(t, Abs("testdata/directory/example.bmp"), result)
	})
}

func TestFileType_FindAll(t *testing.T) {
	dirs := []string{HiddenPath}

	t.Run("CATYELLOW.jpg", func(t *testing.T) {
		result := FormatJpeg.FindAll("testdata/CATYELLOW.JSON", dirs, "", false)
		assert.Contains(t, result, "testdata/CATYELLOW.jpg")
	})
}

func TestFileExt(t *testing.T) {
	t.Run("mp", func(t *testing.T) {
		assert.True(t, FileExt.Known("file.mp"))
	})
}

func TestGetFileFormat(t *testing.T) {
	t.Run("mp", func(t *testing.T) {
		assert.Equal(t, FileFormat("mp4"), GetFileFormat("file.mp"))
	})
}
