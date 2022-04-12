package fs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetMediaType(t *testing.T) {
	t.Run("jpeg", func(t *testing.T) {
		result := GetMediaType("testdata/test.jpg")
		assert.Equal(t, MediaImage, result)
	})

	t.Run("raw", func(t *testing.T) {
		result := GetMediaType("testdata/test (jpg).CR2")
		assert.Equal(t, MediaRaw, result)
	})

	t.Run("video", func(t *testing.T) {
		result := GetMediaType("testdata/gopher.mp4")
		assert.Equal(t, MediaVideo, result)
	})

	t.Run("sidecar", func(t *testing.T) {
		result := GetMediaType("/IMG_4120.AAE")
		assert.Equal(t, MediaSidecar, result)
	})

	t.Run("empty", func(t *testing.T) {
		result := GetMediaType("")
		assert.Equal(t, MediaOther, result)
	})

	t.Run("invalid type", func(t *testing.T) {
		result := GetMediaType("/IMG_4120.XXX")
		assert.Equal(t, MediaOther, result)
	})
}

func TestIsMedia(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		assert.True(t, IsMedia("testdata/test.jpg"))
	})

	t.Run("false", func(t *testing.T) {
		assert.False(t, IsMedia("/IMG_4120.XXX"))
	})
}
