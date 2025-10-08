package fs

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/service/http/header"
)

func TestDetectMimeType(t *testing.T) {
	t.Run("MP4", func(t *testing.T) {
		filename := Abs("./testdata/test.mp4")
		mimeType, _ := DetectMimeType(filename)
		assert.Equal(t, "video/mp4", mimeType)
	})
	t.Run("MOV", func(t *testing.T) {
		filename := Abs("./testdata/test.mov")
		mimeType, _ := DetectMimeType(filename)
		assert.Equal(t, "video/quicktime", mimeType)
		assert.Equal(t, "video/quicktime", MimeType(filename))
	})
	t.Run("JPEG", func(t *testing.T) {
		filename := Abs("./testdata/test.jpg")
		mimeType, _ := DetectMimeType(filename)
		assert.Equal(t, "image/jpeg", mimeType)
		assert.Equal(t, "image/jpeg", MimeType(filename))
	})
	t.Run("InvalidFilename", func(t *testing.T) {
		filename := Abs("./testdata/xxx.jpg")
		mimeType, _ := DetectMimeType(filename)
		assert.Equal(t, "", mimeType)
		assert.Equal(t, "", MimeType(filename))
	})
	t.Run("EmptyFilename", func(t *testing.T) {
		mimeType, _ := DetectMimeType("")
		assert.Equal(t, "", mimeType)
		assert.Equal(t, "", MimeType(""))
	})
	t.Run("AVIF", func(t *testing.T) {
		filename := Abs("./testdata/test.avif")
		mimeType, _ := DetectMimeType(filename)
		assert.Equal(t, "image/avif", mimeType)
		assert.Equal(t, "image/avif", MimeType(filename))
	})
	t.Run("AVIFS", func(t *testing.T) {
		filename := Abs("./testdata/test.avifs")
		mimeType, _ := DetectMimeType(filename)
		assert.Equal(t, "image/avif-sequence", mimeType)
		assert.Equal(t, "image/avif-sequence", MimeType(filename))
	})
	t.Run("HEIC", func(t *testing.T) {
		filename := Abs("./testdata/test.heic")
		mimeType, _ := DetectMimeType(filename)
		assert.Equal(t, "image/heic", mimeType)
		assert.Equal(t, "image/heic", MimeType(filename))
	})
	t.Run("HEICS", func(t *testing.T) {
		filename := Abs("./testdata/test.heics")
		mimeType, _ := DetectMimeType(filename)
		assert.Equal(t, "image/heic-sequence", mimeType)
	})
	t.Run("DNG", func(t *testing.T) {
		filename := Abs("./testdata/test.dng")
		mimeType, _ := DetectMimeType(filename)
		assert.Equal(t, "image/dng", mimeType)
	})
	t.Run("SVG", func(t *testing.T) {
		filename := Abs("./testdata/test.svg")
		mimeType, _ := DetectMimeType(filename)
		assert.Equal(t, "image/svg+xml", mimeType)
		assert.Equal(t, "image/svg+xml", MimeType(filename))
	})
	t.Run("AI", func(t *testing.T) {
		filename := Abs("./testdata/test.ai")
		mimeType, _ := DetectMimeType(filename)
		assert.Equal(t, "application/vnd.adobe.illustrator", mimeType)
		assert.Equal(t, "application/vnd.adobe.illustrator", MimeType(filename))
	})
	t.Run("PS", func(t *testing.T) {
		filename := Abs("./testdata/test.ps")
		mimeType, _ := DetectMimeType(filename)
		assert.Equal(t, "application/postscript", mimeType)
		assert.Equal(t, "application/postscript", MimeType(filename))
	})
	t.Run("EPS", func(t *testing.T) {
		filename := Abs("./testdata/test.eps")
		mimeType, _ := DetectMimeType(filename)
		assert.Equal(t, "image/eps", mimeType)
		assert.Equal(t, "image/eps", MimeType(filename))
	})
}

func TestBaseType(t *testing.T) {
	t.Run("MP4", func(t *testing.T) {
		filename := Abs("./testdata/test.mp4")
		result := BaseType(MimeType(filename))
		assert.Equal(t, "video/mp4", result)
	})
	t.Run("MOV", func(t *testing.T) {
		filename := Abs("./testdata/test.mov")
		result := BaseType(MimeType(filename))
		assert.Equal(t, "video/quicktime", result)
	})
	t.Run("JPEG", func(t *testing.T) {
		filename := Abs("./testdata/test.jpg")
		result := BaseType(MimeType(filename))
		assert.Equal(t, "image/jpeg", result)
	})
	t.Run("InvalidFilename", func(t *testing.T) {
		filename := Abs("./testdata/xxx.jpg")
		result := BaseType(MimeType(filename))
		assert.Equal(t, "", result)
	})
	t.Run("EmptyFilename", func(t *testing.T) {
		assert.Equal(t, "", BaseType(""))
	})
	t.Run("AVIF", func(t *testing.T) {
		filename := Abs("./testdata/test.avif")
		result := BaseType(MimeType(filename))
		assert.Equal(t, "image/avif", result)
	})
	t.Run("AVIFS", func(t *testing.T) {
		filename := Abs("./testdata/test.avifs")
		mimeType, _ := DetectMimeType(filename)
		assert.Equal(t, "image/avif-sequence", mimeType)
		assert.Equal(t, "image/avif-sequence", BaseType(mimeType))
	})
	t.Run("HEIC", func(t *testing.T) {
		filename := Abs("./testdata/test.heic")
		result := BaseType(MimeType(filename))
		assert.Equal(t, "image/heic", result)
	})
	t.Run("HEICS", func(t *testing.T) {
		filename := Abs("./testdata/test.heics")
		result := BaseType(MimeType(filename))
		assert.Equal(t, "image/heic-sequence", result)
	})
	t.Run("DNG", func(t *testing.T) {
		filename := Abs("./testdata/test.dng")
		result := BaseType(MimeType(filename))
		assert.Equal(t, "image/dng", result)
	})
	t.Run("SVG", func(t *testing.T) {
		filename := Abs("./testdata/test.svg")
		result := BaseType(MimeType(filename))
		assert.Equal(t, "image/svg+xml", result)
	})
	t.Run("AI", func(t *testing.T) {
		filename := Abs("./testdata/test.ai")
		result := BaseType(MimeType(filename))
		assert.Equal(t, "application/vnd.adobe.illustrator", result)
	})
	t.Run("PS", func(t *testing.T) {
		filename := Abs("./testdata/test.ps")
		result := BaseType(MimeType(filename))
		assert.Equal(t, "application/postscript", result)
	})
	t.Run("EPS", func(t *testing.T) {
		filename := Abs("./testdata/test.eps")
		result := BaseType(MimeType(filename))
		assert.Equal(t, "image/eps", result)
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
