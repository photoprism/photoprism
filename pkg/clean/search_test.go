package clean

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

func TestSearchString(t *testing.T) {
	t.Run("Replace", func(t *testing.T) {
		q := SearchString("table spoon & usa | img% json OR BILL!\n")
		assert.Equal(t, "table spoon & usa | img* json OR BILL!", q)
	})
	t.Run("AndOr", func(t *testing.T) {
		q := SearchString("Jens AND Mander and me Or Kitty AND ")
		assert.Equal(t, "Jens AND Mander and me Or Kitty AND ", q)
	})
	t.Run("FlowersInThePark", func(t *testing.T) {
		q := SearchString(" Flowers in the Park ")
		assert.Equal(t, " Flowers in the Park ", q)
	})
	t.Run("Empty", func(t *testing.T) {
		q := SearchString("")
		assert.Equal(t, "", q)
	})
}

func TestSearchQuery(t *testing.T) {
	t.Run("Replace", func(t *testing.T) {
		q := SearchQuery("table spoon & usa | img% json OR BILL!\n")
		assert.Equal(t, "table spoon & usa | img* json|BILL!", q)
	})
	t.Run("AndOr", func(t *testing.T) {
		q := SearchQuery("Jens AND Mander and me Or Kitty AND ")
		assert.Equal(t, "Jens&Mander&me|Kitty&", q)
	})
	t.Run("FlowersInThePark", func(t *testing.T) {
		q := SearchQuery(" Flowers in the Park ")
		assert.Equal(t, "Flowers&the Park", q)
	})
	t.Run("Empty", func(t *testing.T) {
		q := SearchQuery("")
		assert.Equal(t, "", q)
	})
}

func BenchmarkSearchQuery_Complex(b *testing.B) {
	s := "Jens AND Mander and me Or Kitty WITH flowers IN the park AT noon | img% json OR BILL!\n"
	b.ReportAllocs()
	for b.Loop() {
		_ = SearchQuery(s)
	}
}

func BenchmarkSearchQuery_Short(b *testing.B) {
	s := "cat and dog"
	b.ReportAllocs()
	for b.Loop() {
		_ = SearchQuery(s)
	}
}

func BenchmarkSearchQuery_LongNoOps(b *testing.B) {
	// No tokens to replace, primarily tests normalization + trim.
	s := strings.Repeat("alpha beta gamma ", 50)
	b.ReportAllocs()
	for b.Loop() {
		_ = SearchQuery(s)
	}
}

func TestSearchTerms(t *testing.T) {
	t.Run("WithinLimit", func(t *testing.T) {
		// Returned byte-identical, whitespace included.
		for _, s := range []string{
			strings.Repeat("a", LengthLimit),
			" " + strings.Repeat("a", 1400) + " ",
			" " + strings.Repeat("日", 1300) + " ",
		} {
			assert.Equal(t, s, SearchTerms(s))
		}
	})
	t.Run("AboveLimit", func(t *testing.T) {
		assert.Len(t, SearchTerms(strings.Repeat("a", LengthLimit*4)), LengthLimit)
	})
	t.Run("MultiByteRunes", func(t *testing.T) {
		// A term must stay valid UTF-8 to bind, so clipping cannot split a rune.
		for _, r := range []string{"é", "€", "日", "😀"} {
			s := SearchTerms(strings.Repeat(r, LengthLimit))
			assert.True(t, utf8.ValidString(s), "clipping %s produced invalid UTF-8", r)
			assert.LessOrEqual(t, len(s), LengthLimit)
			assert.Greater(t, len(s), LengthLimit-utf8.UTFMax)
		}
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", SearchTerms(""))
	})
}

func TestSearchFilter(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		for _, s := range []string{"", "cat", "2020/Holiday", "20*", "a|b", "Grüße 日本"} {
			value, ok := SearchFilter(s)
			assert.True(t, ok, s)
			assert.Equal(t, s, value)
		}
	})
	t.Run("Cleaned", func(t *testing.T) {
		value, ok := SearchFilter("cat\xff\x00")
		assert.True(t, ok)
		assert.Equal(t, "cat", value)
	})
	t.Run("NothingLeft", func(t *testing.T) {
		for _, s := range []string{"\xff", "\x00", "\x00\xfe\xff", "\xc3"} {
			value, ok := SearchFilter(s)
			assert.False(t, ok, "%q", s)
			assert.Equal(t, "", value)
		}
	})
	t.Run("OnlySyntaxLeft", func(t *testing.T) {
		for _, s := range []string{"\x00 ", "\xff|", "\xff*", "%\xff%", "\xff\t", "\"\xff\"", "\xff&+", "\xff/", "\xff!", "\xff/*"} {
			_, ok := SearchFilter(s)
			assert.False(t, ok, "%q", s)
		}
	})
	t.Run("SyntaxWithoutRemovedBytes", func(t *testing.T) {
		for _, s := range []string{" ", "|", "*", "%"} {
			value, ok := SearchFilter(s)
			assert.True(t, ok, "%q", s)
			assert.Equal(t, s, value)
		}
	})
}

func TestSearchUsable(t *testing.T) {
	t.Run("Usable", func(t *testing.T) {
		for _, s := range []string{"cat", "*cat*", "a|b", "/a", "-", "1", "\U0001F600", "\uFFFD"} {
			assert.True(t, SearchUsable(s), "%q", s)
		}
	})
	t.Run("NotUsable", func(t *testing.T) {
		for _, s := range []string{"", " ", "|", "*", "% ", "\t\n", "\x01", "${}", ":", "\"'", "+&", "<>\\", "/", "!", "/*"} {
			assert.False(t, SearchUsable(s), "%q", s)
		}
	})
}

func TestSearchBindable(t *testing.T) {
	t.Run("Unchanged", func(t *testing.T) {
		for _, s := range []string{"", "cat", "Grüße 日本 👩\u200D👩\u200D👧", "a\tb\x01c"} {
			assert.Equal(t, s, SearchBindable(s))
		}
		assert.Zero(t, testing.AllocsPerRun(10, func() { _ = SearchBindable("cat dog") }))
	})
	t.Run("Removed", func(t *testing.T) {
		assert.Equal(t, "", SearchBindable("\x00"))
		assert.Equal(t, "catdog", SearchBindable("cat\x00dog"))
		assert.Equal(t, "catdog", SearchBindable("cat\xffdog"))
		assert.Equal(t, "ab", SearchBindable("a\x00\xc3b"))
	})
}

func TestSearchNul(t *testing.T) {
	for name, fn := range map[string]func(string) string{"SearchString": SearchString, "SearchQuery": SearchQuery, "SearchTerms": SearchTerms} {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, "", fn("\x00"))
			assert.Equal(t, "catdog", fn("cat\x00dog"))
			assert.Equal(t, "catdog", fn("cat\xffdog"))
			assert.Equal(t, "cat dog", fn("cat dog"))
		})
	}
}
