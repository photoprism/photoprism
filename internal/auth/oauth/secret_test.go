package oauth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashSecret(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		assert.Equal(t, "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824", HashSecret("hello"))
	})
	t.Run("Distinct", func(t *testing.T) {
		assert.NotEqual(t, HashSecret("a"), HashSecret("b"))
		assert.Len(t, HashSecret(""), 64)
	})
}
