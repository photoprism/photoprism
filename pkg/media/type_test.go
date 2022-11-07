package media

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestType_Main(t *testing.T) {
	t.Run("Unknown", func(t *testing.T) {
		assert.False(t, Unknown.Main())
	})
	t.Run("Image", func(t *testing.T) {
		assert.True(t, Image.Main())
	})
	t.Run("Video", func(t *testing.T) {
		assert.True(t, Video.Main())
	})
	t.Run("Sidecar", func(t *testing.T) {
		assert.False(t, Sidecar.Main())
	})
}

func TestType_Unknown(t *testing.T) {
	t.Run("Unknown", func(t *testing.T) {
		assert.True(t, Unknown.Unknown())
	})
	t.Run("Image", func(t *testing.T) {
		assert.False(t, Image.Unknown())
	})
	t.Run("Video", func(t *testing.T) {
		assert.False(t, Video.Unknown())
	})
	t.Run("Sidecar", func(t *testing.T) {
		assert.False(t, Sidecar.Unknown())
	})
}

func TestType_Equal(t *testing.T) {
	t.Run("UnknownUnknown", func(t *testing.T) {
		assert.True(t, Unknown.Equal(""))
	})
	t.Run("ImageImage", func(t *testing.T) {
		assert.True(t, Image.Equal(Image.String()))
	})
	t.Run("VideoImage", func(t *testing.T) {
		assert.False(t, Video.Equal(Image.String()))
	})
	t.Run("SidecarUnknown", func(t *testing.T) {
		assert.False(t, Sidecar.Equal(Unknown.String()))
	})
}

func TestType_NotEqual(t *testing.T) {
	t.Run("UnknownUnknown", func(t *testing.T) {
		assert.False(t, Unknown.NotEqual(""))
	})
	t.Run("ImageImage", func(t *testing.T) {
		assert.False(t, Image.NotEqual(Image.String()))
	})
	t.Run("VideoImage", func(t *testing.T) {
		assert.True(t, Video.NotEqual(Image.String()))
	})
	t.Run("SidecarUnknown", func(t *testing.T) {
		assert.True(t, Sidecar.NotEqual(Unknown.String()))
	})
}
