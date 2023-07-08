package fs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMimeType(t *testing.T) {
	t.Run("jpg", func(t *testing.T) {
		filename := Abs("./testdata/test.jpg")
		mimeType := MimeType(filename)
		assert.Equal(t, "image/jpeg", mimeType)
	})
	t.Run("not existing filename", func(t *testing.T) {
		filename := Abs("./testdata/xxx.jpg")
		mimeType := MimeType(filename)
		assert.Equal(t, "", mimeType)
	})
	t.Run("Empty", func(t *testing.T) {
		mimeType := MimeType("")
		assert.Equal(t, "", mimeType)
	})
	t.Run("avif", func(t *testing.T) {
		filename := Abs("./testdata/test.avif")
		mimeType := MimeType(filename)
		assert.Equal(t, "image/avif", mimeType)
	})
	t.Run("DNG", func(t *testing.T) {
		filename := Abs("./testdata/test.dng")
		mimeType := MimeType(filename)
		assert.Equal(t, "image/dng", mimeType)
	})
	t.Run("avifs", func(t *testing.T) {
		filename := Abs("./testdata/test.avifs")
		mimeType := MimeType(filename)
		assert.Equal(t, "image/avif-sequence", mimeType)
	})
	t.Run("Heic", func(t *testing.T) {
		filename := Abs("./testdata/test.heic")
		mimeType := MimeType(filename)
		assert.Equal(t, "image/heic", mimeType)
	})
	t.Run("Heics", func(t *testing.T) {
		filename := Abs("./testdata/test.heics")
		mimeType := MimeType(filename)
		assert.Equal(t, "image/heic-sequence", mimeType)
	})
	t.Run("Mp4", func(t *testing.T) {
		filename := Abs("./testdata/test.mp4")
		mimeType := MimeType(filename)
		assert.Equal(t, "video/mp4", mimeType)
	})
	t.Run("Mov", func(t *testing.T) {
		filename := Abs("./testdata/test.mov")
		mimeType := MimeType(filename)
		assert.Equal(t, "video/quicktime", mimeType)
	})
	t.Run("Svg", func(t *testing.T) {
		filename := Abs("./testdata/test.svg")
		mimeType := MimeType(filename)
		assert.Equal(t, "image/svg+xml", mimeType)
	})
	t.Run("Ai", func(t *testing.T) {
		filename := Abs("./testdata/test.ai")
		mimeType := MimeType(filename)
		assert.Equal(t, "application/vnd.adobe.illustrator", mimeType)
	})
	t.Run("ps", func(t *testing.T) {
		filename := Abs("./testdata/test.ps")
		mimeType := MimeType(filename)
		assert.Equal(t, "application/ps", mimeType)
	})
	t.Run("eps", func(t *testing.T) {
		filename := Abs("./testdata/test.eps")
		mimeType := MimeType(filename)
		assert.Equal(t, "image/eps", mimeType)
	})
}
