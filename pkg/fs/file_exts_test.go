package fs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileExtensions_Known(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.False(t, Extensions.Known(""))
	})
	t.Run("Jpg", func(t *testing.T) {
		assert.True(t, Extensions.Known("testdata/test.jpg"))
	})
	t.Run("Jpeg", func(t *testing.T) {
		assert.True(t, Extensions.Known("testdata/test.jpeg"))
	})
	t.Run("Cr2", func(t *testing.T) {
		assert.True(t, Extensions.Known("testdata/.xxx/test (jpg).cr2"))
	})
	t.Run("CR2", func(t *testing.T) {
		assert.True(t, Extensions.Known("testdata/test (jpg).CR2"))
	})
	t.Run("CR5", func(t *testing.T) {
		assert.False(t, Extensions.Known("testdata/test (jpg).CR5"))
	})
	t.Run("Mp4", func(t *testing.T) {
		assert.True(t, Extensions.Known("file.mp4"))
	})
	t.Run("Mxf", func(t *testing.T) {
		assert.True(t, Extensions.Known("file.mxf"))
	})
}
