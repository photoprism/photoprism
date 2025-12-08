package txt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsSeparator(t *testing.T) {
	t.Run("RuneA", func(t *testing.T) {
		assert.Equal(t, false, isSeparator('A'))
	})
	t.Run("RuneNum99", func(t *testing.T) {
		assert.Equal(t, false, isSeparator('9'))
	})
	t.Run("Rune", func(t *testing.T) {
		assert.Equal(t, true, isSeparator('/'))
	})
	t.Run("Rune", func(t *testing.T) {
		assert.Equal(t, true, isSeparator('\\'))
	})
	t.Run("Rune", func(t *testing.T) {
		assert.Equal(t, false, isSeparator('♥'))
	})
	t.Run("RuneSpace", func(t *testing.T) {
		assert.Equal(t, true, isSeparator(' '))
	})
	t.Run("Rune", func(t *testing.T) {
		assert.Equal(t, false, isSeparator('\''))
	})
	t.Run("Rune", func(t *testing.T) {
		assert.Equal(t, false, isSeparator('ý'))
	})
}
