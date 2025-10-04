package txt

import (
	"testing"
	"unicode"

	"github.com/stretchr/testify/assert"
)

func TestIs(t *testing.T) {
	t.Run("TheQuickBrownFox", func(t *testing.T) {
		assert.False(t, Is(unicode.Latin, "The quick brown fox."))
		assert.False(t, Is(unicode.L, "The quick brown fox."))
		assert.False(t, Is(unicode.Letter, "The quick brown fox."))
	})
	t.Run("Bridge", func(t *testing.T) {
		assert.True(t, Is(unicode.Latin, "bridge"))
		assert.True(t, Is(unicode.L, "bridge"))
		assert.True(t, Is(unicode.Letter, "bridge"))
	})
	t.Run("Case", func(t *testing.T) {
		assert.False(t, Is(unicode.Latin, "桥"))
		assert.True(t, Is(unicode.L, "桥"))
		assert.True(t, Is(unicode.Letter, "桥"))
	})
	t.Run("Case", func(t *testing.T) {
		assert.False(t, Is(unicode.Latin, "桥船"))
		assert.True(t, Is(unicode.L, "桥船"))
		assert.True(t, Is(unicode.Letter, "桥船"))
	})
	t.Run("Case", func(t *testing.T) {
		assert.False(t, Is(unicode.Latin, "स्थान"))
		assert.False(t, Is(unicode.L, "स्थान"))
		assert.False(t, Is(unicode.Letter, "स्थान"))
		assert.False(t, Is(unicode.Tamil, "स्थान"))
	})
	t.Run("RSeau", func(t *testing.T) {
		assert.True(t, Is(unicode.Latin, "réseau"))
		assert.True(t, Is(unicode.L, "réseau"))
		assert.True(t, Is(unicode.Letter, "réseau"))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.False(t, Is(unicode.Latin, ""))
	})
}

func TestIsASCII(t *testing.T) {
	t.Run("One", func(t *testing.T) {
		assert.True(t, IsASCII("1"))
	})
	t.Run("Num123", func(t *testing.T) {
		assert.True(t, IsASCII("123"))
	})
	t.Run("TheQuickBrownFox", func(t *testing.T) {
		assert.True(t, IsASCII("The quick brown fox."))
	})
	t.Run("Bridge", func(t *testing.T) {
		assert.True(t, IsASCII("bridge"))
	})
	t.Run("Case", func(t *testing.T) {
		assert.False(t, IsASCII("桥"))
	})
	t.Run("Case", func(t *testing.T) {
		assert.False(t, IsASCII("桥船"))
	})
	t.Run("Case", func(t *testing.T) {
		assert.False(t, IsASCII("स्थान"))
	})
	t.Run("RSeau", func(t *testing.T) {
		assert.False(t, IsASCII("réseau"))
	})
	t.Run("Num80S", func(t *testing.T) {
		assert.True(t, IsASCII("80s"))
	})
}

func TestIsNumeric(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.False(t, IsNumeric(""))
	})
	t.Run("One", func(t *testing.T) {
		assert.True(t, IsNumeric("1"))
	})
	t.Run("Num123", func(t *testing.T) {
		assert.True(t, IsNumeric("123"))
	})
	t.Run("Num123", func(t *testing.T) {
		assert.False(t, IsNumeric("123."))
	})
	t.Run("Num2024TenNum23", func(t *testing.T) {
		assert.True(t, IsNumeric("2024-10-23"))
	})
	t.Run("Num20200102Num204030", func(t *testing.T) {
		assert.True(t, IsNumeric("20200102-204030"))
	})
	t.Run("ABC", func(t *testing.T) {
		assert.False(t, IsNumeric("ABC"))
	})
	t.Run("Num80S", func(t *testing.T) {
		assert.False(t, IsNumeric("80s"))
	})
	t.Run("TwoE4", func(t *testing.T) {
		assert.True(t, IsNumeric("2e4"))
	})
	t.Run("TwoE", func(t *testing.T) {
		assert.False(t, IsNumeric("2e"))
	})
}

