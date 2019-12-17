package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandomToken(t *testing.T) {
	t.Run("size 4", func(t *testing.T) {
		token := RandomToken(4)
		assert.NotEmpty(t, token)
	})
	t.Run("size 8", func(t *testing.T) {
		token := RandomToken(9)
		assert.NotEmpty(t, token)
	})
}

func TestUUID(t *testing.T) {
	t.Run("size 4", func(t *testing.T) {
		uuid := UUID()
		assert.Equal(t, 36, len(uuid))
	})
	t.Run("size 8", func(t *testing.T) {
		uuid := UUID()
		assert.Equal(t, 36, len(uuid))
	})
}
