package clean

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestID(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", ID(""))
	})
	t.Run("ID", func(t *testing.T) {
		assert.Equal(t, "8045b8926260372bb9afd65e451633ec4ab08817", ID("8045b8926260372bb9afd65e451633ec4ab08817"))
	})
	t.Run("BearerToken", func(t *testing.T) {
		assert.Equal(t, "Bearer S0VLU0UhIExFQ0tFUiEK", ID("Bearer S0VLU0UhIExFQ0tFUiEK"))
	})
	t.Run("Invalid", func(t *testing.T) {
		assert.Equal(t, "@#$$#@#$_8oags JKHUGIUYG gsr\":", ID("  !@#$%^&*(*&^%$#@#$%^&*()_8oags JKHUGIUYG gsr<>?\":{}|{})"))
	})
}

func TestUID(t *testing.T) {
	t.Run("UUID", func(t *testing.T) {
		assert.Equal(t, "123e4567-e89b-12d3-a456-426614174000", UID("123e4567-e89b-12d3-A456-426614174000 "))
	})
	t.Run("ThumbSize", func(t *testing.T) {
		assert.Equal(t, "", UID("left_224"))
	})
	t.Run("SHA1", func(t *testing.T) {
		assert.Equal(t, "5c50ae14f339364eb8224f23c2d3abc7e79016f3", UID(" 5c50ae14f339364eb8224f23c2d3abc7e79016f3 "))
	})
	t.Run("SHA256", func(t *testing.T) {
		assert.Equal(t, "a948904f2f0f479b8f8197694b30184b0d2ed1c1cd2a1ec0fb85d299a192a447", UID("a948904f2f0f479b8f8197694b30184b0d2ed1c1cd2a1ec0fb85d299a192a447"))
	})
	t.Run("Quotes", func(t *testing.T) {
		assert.Equal(t, "foobaaar23", UID("\"foo\" baa'ar 2```3"))
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
