package clip

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunes(t *testing.T) {
	t.Run("Foo", func(t *testing.T) {
		result := Runes("Foo", 16)
		assert.Equal(t, "Foo", result)
		assert.Equal(t, 3, len(result))
	})
	t.Run("TrimFoo", func(t *testing.T) {
		result := Runes(" Foo ", 16)
		assert.Equal(t, "Foo", result)
		assert.Equal(t, 3, len(result))
	})
	t.Run("TooLong", func(t *testing.T) {
		result := Runes(" 幸福 Hanzi are logograms developed for the writing of Chinese! ", 16)
		assert.Equal(t, "幸福 Hanzi are log", result)
		assert.Equal(t, 20, len(result))
	})
	t.Run("Empty", func(t *testing.T) {
		result := Runes("", 999)
		assert.Equal(t, "", result)
		assert.Equal(t, 0, len(result))
	})
}
