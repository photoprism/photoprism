package fs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFileType(t *testing.T) {
	t.Run("jpeg", func(t *testing.T) {
		result := GetFileType("testdata/test.jpg")
		assert.Equal(t, TypeJpeg, result)
	})

	t.Run("raw", func(t *testing.T) {
		result := GetFileType("testdata/test (jpg).CR2")
		assert.Equal(t, TypeRaw, result)
	})

	t.Run("empty", func(t *testing.T) {
		result := GetFileType("")
		assert.Equal(t, TypeOther, result)
	})
}
