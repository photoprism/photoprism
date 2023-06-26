package clean

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHex(t *testing.T) {
	t.Run("UUID", func(t *testing.T) {
		assert.Equal(t, "123e4567e89b12d3a456426614174000", Hex("123e4567-e89b-12d3-A456-426614174000 "))
	})
	t.Run("ThumbSize", func(t *testing.T) {
		assert.Equal(t, "ef224", Hex("left_224"))
	})
	t.Run("SHA1", func(t *testing.T) {
		assert.Equal(t, "5c50ae14f339364eb8224f23c2d3abc7e79016f3eaded", Hex("5c50ae14f339364eb8224f23c2d3abc7e79016f3  README.md"))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", Hex(""))
	})
}
