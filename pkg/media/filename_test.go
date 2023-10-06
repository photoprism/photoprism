package media

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFromName(t *testing.T) {
	t.Run("jpeg", func(t *testing.T) {
		result := FromName("testdata/test.jpg")
		assert.Equal(t, Image, result)
	})
	t.Run("raw", func(t *testing.T) {
		result := FromName("testdata/test (jpg).CR2")
		assert.Equal(t, Raw, result)
	})
	t.Run("video", func(t *testing.T) {
		result := FromName("testdata/gopher.mp4")
		assert.Equal(t, Video, result)
	})
	t.Run("sidecar", func(t *testing.T) {
		result := FromName("/IMG_4120.AAE")
		assert.Equal(t, Sidecar, result)
	})
	t.Run("other", func(t *testing.T) {
		result := FromName("/IMG_4120.XXX")
		assert.Equal(t, Sidecar, result)
	})
	t.Run("empty", func(t *testing.T) {
		result := FromName("")
		assert.Equal(t, Unknown, result)
	})
}

func TestMainFile(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		assert.True(t, MainFile("testdata/test.jpg"))
	})
	t.Run("false", func(t *testing.T) {
		assert.False(t, MainFile("/IMG_4120.XXX"))
	})
}
