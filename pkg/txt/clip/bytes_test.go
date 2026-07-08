package clip

import (
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

func TestBytes(t *testing.T) {
	t.Run("ASCIIShortEnough", func(t *testing.T) {
		result := Bytes("Foo", 16)
		assert.Equal(t, "Foo", result)
	})
	t.Run("ASCIITooLong", func(t *testing.T) {
		result := Bytes("abcdefghij", 4)
		assert.Equal(t, "abcd", result)
		assert.Len(t, result, 4)
	})
	t.Run("TrimSpace", func(t *testing.T) {
		assert.Equal(t, "abc", Bytes("  abc  ", 16))
	})
	t.Run("DoesNotSplitRune", func(t *testing.T) {
		// "äöü" is three 2-byte runes (6 bytes). A 3-byte budget must keep
		// only the first whole rune (2 bytes), never a partial one.
		result := Bytes("äöü", 3)
		assert.Equal(t, "ä", result)
		assert.Len(t, result, 2)
		assert.True(t, utf8.ValidString(result))
	})
	t.Run("ExactRuneBoundary", func(t *testing.T) {
		result := Bytes("äöü", 4)
		assert.Equal(t, "äö", result)
		assert.Len(t, result, 4)
		assert.True(t, utf8.ValidString(result))
	})
	t.Run("EmojiNeverPartial", func(t *testing.T) {
		// Each emoji is 4 bytes; a 1024-byte budget keeps 256 whole emoji.
		s := ""
		for i := 0; i < 400; i++ {
			s += "😀"
		}
		result := Bytes(s, 1024)
		assert.LessOrEqual(t, len(result), 1024)
		assert.True(t, utf8.ValidString(result))
		assert.Equal(t, 256, utf8.RuneCountInString(result))
	})
	t.Run("SingleRuneLargerThanBudget", func(t *testing.T) {
		assert.Equal(t, "", Bytes("😀", 3))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", Bytes("", 16))
	})
	t.Run("NonPositiveBudget", func(t *testing.T) {
		assert.Equal(t, "", Bytes("abc", 0))
		assert.Equal(t, "", Bytes("abc", -1))
	})
}
