package fs

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/media/http/header"
)

func TestMimeType(t *testing.T) {
	t.Run("Mp4", func(t *testing.T) {
		filename := Abs("./testdata/test.mp4")
		mimeType := MimeType(filename)
		assert.Equal(t, "video/mp4", mimeType)
	})
	t.Run("MOV", func(t *testing.T) {
		filename := Abs("./testdata/test.mov")
		mimeType := MimeType(filename)
		assert.Equal(t, "video/quicktime", mimeType)
	})
	t.Run("JPEG", func(t *testing.T) {
		filename := Abs("./testdata/test.jpg")
		mimeType := MimeType(filename)
		assert.Equal(t, "image/jpeg", mimeType)
	})
	t.Run("InvalidFilename", func(t *testing.T) {
		filename := Abs("./testdata/xxx.jpg")
		mimeType := MimeType(filename)
		assert.Equal(t, "", mimeType)
	})
	t.Run("EmptyFilename", func(t *testing.T) {
		mimeType := MimeType("")
		assert.Equal(t, "", mimeType)
	})
	t.Run("AVIF", func(t *testing.T) {
		filename := Abs("./testdata/test.avif")
		mimeType := MimeType(filename)
		assert.Equal(t, "image/avif", mimeType)
	})
	t.Run("AVIFS", func(t *testing.T) {
		filename := Abs("./testdata/test.avifs")
		mimeType := MimeType(filename)
		assert.Equal(t, "image/avif-sequence", mimeType)
	})
	t.Run("HEIC", func(t *testing.T) {
		filename := Abs("./testdata/test.heic")
		mimeType := MimeType(filename)
		assert.Equal(t, "image/heic", mimeType)
	})
	t.Run("HEICS", func(t *testing.T) {
		filename := Abs("./testdata/test.heics")
		mimeType := MimeType(filename)
		assert.Equal(t, "image/heic-sequence", mimeType)
	})
	t.Run("DNG", func(t *testing.T) {
		filename := Abs("./testdata/test.dng")
		mimeType := MimeType(filename)
		assert.Equal(t, "image/dng", mimeType)
	})
	t.Run("SVG", func(t *testing.T) {
		filename := Abs("./testdata/test.svg")
		mimeType := MimeType(filename)
		assert.Equal(t, "image/svg+xml", mimeType)
	})
	t.Run("AI", func(t *testing.T) {
		filename := Abs("./testdata/test.ai")
		mimeType := MimeType(filename)
		assert.Equal(t, "application/vnd.adobe.illustrator", mimeType)
	})
	t.Run("PS", func(t *testing.T) {
		filename := Abs("./testdata/test.ps")
		mimeType := MimeType(filename)
		assert.Equal(t, "application/postscript", mimeType)
	})
	t.Run("EPS", func(t *testing.T) {
		filename := Abs("./testdata/test.eps")
		mimeType := MimeType(filename)
		assert.Equal(t, "image/eps", mimeType)
	})
}

func TestBaseType(t *testing.T) {
	t.Run("Mp4", func(t *testing.T) {
		filename := Abs("./testdata/test.mp4")
		mimeType := BaseType(MimeType(filename))
		assert.Equal(t, "video/mp4", mimeType)
	})
	t.Run("MOV", func(t *testing.T) {
		filename := Abs("./testdata/test.mov")
		mimeType := BaseType(MimeType(filename))
		assert.Equal(t, "video/quicktime", mimeType)
	})
	t.Run("JPEG", func(t *testing.T) {
		filename := Abs("./testdata/test.jpg")
		mimeType := BaseType(MimeType(filename))
		assert.Equal(t, "image/jpeg", mimeType)
	})
	t.Run("InvalidFilename", func(t *testing.T) {
		filename := Abs("./testdata/xxx.jpg")
		mimeType := BaseType(MimeType(filename))
		assert.Equal(t, "", mimeType)
	})
	t.Run("EmptyFilename", func(t *testing.T) {
		mimeType := BaseType("")
		assert.Equal(t, "", mimeType)
	})
	t.Run("AVIF", func(t *testing.T) {
		filename := Abs("./testdata/test.avif")
		mimeType := BaseType(MimeType(filename))
		assert.Equal(t, "image/avif", mimeType)
	})
	t.Run("AVIFS", func(t *testing.T) {
		filename := Abs("./testdata/test.avifs")
		mimeType := MimeType(filename)
		assert.Equal(t, "image/avif-sequence", mimeType)
	})
	t.Run("HEIC", func(t *testing.T) {
		filename := Abs("./testdata/test.heic")
		mimeType := BaseType(MimeType(filename))
		assert.Equal(t, "image/heic", mimeType)
	})
	t.Run("HEICS", func(t *testing.T) {
		filename := Abs("./testdata/test.heics")
		mimeType := BaseType(MimeType(filename))
		assert.Equal(t, "image/heic-sequence", mimeType)
	})
	t.Run("DNG", func(t *testing.T) {
		filename := Abs("./testdata/test.dng")
		mimeType := BaseType(MimeType(filename))
		assert.Equal(t, "image/dng", mimeType)
	})
	t.Run("SVG", func(t *testing.T) {
		filename := Abs("./testdata/test.svg")
		mimeType := BaseType(MimeType(filename))
		assert.Equal(t, "image/svg+xml", mimeType)
	})
	t.Run("AI", func(t *testing.T) {
		filename := Abs("./testdata/test.ai")
		mimeType := BaseType(MimeType(filename))
		assert.Equal(t, "application/vnd.adobe.illustrator", mimeType)
	})
	t.Run("PS", func(t *testing.T) {
		filename := Abs("./testdata/test.ps")
		mimeType := BaseType(MimeType(filename))
		assert.Equal(t, "application/postscript", mimeType)
	})
	t.Run("EPS", func(t *testing.T) {
		filename := Abs("./testdata/test.eps")
		mimeType := BaseType(MimeType(filename))
		assert.Equal(t, "image/eps", mimeType)
	})
}

func TestIsType(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		assert.True(t, SameType("", MimeTypeUnknown))
		assert.True(t, SameType("video/jpg", "video/jpg"))
		assert.True(t, SameType("video/jpeg", "video/jpeg"))
		assert.True(t, SameType("video/mp4", "video/mp4"))
		assert.True(t, SameType("video/mp4", header.ContentTypeMp4))
		assert.True(t, SameType("video/mp4", "video/Mp4"))
		assert.True(t, SameType("video/mp4", "video/Mp4; codecs=\"avc1.640028\""))
	})
	t.Run("False", func(t *testing.T) {
		assert.False(t, SameType("", header.ContentTypeMp4))
		assert.False(t, SameType("video/jpeg", "video/jpg"))
		assert.False(t, SameType("video/mp4", MimeTypeUnknown))
		assert.False(t, SameType(header.ContentTypeMp4, header.ContentTypeJpeg))
	})
}
