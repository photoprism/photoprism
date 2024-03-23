package fs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMimeType(t *testing.T) {
	t.Run("MP4", func(t *testing.T) {
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
