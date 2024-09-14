package fs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileExtensions_Known(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.False(t, Extensions.Known(""))
	})
	t.Run("jpg", func(t *testing.T) {
		assert.True(t, Extensions.Known("testdata/test.jpg"))
	})
	t.Run("jpeg", func(t *testing.T) {
		assert.True(t, Extensions.Known("testdata/test.jpeg"))
	})
	t.Run("cr2", func(t *testing.T) {
		assert.True(t, Extensions.Known("testdata/.xxx/test (jpg).cr2"))
	})
	t.Run("CR2", func(t *testing.T) {
		assert.True(t, Extensions.Known("testdata/test (jpg).CR2"))
	})
	t.Run("CR5", func(t *testing.T) {
		assert.False(t, Extensions.Known("testdata/test (jpg).CR5"))
	})
	t.Run("mp", func(t *testing.T) {
		assert.True(t, Extensions.Known("file.mp"))
	})
	t.Run("mxf", func(t *testing.T) {
		assert.True(t, Extensions.Known("file.mxf"))
	})
}