func TestIsNumeral(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.False(t, IsNumeral(""))
	})
	t.Run("One", func(t *testing.T) {
		assert.False(t, IsNumeral("1"))
	})
	t.Run("Num123", func(t *testing.T) {
		assert.False(t, IsNumeral("123"))
	})
	t.Run("Num123", func(t *testing.T) {
		assert.False(t, IsNumeral("123."))
	})
	t.Run("Num2024TenNum23", func(t *testing.T) {
		assert.False(t, IsNumeral("2024-10-23"))
	})
	t.Run("Num20200102Num204030", func(t *testing.T) {
		assert.False(t, IsNumeral("20200102-204030"))
	})
	t.Run("ABC", func(t *testing.T) {
		assert.False(t, IsNumeral("ABC"))
	})
	t.Run("OneSt", func(t *testing.T) {
		assert.True(t, IsNumeral("1st"))
	})
	t.Run("TwoNd", func(t *testing.T) {
		assert.True(t, IsNumeral("1ND")) // codespell:ignore
	})
	t.Run("Num40Th", func(t *testing.T) {
		assert.True(t, IsNumeral("40th"))
	})
	t.Run("One", func(t *testing.T) {
		assert.False(t, IsNumeral("-1."))
	})
	t.Run("One", func(t *testing.T) {
		assert.False(t, IsNumeral("1."))
	})
	t.Run("Num40", func(t *testing.T) {
		assert.False(t, IsNumeral("40."))
	})
	t.Run("Num80S", func(t *testing.T) {
		assert.True(t, IsNumeral("80s"))
	})
}

func TestIsNumber(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.False(t, IsNumber(""))
	})
	t.Run("One", func(t *testing.T) {
		assert.True(t, IsNumber("1"))
	})
	t.Run("Num123", func(t *testing.T) {
		assert.True(t, IsNumber("123"))
	})
	t.Run("Num123", func(t *testing.T) {
		assert.False(t, IsNumber("123."))
	})
	t.Run("Num2024TenNum23", func(t *testing.T) {
		assert.False(t, IsNumber("2024-10-23"))
	})
	t.Run("ABC", func(t *testing.T) {
		assert.False(t, IsNumber("ABC"))
	})
	t.Run("Num80S", func(t *testing.T) {
		assert.False(t, IsNumber("80s"))
	})
}

func TestIsDateNumber(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.False(t, IsDateNumber(""))
	})
	t.Run("Num123", func(t *testing.T) {
		assert.True(t, IsDateNumber("123"))
	})
	t.Run("Num123", func(t *testing.T) {
		assert.False(t, IsDateNumber("123."))
	})
	t.Run("Num2024TenNum23", func(t *testing.T) {
		assert.True(t, IsDateNumber("2024-10-23"))
	})
	t.Run("Num20200102Num204030", func(t *testing.T) {
		assert.True(t, IsDateNumber("20200102-204030"))
	})
	t.Run("ABC", func(t *testing.T) {
		assert.False(t, IsDateNumber("ABC"))
	})
}

func TestIsLatin(t *testing.T) {
	t.Run("TheQuickBrownFox", func(t *testing.T) {
		assert.False(t, IsLatin("The quick brown fox."))
	})
	t.Run("Bridge", func(t *testing.T) {
		assert.True(t, IsLatin("bridge"))
	})
	t.Run("Case", func(t *testing.T) {
		assert.False(t, IsLatin("桥"))
	})
	t.Run("Case", func(t *testing.T) {
		assert.False(t, IsLatin("桥船"))
	})
	t.Run("Case", func(t *testing.T) {
		assert.False(t, IsLatin("स्थान"))
	})
	t.Run("RSeau", func(t *testing.T) {
		assert.True(t, IsLatin("réseau"))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.False(t, IsLatin(""))
	})
}
