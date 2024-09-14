package fs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestType_String(t *testing.T) {
	t.Run("jpg", func(t *testing.T) {
		assert.Equal(t, "jpg", ImageJPEG.String())
	})
}

func TestType_Equal(t *testing.T) {
	t.Run("jpg", func(t *testing.T) {
		assert.True(t, ImageJPEG.Equal("jpg"))
	})
}

func TestType_NotEqual(t *testing.T) {
	t.Run("jpg", func(t *testing.T) {
		assert.False(t, ImageJPEG.NotEqual("JPG"))
		assert.True(t, ImageJPEG.NotEqual("xmp"))
	})
}

func TestType_DefaultExt(t *testing.T) {
	t.Run("jpg", func(t *testing.T) {
		assert.Equal(t, ".jpg", ImageJPEG.DefaultExt())
	})
	t.Run("avif", func(t *testing.T) {
		assert.Equal(t, ".avif", ImageAVIF.DefaultExt())
	})
}

func TestToType(t *testing.T) {
	t.Run("jpg", func(t *testing.T) {
		assert.Equal(t, "jpg", NewType("JPG").String())
	})
	t.Run("JPEG", func(t *testing.T) {
		assert.Equal(t, Type("jpeg"), NewType("JPEG"))
	})
	t.Run(".jpg", func(t *testing.T) {
		assert.Equal(t, "jpg", NewType(".jpg").String())
	})
}

func TestType_Is(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.False(t, ImageJPEG.Equal(""))
	})
	t.Run("Upper", func(t *testing.T) {
		assert.True(t, ImageJPEG.Equal("JPG"))
	})
	t.Run("Lower", func(t *testing.T) {
		assert.True(t, ImageJPEG.Equal("jpg"))
	})
	t.Run("False", func(t *testing.T) {
		assert.False(t, ImageJPEG.Equal("raw"))
	})
}

func TestType_Find(t *testing.T) {
	t.Run("find jpg", func(t *testing.T) {
		result := ImageJPEG.Find("testdata/test.xmp", false)
		assert.Equal(t, "testdata/test.jpg", result)
	})
	t.Run("upper ext", func(t *testing.T) {
		result := ImageJPEG.Find("testdata/test.XMP", false)
		assert.Equal(t, "testdata/test.jpg", result)
	})
	t.Run("with sequence", func(t *testing.T) {
		result := ImageJPEG.Find("testdata/test (2).xmp", false)
		assert.Equal(t, "", result)
	})
	t.Run("strip sequence", func(t *testing.T) {
		result := ImageJPEG.Find("testdata/test (2).xmp", true)
		assert.Equal(t, "testdata/test.jpg", result)
	})

	t.Run("name upper", func(t *testing.T) {
		result := ImageJPEG.Find("testdata/CATYELLOW.xmp", true)
		assert.Equal(t, "testdata/CATYELLOW.jpg", result)
	})

	t.Run("name lower", func(t *testing.T) {
		result := ImageJPEG.Find("testdata/chameleon_lime.xmp", true)
		assert.Equal(t, "testdata/chameleon_lime.jpg", result)
	})
}

