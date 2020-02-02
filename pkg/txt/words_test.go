package txt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWords(t *testing.T) {
	t.Run("I'm a lazy brown fox!", func(t *testing.T) {
		result := Words("I'm a lazy BRoWN fox!")
		assert.Equal(t, []string{"lazy", "BRoWN", "fox"}, result)
	})
	t.Run("no result", func(t *testing.T) {
		result := Words("x")
		assert.Equal(t, []string(nil), result)
	})
}

func TestKeywords(t *testing.T) {
	t.Run("I'm a lazy brown fox!", func(t *testing.T) {
		result := Keywords("I'm a lazy BRoWN img!")
		assert.Equal(t, []string{"lazy", "brown"}, result)
	})
	t.Run("no result", func(t *testing.T) {
		result := Keywords("was")
		assert.Equal(t, []string(nil), result)
	})
}
