package clip

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMiddle(t *testing.T) {
	t.Run("ShortEnough", func(t *testing.T) {
		assert.Equal(t, "IMG_1234.jpg", Middle("IMG_1234.jpg", 12, 0))
		assert.Equal(t, "", Middle("", 8, 0))
	})
	t.Run("KeepsEnd", func(t *testing.T) {
		result := Middle("2026/10/vacation-at-the-sea/IMG_20261013_154233.jpg", 24, 8)
		assert.Equal(t, "2026/10/vacat…4233.jpg", result)
		assert.LessOrEqual(t, len(result), 24)
	})
	t.Run("RelativeTail", func(t *testing.T) {
		result := Middle(strings.Repeat("a", 40)+"-end", 23, 0)
		assert.Equal(t, strings.Repeat("a", 14)+"…aa-end", result)
		assert.LessOrEqual(t, len(result), 23)
	})
	t.Run("Multibyte", func(t *testing.T) {
		result := Middle("رحلة إلى البحر.heic", 16, 6)
		assert.Equal(t, "رحلة….heic", result)
		assert.LessOrEqual(t, len(result), 16)
	})
	t.Run("TailBeyondBudget", func(t *testing.T) {
		result := Middle("abcdefghijklmnopqrstuvwxyz", 10, 50)
		assert.Equal(t, "abcde…yz", result)
		assert.LessOrEqual(t, len(result), 10)
	})
	t.Run("LongInput", func(t *testing.T) {
		result := Middle(strings.Repeat("a", 1<<20)+"-end.jpg", 30, 8)
		assert.Equal(t, strings.Repeat("a", 19)+"…-end.jpg", result)
	})
	t.Run("NoRoom", func(t *testing.T) {
		assert.Equal(t, "", Middle("abcdef", 3, 0))
		assert.Equal(t, "a…", Middle("abcdef", 4, 0))
	})
}

func TestMiddleFunc(t *testing.T) {
	// unit keeps a bracketed sequence in one piece.
	unit := func(s string) int {
		if strings.HasPrefix(s, "[") {
			if i := strings.IndexByte(s, ']'); i > 0 {
				return i + 1
			}
		}
		return runeLen(s)
	}

	t.Run("KeepsUnits", func(t *testing.T) {
		result := MiddleFunc("ab[U+202E]cdef[U+200B]gh", 16, 8, unit)
		assert.Equal(t, "ab[U+202E]c…gh", result)
		assert.LessOrEqual(t, len(result), 16)
		result = MiddleFunc("ab[U+202E]cdef[U+200B]gh", 16, 10, unit)
		assert.Equal(t, "ab…[U+200B]gh", result)
		assert.LessOrEqual(t, len(result), 16)
	})
	t.Run("ShortEnough", func(t *testing.T) {
		assert.Equal(t, "a[b]c", MiddleFunc("a[b]c", 5, 0, unit))
	})
	t.Run("ZeroUnit", func(t *testing.T) {
		zero := func(string) int { return 0 }
		assert.Equal(t, "abc…yz", MiddleFunc("abcdefghijklmnopqrstuvwxyz", 8, 2, zero))
	})
	t.Run("LastUnitTooLong", func(t *testing.T) {
		assert.Equal(t, "abcdef…", MiddleFunc("abcdefgh[U+202E]", 9, 3, unit))
	})
}

func TestEnds(t *testing.T) {
	assert.Equal(t, "abcdef", Ends("abcdef", 6))
	assert.Equal(t, "abef", Ends("abcdef", 4))
	assert.Equal(t, "ä", Ends("äbc€", 5))
	assert.Equal(t, "äb€", Ends("äbc€", 6))
	assert.Equal(t, "", Ends("abc", 0))
}

func TestRuneLen(t *testing.T) {
	assert.Equal(t, 1, runeLen("abc"))
	assert.Equal(t, 2, runeLen("ä"))
	assert.Equal(t, 3, runeLen("€"))
	assert.Equal(t, 1, runeLen("\xff"))
	assert.Equal(t, 0, runeLen(""))
}
