package clean

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIdString(t *testing.T) {
	t.Run("UUID", func(t *testing.T) {
		assert.Equal(t, "123e4567-e89b-12d3-a456-426614174000", IdString("123e4567-e89b-12d3-A456-426614174000 "))
	})
	t.Run("ThumbSize", func(t *testing.T) {
		assert.Equal(t, "left_224", IdString("left_224"))
	})
	t.Run("SHA1", func(t *testing.T) {
		assert.Equal(t, "5c50ae14f339364eb8224f23c2d3abc7e79016f3readmemd", IdString("5c50ae14f339364eb8224f23c2d3abc7e79016f3  README.md"))
	})
	t.Run("Quotes", func(t *testing.T) {
		assert.Equal(t, "foobaaar23", IdString("\"foo\" baa'ar 2```3"))
	})
}

func TestIdUint(t *testing.T) {
	t.Run("12334545", func(t *testing.T) {
		assert.Equal(t, uint(12334545), IdUint("12334545"))
	})
	t.Run("ThumbSize", func(t *testing.T) {
		assert.Equal(t, uint(0), IdUint("left_224"))
	})
	t.Run("SHA1", func(t *testing.T) {
		assert.Equal(t, uint(0), IdUint("5c50ae14f339364eb8224f23c2d3abc7e79016f3  README.md"))
	})
}
