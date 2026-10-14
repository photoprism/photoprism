package clean

import (
	"strings"
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
	t.Run("Controls", func(t *testing.T) {
		assert.Equal(t, "a[U+0000]b[U+0007]c[U+007F]d", Display("a\x00b\x07c\x7fd"))
		assert.Equal(t, "[U+0001][U+001F]", Display("\x01\x1f"))
		assert.Equal(t, "a[U+0080]b[U+009B]c", Display("a\u0080b\u009bc"))
	})
	t.Run("Breaks", func(t *testing.T) {
		for _, s := range []string{"\t", "\n", "\v", "\f", "\r", "\u0085", "\u2028", "\u2029"} {
			assert.Equal(t, "a b", Display("a"+s+"b"), "%q", s)
		}
	})
	t.Run("Marked", func(t *testing.T) {
		assert.Equal(t, "a[U+202E]b", Display("a\u202Eb"))
		assert.Equal(t, "a[U+200B]b", Display("a\u200Bb"))
		assert.Equal(t, "a[U+E0041]b", Display("a\U000E0041b"))
		assert.Equal(t, "a[U+10FFFD]b", Display("a\U0010FFFDb"))
		assert.Equal(t, "a[0xFF]b[0x9B]", Display("a\xffb\x9b"))
		assert.Equal(t, "a\uFFFDb", Display("a\uFFFDb"))
	})
	t.Run("ControlTable", func(t *testing.T) {
		assert.Equal(t, "Ab[U+0000]c[U+0001]d e f[U+001B]g[U+007F]h[0xFF]i", Display(controlSample))
	})
	t.Run("KeepsSpaces", func(t *testing.T) {
		eachRune(func(r rune) {
			if unicode.Is(unicode.Zs, r) {
				assert.Equal(t, "a"+string(r)+"b", Display("a"+string(r)+"b"), "U+%04X", r)
			}
		})
	})
}

func TestDisplayLines(t *testing.T) {
	t.Run("SingleLine", func(t *testing.T) {
		assert.Equal(t, "", DisplayLines(""))
		assert.Equal(t, "status 503: service unavailable", DisplayLines("status 503: service unavailable"))
		assert.Equal(t, "a b[U+200B]c", DisplayLines("a\tb\u200Bc"))
	})
	t.Run("KeepsLineBreaks", func(t *testing.T) {
		assert.Equal(t, "first line\nsecond line", DisplayLines("first line\r\nsecond line"))
		assert.Equal(t, "a\n\nb[U+2069]\n", DisplayLines("a\n\nb\u2069\n"))
		assert.Equal(t, "name: a[U+0007]b\nnext", DisplayLines("name: a\x07b\nnext"))
	})
	t.Run("KeepsText", func(t *testing.T) {
		assert.Equal(t, "رحلة 2026\n日本語 👩\u200D👩\u200D👧", DisplayLines("رحلة 2026\n日本語 👩\u200D👩\u200D👧"))
	})
}

func TestDisplayRunes(t *testing.T) {
	assert.Equal(t, "", DisplayRunes("admin", 0))
	assert.Equal(t, "admin", DisplayRunes("admin", 64))
	assert.Equal(t, strings.Repeat("[U+0000]", 8), DisplayRunes(strings.Repeat("\x00", 300)+"admin", 64))
	assert.Equal(t, "ad[U+0007]", DisplayRunes("ad\x07min", 10))
	assert.Equal(t, "ad", DisplayRunes("ad\x07min", 9))
	assert.Equal(t, strings.Repeat("[0xFF]", 10), DisplayRunes(strings.Repeat("\xff", 1<<20), 64))
	assert.Equal(t, "رحلة", DisplayRunes("رحلة 2026", 4))
	assert.Equal(t, "admin", DisplayRunes(strings.Repeat(" ", 256)+"admin", 64))
	assert.Equal(t, "admin", DisplayRunes(strings.Repeat("\u3000", 86)+"\u2028admin", 64))
}

func TestDisplayRune(t *testing.T) {
	assert.Equal(t, 'a', displayRune('a'))
	assert.Equal(t, 'ر', displayRune('ر'))
	assert.Equal(t, rune(0x3000), displayRune(0x3000))
	assert.Equal(t, ' ', displayRune('\n'))
	assert.Equal(t, ' ', displayRune(0x2029))
	assert.Equal(t, markRune, displayRune(0))
	assert.Equal(t, markRune, displayRune(0x202E))
	assert.Equal(t, rune(0x200D), displayRune(0x200D))
}
