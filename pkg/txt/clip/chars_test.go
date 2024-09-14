package clip

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChars(t *testing.T) {
	t.Run("Foo", func(t *testing.T) {
		result := Chars("Foo", 16)
		assert.Equal(t, "Foo", result)
		assert.Equal(t, 3, len(result))
	})
	t.Run("TrimFoo", func(t *testing.T) {
		result := Chars(" Foo ", 16)
		assert.Equal(t, "Foo", result)
		assert.Equal(t, 3, len(result))
	})
	t.Run("TooLong", func(t *testing.T) {
		result := Chars(" 幸福 Hanzi are logograms developed for the writing of Chinese! ", 16)
		assert.Equal(t, "幸福 Hanzi are", result)
		assert.Equal(t, 16, len(result))
	})
	t.Run("Empty", func(t *testing.T) {
		result := Chars("", 999)
		assert.Equal(t, "", result)
		assert.Equal(t, 0, len(result))
	})
}
