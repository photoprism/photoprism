package fs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeExt(t *testing.T) {
	t.Run("none", func(t *testing.T) {
		result := NormalizeExt("testdata/test")
		assert.Equal(t, "", result)
	})

	t.Run("dot", func(t *testing.T) {
		result := NormalizeExt("testdata/test.")
		assert.Equal(t, "", result)
	})

	t.Run("test.z", func(t *testing.T) {
		result := NormalizeExt("testdata/test.z")
		assert.Equal(t, "z", result)
	})

	t.Run("test.jpg", func(t *testing.T) {
		result := NormalizeExt("testdata/test.jpg")
		assert.Equal(t, "jpg", result)
	})

	t.Run("test.PNG", func(t *testing.T) {
		result := NormalizeExt("testdata/test.PNG")
		assert.Equal(t, "png", result)
	})

	t.Run("test.MOV", func(t *testing.T) {
		result := NormalizeExt("testdata/test.MOV")
		assert.Equal(t, "mov", result)
	})

	t.Run("test.xmp", func(t *testing.T) {
		result := NormalizeExt("testdata/test.xMp")
		assert.Equal(t, "xmp", result)
	})

	t.Run("test.MP", func(t *testing.T) {
		result := NormalizeExt("testdata/test.mp")
		assert.Equal(t, "mp", result)
	})
}

func TestTrimExt(t *testing.T) {
	t.Run("WithDot", func(t *testing.T) {
		assert.Equal(t, "raf", TrimExt(".raf"))
	})
	t.Run("Normalized", func(t *testing.T) {
		assert.Equal(t, "cr3", TrimExt("cr3"))
	})
	t.Run("Uppercase", func(t *testing.T) {
		assert.Equal(t, "aaf", TrimExt("AAF"))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", TrimExt(""))
	})
	t.Run("MixedCaseWithDot", func(t *testing.T) {
		assert.Equal(t, "raw", TrimExt(".Raw"))
	})
	t.Run("TypographicQuotes", func(t *testing.T) {
		assert.Equal(t, "jpeg", TrimExt(" “JPEG” "))
	})
}
