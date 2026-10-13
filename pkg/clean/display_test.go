package clean

import (
	"testing"
	"unicode"

	"github.com/stretchr/testify/assert"
)

func TestDisplay(t *testing.T) {
	t.Run("Unchanged", func(t *testing.T) {
		for _, s := range []string{
			"",
			"admin",
			"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko)",
			"invalid credentials: 'user' <x> {y} $z \\w",
			"صورة.heic",
			"שלום",
			"日本語のファイル.png",
			"👩\u200D👩\u200D👧",
			"می\u200Cخواهم",
			"*optional*\u2002value 100\u202F%",
		} {
			assert.Equal(t, s, Display(s))
		}
		assert.Zero(t, testing.AllocsPerRun(10, func() { _ = Display("Europe/Berlin") }))
	})
	t.Run("Whitespace", func(t *testing.T) {
		assert.Equal(t, "a b c d e f", Display("a\tb\nc\vd\fe\rf"))
	})
	t.Run("Removed", func(t *testing.T) {
		assert.Equal(t, "ab[31mcd", Display("a\x00b\x1b[31mc\x7fd"))
		assert.Equal(t, "[31mred[0m", Display("\x1b[31mred\x1b[0m\x07"))
		assert.Equal(t, "", Display("\x00\x01\x1f\x7f"))
	})
	t.Run("Marked", func(t *testing.T) {
		assert.Equal(t, "a?b", Display("a\u202Eb"))
		assert.Equal(t, "a?b", Display("a\u0085b"))
		assert.Equal(t, "a?b", Display("a\u009bb"))
		assert.Equal(t, "a?b", Display("a\u2028b"))
		assert.Equal(t, "a?b", Display("a\u200Bb"))
		assert.Equal(t, "a�b", Display("a\xffb"))
	})
	t.Run("ControlTable", func(t *testing.T) {
		assert.Equal(t, "Abcd e fgh�i", Display(controlSample))
	})
	t.Run("KeepsSpaces", func(t *testing.T) {
		eachRune(func(r rune) {
			if unicode.Is(unicode.Zs, r) {
				assert.Equal(t, "a"+string(r)+"b", Display("a"+string(r)+"b"), "U+%04X", r)
			}
		})
	})
}
