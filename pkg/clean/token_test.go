package clean

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToken(t *testing.T) {
	t.Run("ThumbSize", func(t *testing.T) {
		assert.Equal(t, "left_224", Token("left_224"))
	})
	t.Run("UUID", func(t *testing.T) {
		assert.Equal(t, "123e4567-e89b-12d3-A456-426614174000", Token("123e4567-e89b-12d3-A456-426614174000 "))
	})
	t.Run("SHA1", func(t *testing.T) {
		assert.Equal(t, "5c50ae14f339364eb8224f23c2d3abc7e79016f3READMEmd", Token("5c50ae14f339364eb8224f23c2d3abc7e79016f3  README.md"))
	})
	t.Run("SHA256", func(t *testing.T) {
		assert.Equal(t, "a948904f2f0f479b8f8197694b30184b0d2ed1c1cd2a1ec0fb85d299a192a447", Token("a948904f2f0f479b8f8197694b30184b0d2ed1c1cd2a1ec0fb85d299a192a447"))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", Token(""))
	})
}

func TestUrlToken(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		assert.Equal(t, "4f339364eb8224f23c2d3a", UrlToken("4f339364eb8224f23c2d3a"))
	})
	t.Run("Trim", func(t *testing.T) {
		assert.Equal(t, "123e4567-e89b-12d3-A456-426614174000", UrlToken("123e4567-e89b-12d3-A456-426614174000 "))
	})
	t.Run("SHA1", func(t *testing.T) {
		assert.Equal(t, "5c50ae14f339364eb8224f23c2d3abc7e79016f3", UrlToken("5c50ae14f339364eb8224f23c2d3abc7e79016f3"))
	})
	t.Run("SHA256", func(t *testing.T) {
		assert.Equal(t, "a948904f2f0f479b8f8197694b30184b0d2ed1c1cd2a1ec0fb85d299a192a447", UrlToken("a948904f2f0f479b8f8197694b30184b0d2ed1c1cd2a1ec0fb85d299a192a447"))
	})
	t.Run("TooLong", func(t *testing.T) {
		assert.Equal(t, "", UrlToken("a948904f2f0f479b8f8197694b30184b0d2ed1c1cd2a1ec0fb85d299a192a4471"))
	})
}

func TestShareToken(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		assert.Equal(t, "4f339364eb8224f23c2d3a", ShareToken("4f339364eb8224f23c2d3a"))
	})
	t.Run("Trim", func(t *testing.T) {
		assert.Equal(t, "4f339364eb8224f23c2d3a", ShareToken("4f339364eb82  24f23c2d  3a "))
	})
	t.Run("SHA1", func(t *testing.T) {
		assert.Equal(t, "5c50ae14f339364eb8224f23c2d3abc7e79016f3", ShareToken("5c50ae14f339364eb8224f23c2d3abc7e79016f3"))
	})
	t.Run("SHA256", func(t *testing.T) {
		assert.Equal(t, "a948904f2f0f479b8f8197694b30184b0d2ed1c1cd2a1ec0fb85d299a192a447", ShareToken("a948904f2f0f479b8f8197694b30184b0d2ed1c1cd2a1ec0fb85d299a192a447"))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", ShareToken(""))
	})
}
