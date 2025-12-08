package media

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFromName(t *testing.T) {
	t.Run("Jpeg", func(t *testing.T) {
		result := FromName("testdata/test.jpg")
		assert.Equal(t, Image, result)
	})
	t.Run("Raw", func(t *testing.T) {
		result := FromName("testdata/test (jpg).CR2")
		assert.Equal(t, Raw, result)
	})
	t.Run("Video", func(t *testing.T) {
		result := FromName("testdata/gopher.mp4")
		assert.Equal(t, Video, result)
	})
	t.Run("Sidecar", func(t *testing.T) {
		result := FromName("/IMG_4120.AAE")
		assert.Equal(t, Sidecar, result)
	})
	t.Run("Other", func(t *testing.T) {
		result := FromName("/IMG_4120.XXX")
		assert.Equal(t, Sidecar, result)
	})
	t.Run("Empty", func(t *testing.T) {
		result := FromName("")
		assert.Equal(t, Unknown, result)
	})
}

func TestMainFile(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		assert.True(t, MainFile("testdata/test.jpg"))
	})
	t.Run("False", func(t *testing.T) {
		assert.False(t, MainFile("/IMG_4120.XXX"))
	})
}
