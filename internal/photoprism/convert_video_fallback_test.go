package photoprism

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/ffmpeg/encode"
)

func TestFirstTranscodeFallback(t *testing.T) {
	resetTranscodeFallbacks()
	t.Cleanup(resetTranscodeFallbacks)

	t.Run("Success", func(t *testing.T) {
		assert.True(t, firstTranscodeFallback(encode.NvidiaAvc))
		assert.False(t, firstTranscodeFallback(encode.NvidiaAvc))
		assert.False(t, firstTranscodeFallback(encode.NvidiaAvc))
	})
	t.Run("OtherEncoder", func(t *testing.T) {
		firstTranscodeFallback(encode.NvidiaAvc)
		assert.True(t, firstTranscodeFallback(encode.IntelAvc))
		assert.False(t, firstTranscodeFallback(encode.IntelAvc))
		assert.False(t, firstTranscodeFallback(encode.NvidiaAvc))
	})
}

func TestClearTranscodeFallback(t *testing.T) {
	resetTranscodeFallbacks()
	t.Cleanup(resetTranscodeFallbacks)

	t.Run("Success", func(t *testing.T) {
		firstTranscodeFallback(encode.NvidiaAvc)
		firstTranscodeFallback(encode.IntelAvc)
		clearTranscodeFallback(encode.NvidiaAvc)
		assert.True(t, firstTranscodeFallback(encode.NvidiaAvc))
		assert.False(t, firstTranscodeFallback(encode.IntelAvc))
	})
	t.Run("NotRecorded", func(t *testing.T) {
		clearTranscodeFallback(encode.VulkanAvc)
		assert.True(t, firstTranscodeFallback(encode.VulkanAvc))
	})
}

func TestResetTranscodeFallbacks(t *testing.T) {
	t.Cleanup(resetTranscodeFallbacks)

	t.Run("Success", func(t *testing.T) {
		firstTranscodeFallback(encode.VaapiAvc)
		assert.False(t, firstTranscodeFallback(encode.VaapiAvc))
		resetTranscodeFallbacks()
		assert.True(t, firstTranscodeFallback(encode.VaapiAvc))
	})
}
