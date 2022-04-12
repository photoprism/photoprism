package fs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileFormat_String(t *testing.T) {
	t.Run("jpg", func(t *testing.T) {
		assert.Equal(t, "jpg", FormatJpeg.String())
	})
}

func TestFileFormat_Is(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.False(t, FormatJpeg.Is(""))
	})
	t.Run("Upper", func(t *testing.T) {
		assert.True(t, FormatJpeg.Is("JPG"))
	})
	t.Run("Lower", func(t *testing.T) {
		assert.True(t, FormatJpeg.Is("jpg"))
	})
	t.Run("False", func(t *testing.T) {
		assert.False(t, FormatJpeg.Is("raw"))
	})
}

func TestFileFormat_Find(t *testing.T) {
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

func TestFileFormat_FindFirst(t *testing.T) {
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

func TestFileFormat_FindAll(t *testing.T) {
	dirs := []string{HiddenPath}

	t.Run("CATYELLOW.jpg", func(t *testing.T) {
		result := FormatJpeg.FindAll("testdata/CATYELLOW.JSON", dirs, "", false)
		assert.Contains(t, result, "testdata/CATYELLOW.jpg")
	})
}

func TestFileFormat(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		result := FileFormat("")
		assert.Equal(t, FormatOther, result)
	})
	t.Run("JPEG", func(t *testing.T) {
		result := FileFormat("testdata/test.jpg")
		assert.Equal(t, FormatJpeg, result)
	})
	t.Run("RawCRw", func(t *testing.T) {
		result := FileFormat("testdata/test (jpg).crw")
		assert.Equal(t, FormatRaw, result)
	})
	t.Run("RawCR2", func(t *testing.T) {
		result := FileFormat("testdata/test (jpg).CR2")
		assert.Equal(t, FormatRaw, result)
	})
	t.Run("MP4", func(t *testing.T) {
		assert.Equal(t, Format("mp4"), FileFormat("file.mp"))
	})
}
