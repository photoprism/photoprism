package clean

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCodec(t *testing.T) {
	t.Run("UUID", func(t *testing.T) {
		assert.Equal(t, "123e4567e89b12d3A456426614174000", Codec("123e4567-e89b-12d3-A456-426614174000 "))
	})
	t.Run("left_224", func(t *testing.T) {
		assert.Equal(t, "left_224", Codec("left_224"))
	})
	t.Run("VP09", func(t *testing.T) {
		assert.Equal(t, "VP09", Codec("VP09"))
	})
	t.Run("v_vp9", func(t *testing.T) {
		assert.Equal(t, "v_vp9", Codec("v_vp9"))
	})
	t.Run("SHA1", func(t *testing.T) {
		assert.Equal(t, "5c50ae14f339364eb8224f23c2d3abc7e79016f3READMEmd", Codec("5c50ae14f339364eb8224f23c2d3abc7e79016f3  README.md"))
	})
	t.Run("Quotes", func(t *testing.T) {
		assert.Equal(t, "fooBaaar23", Codec("\"foo\" Baa'ar 2```3"))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", Codec(""))
	})
}