func TestType_FindFirst(t *testing.T) {
	dirs := []string{PPHiddenPathname}

	t.Run("find xmp", func(t *testing.T) {
		result := SidecarXMP.FindFirst("testdata/test.jpg", dirs, "", false)
		assert.Equal(t, "testdata/.photoprism/test.xmp", result)
	})
	t.Run("find xmp upper ext", func(t *testing.T) {
		result := SidecarXMP.FindFirst("testdata/test.PNG", dirs, "", false)
		assert.Equal(t, "testdata/.photoprism/test.xmp", result)
	})
	t.Run("find xmp without sequence", func(t *testing.T) {
		result := SidecarXMP.FindFirst("testdata/test (2).jpg", dirs, "", false)
		assert.Equal(t, "", result)
	})
	t.Run("find xmp with sequence", func(t *testing.T) {
		result := SidecarXMP.FindFirst("testdata/test (2).jpg", dirs, "", true)
		assert.Equal(t, "testdata/.photoprism/test.xmp", result)
	})
	t.Run("find jpg", func(t *testing.T) {
		result := ImageJPEG.FindFirst("testdata/test.xmp", dirs, "", false)
		assert.Equal(t, "testdata/test.jpg", result)
	})
	t.Run("find jpg abs", func(t *testing.T) {
		result := ImageJPEG.FindFirst(Abs("testdata/test.xmp"), dirs, "", false)
		assert.Equal(t, Abs("testdata/test.jpg"), result)
	})
	t.Run("upper ext", func(t *testing.T) {
		result := ImageJPEG.FindFirst("testdata/test.XMP", dirs, "", false)
		assert.Equal(t, "testdata/test.jpg", result)
	})
	t.Run("with sequence", func(t *testing.T) {
		result := ImageJPEG.FindFirst("testdata/test (2).xmp", dirs, "", false)
		assert.Equal(t, "", result)
	})
	t.Run("strip sequence", func(t *testing.T) {
		result := ImageJPEG.FindFirst("testdata/test (2).xmp", dirs, "", true)
		assert.Equal(t, "testdata/test.jpg", result)
	})
	t.Run("name upper", func(t *testing.T) {
		result := ImageJPEG.FindFirst("testdata/CATYELLOW.xmp", dirs, "", true)
		assert.Equal(t, "testdata/CATYELLOW.jpg", result)
	})
	t.Run("name lower", func(t *testing.T) {
		result := ImageJPEG.FindFirst("testdata/chameleon_lime.xmp", dirs, "", true)
		assert.Equal(t, "testdata/chameleon_lime.jpg", result)
	})
	t.Run("example_bmp_notfound", func(t *testing.T) {
		result := ImageBMP.FindFirst("testdata/example.00001.jpg", dirs, "", true)
		assert.Equal(t, "", result)
	})
	t.Run("example_bmp_found", func(t *testing.T) {
		result := ImageBMP.FindFirst("testdata/example.00001.jpg", []string{"directory"}, "", true)
		assert.Equal(t, "testdata/directory/example.bmp", result)
	})
	t.Run("example_png_found", func(t *testing.T) {
		result := ImagePNG.FindFirst("testdata/example.00001.jpg", []string{"directory", "directory/subdirectory"}, "", true)
		assert.Equal(t, "testdata/directory/subdirectory/example.png", result)
	})
	t.Run("example_bmp_found", func(t *testing.T) {
		result := ImageBMP.FindFirst(Abs("testdata/example.00001.jpg"), []string{"directory"}, Abs("testdata"), true)
		assert.Equal(t, Abs("testdata/directory/example.bmp"), result)
	})
}

func TestType_FindAll(t *testing.T) {
	dirs := []string{PPHiddenPathname}

	t.Run("CATYELLOW.jpg", func(t *testing.T) {
		result := ImageJPEG.FindAll("testdata/CATYELLOW.JSON", dirs, "", false)
		assert.Contains(t, result, "testdata/CATYELLOW.jpg")
	})
}

func TestFileType(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		result := FileType("")
		assert.Equal(t, TypeUnknown, result)
	})
	t.Run("JPEG", func(t *testing.T) {
		result := FileType("testdata/test.jpg")
		assert.Equal(t, ImageJPEG, result)
	})
	t.Run("RawCRW", func(t *testing.T) {
		result := FileType("testdata/test (jpg).crw")
		assert.Equal(t, ImageRaw, result)
	})
	t.Run("RawCR2", func(t *testing.T) {
		result := FileType("testdata/test (jpg).CR2")
		assert.Equal(t, ImageRaw, result)
	})
	t.Run("MP4", func(t *testing.T) {
		assert.Equal(t, Type("mp4"), FileType("file.mp"))
	})
}

func TestIsAnimatedImage(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.False(t, IsAnimatedImage(""))
	})
	t.Run("JPEG", func(t *testing.T) {
		assert.False(t, IsAnimatedImage("testdata/test.jpg"))
	})
	t.Run("RawCRW", func(t *testing.T) {
		assert.False(t, IsAnimatedImage("testdata/test (jpg).crw"))
	})
	t.Run("MP4", func(t *testing.T) {
		assert.False(t, IsAnimatedImage("file.mp"))
		assert.False(t, IsAnimatedImage("file.mp4"))
	})
	t.Run("GIF", func(t *testing.T) {
		assert.True(t, IsAnimatedImage("file.gif"))
	})
	t.Run("WebP", func(t *testing.T) {
		assert.True(t, IsAnimatedImage("file.webp"))
	})
	t.Run("PNG", func(t *testing.T) {
		assert.True(t, IsAnimatedImage("file.png"))
		assert.True(t, IsAnimatedImage("file.apng"))
		assert.True(t, IsAnimatedImage("file.pnga"))
	})
	t.Run("AVIF", func(t *testing.T) {
		assert.True(t, IsAnimatedImage("file.avif"))
		assert.True(t, IsAnimatedImage("file.avis"))
		assert.True(t, IsAnimatedImage("file.avifs"))
	})
	t.Run("HEIC", func(t *testing.T) {
		assert.True(t, IsAnimatedImage("file.heic"))
		assert.True(t, IsAnimatedImage("file.heics"))
	})
}
