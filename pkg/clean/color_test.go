package clean

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestColor(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", Color(""))
	})
	t.Run("Black", func(t *testing.T) {
		assert.Equal(t, "#000000", Color("#000000"))
	})
	t.Run("White", func(t *testing.T) {
		assert.Equal(t, "#ffffff", Color("#FFFFFF"))
	})
	t.Run("Short", func(t *testing.T) {
		assert.Equal(t, "#ab1", Color("#aB1"))
	})
	t.Run("Alpha", func(t *testing.T) {
		assert.Equal(t, "#0123456a", Color("#0123456A"))
	})
	t.Run("TooLong", func(t *testing.T) {
		assert.Equal(t, "", Color("#01234567AA"))
	})
	t.Run("TooShort", func(t *testing.T) {
		assert.Equal(t, "", Color("#00"))
	})
}
