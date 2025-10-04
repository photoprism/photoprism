package txt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContainsNumber(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		assert.Equal(t, true, ContainsNumber("f3abcde"))
	})
	t.Run("False", func(t *testing.T) {
		assert.Equal(t, false, ContainsNumber("abcd"))
	})
}

func TestContainsSymbols(t *testing.T) {
	t.Run("Num123", func(t *testing.T) {
		assert.False(t, ContainsSymbols("123"))
	})
	t.Run("TheQuickBrownFox", func(t *testing.T) {
		assert.False(t, ContainsSymbols("The quick brown fox."))
	})
	t.Run("Bridge", func(t *testing.T) {
		assert.False(t, ContainsSymbols("bridge"))
	})
	t.Run("Case", func(t *testing.T) {
		assert.False(t, ContainsSymbols("桥"))
	})
	t.Run("Case", func(t *testing.T) {
		assert.False(t, ContainsSymbols("桥船"))
	})
	t.Run("Case", func(t *testing.T) {
		assert.False(t, ContainsSymbols("स्थान"))
	})
	t.Run("RSeau", func(t *testing.T) {
		assert.False(t, ContainsSymbols("réseau"))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.False(t, ContainsSymbols(""))
	})
	t.Run("Case", func(t *testing.T) {
		assert.True(t, ContainsSymbols("😉"))
	})
}

func TestContainsLetters(t *testing.T) {
	t.Run("Num123", func(t *testing.T) {
		assert.False(t, ContainsLetters("123"))
	})
	t.Run("TheQuickBrownFox", func(t *testing.T) {
		assert.False(t, ContainsLetters("The quick brown fox."))
	})
	t.Run("Bridge", func(t *testing.T) {
		assert.True(t, ContainsLetters("bridge"))
	})
	t.Run("Case", func(t *testing.T) {
		assert.True(t, ContainsLetters("桥"))
	})
	t.Run("Case", func(t *testing.T) {
		assert.True(t, ContainsLetters("桥船"))
	})
	t.Run("Case", func(t *testing.T) {
		assert.False(t, ContainsLetters("स्थान"))
	})
	t.Run("RSeau", func(t *testing.T) {
		assert.True(t, ContainsLetters("réseau"))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, false, ContainsLetters(""))
	})
}

func TestContainsASCIILetters(t *testing.T) {
	t.Run("Num123", func(t *testing.T) {
		assert.False(t, ContainsASCIILetters("123"))
	})
	t.Run("TheQuickBrownFox", func(t *testing.T) {
		assert.False(t, ContainsASCIILetters("The quick brown fox."))
	})
	t.Run("Bridge", func(t *testing.T) {
		assert.True(t, ContainsASCIILetters("bridge"))
	})
	t.Run("Case", func(t *testing.T) {
		assert.False(t, ContainsASCIILetters("桥"))
	})
	t.Run("Case", func(t *testing.T) {
		assert.False(t, ContainsASCIILetters("桥船"))
	})
	t.Run("Case", func(t *testing.T) {
		assert.False(t, ContainsASCIILetters("स्थान"))
	})
	t.Run("RSeau", func(t *testing.T) {
		assert.False(t, ContainsASCIILetters("réseau"))
	})
}

func TestContainsAlnumLower(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		assert.True(t, ContainsAlnumLower(""))
		assert.True(t, ContainsAlnumLower("a"))
		assert.True(t, ContainsAlnumLower("3kmib24yr3"))
		assert.True(t, ContainsAlnumLower("123"))
	})
	t.Run("False", func(t *testing.T) {
		assert.False(t, ContainsAlnumLower("-"))
		assert.False(t, ContainsAlnumLower(" "))
		assert.False(t, ContainsAlnumLower("B"))
		assert.False(t, ContainsAlnumLower("3Km"))
		assert.False(t, ContainsAlnumLower("_3kmib24yr3"))
	})
}

func BenchmarkContainsNumber(b *testing.B) {
	s := "The quick brown fox jumps over 13 lazy dogs"
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = ContainsNumber(s)
	}
}

func BenchmarkSortCaseInsensitive(b *testing.B) {
	words := []string{"Zebra", "apple", "Banana", "cherry", "Apricot", "banana", "zebra", "Cherry"}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		w := append([]string(nil), words...)
		SortCaseInsensitive(w)
	}
}
