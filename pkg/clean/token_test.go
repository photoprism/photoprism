package clean

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToken(t *testing.T) {
	t.Run("UUID", func(t *testing.T) {
		assert.Equal(t, "123e4567-e89b-12d3-A456-426614174000", Token("123e4567-e89b-12d3-A456-426614174000 "))
	})
	t.Run("ThumbSize", func(t *testing.T) {
		assert.Equal(t, "left_224", Token("left_224"))
	})
	t.Run("SHA1", func(t *testing.T) {
		assert.Equal(t, "5c50ae14f339364eb8224f23c2d3abc7e79016f3READMEmd", Token("5c50ae14f339364eb8224f23c2d3abc7e79016f3  README.md"))
	})
}
