package txt

import (
	"testing"
	"unicode"

	"github.com/stretchr/testify/assert"
)

func TestIs(t *testing.T) {
	t.Run("The quick brown fox.", func(t *testing.T) {
		assert.False(t, Is(unicode.Latin, "The quick brown fox."))
		assert.False(t, Is(unicode.L, "The quick brown fox."))
		assert.False(t, Is(unicode.Letter, "The quick brown fox."))
	})
	t.Run("bridge", func(t *testing.T) {
		assert.True(t, Is(unicode.Latin, "bridge"))
		assert.True(t, Is(unicode.L, "bridge"))
		assert.True(t, Is(unicode.Letter, "bridge"))
	})
	t.Run("桥", func(t *testing.T) {
		assert.False(t, Is(unicode.Latin, "桥"))
		assert.True(t, Is(unicode.L, "桥"))
		assert.True(t, Is(unicode.Letter, "桥"))
	})
	t.Run("桥船", func(t *testing.T) {
		assert.False(t, Is(unicode.Latin, "桥船"))
		assert.True(t, Is(unicode.L, "桥船"))
		assert.True(t, Is(unicode.Letter, "桥船"))
	})
	t.Run("स्थान", func(t *testing.T) {
		assert.False(t, Is(unicode.Latin, "स्थान"))
		assert.False(t, Is(unicode.L, "स्थान"))
		assert.False(t, Is(unicode.Letter, "स्थान"))
		assert.False(t, Is(unicode.Tamil, "स्थान"))
	})
	t.Run("réseau", func(t *testing.T) {
		assert.True(t, Is(unicode.Latin, "réseau"))
		assert.True(t, Is(unicode.L, "réseau"))
		assert.True(t, Is(unicode.Letter, "réseau"))
	})
	t.Run("empty", func(t *testing.T) {
		assert.False(t, Is(unicode.Latin, ""))
	})
}

func TestIsASCII(t *testing.T) {
	t.Run("123", func(t *testing.T) {
		assert.True(t, IsASCII("123"))
	})
	t.Run("The quick brown fox.", func(t *testing.T) {
		assert.True(t, IsASCII("The quick brown fox."))
	})
	t.Run("bridge", func(t *testing.T) {
		assert.True(t, IsASCII("bridge"))
	})
	t.Run("桥", func(t *testing.T) {
		assert.False(t, IsASCII("桥"))
	})
	t.Run("桥船", func(t *testing.T) {
		assert.False(t, IsASCII("桥船"))
	})
	t.Run("स्थान", func(t *testing.T) {
		assert.False(t, IsASCII("स्थान"))
	})
	t.Run("réseau", func(t *testing.T) {
		assert.False(t, IsASCII("réseau"))
	})
}

func TestIsNumeric(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.False(t, IsNumeric(""))
	})
	t.Run("123", func(t *testing.T) {
		assert.True(t, IsNumeric("123"))
	})
	t.Run("123.", func(t *testing.T) {
		assert.False(t, IsNumeric("123."))
	})
	t.Run("2024-10-23", func(t *testing.T) {
		assert.True(t, IsNumeric("2024-10-23"))
	})
	t.Run("20200102-204030", func(t *testing.T) {
		assert.True(t, IsNumeric("20200102-204030"))
	})
	t.Run("ABC", func(t *testing.T) {
		assert.False(t, IsNumeric("ABC"))
	})
}

func TestIsNumber(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.False(t, IsNumber(""))
	})
	t.Run("123", func(t *testing.T) {
		assert.True(t, IsNumber("123"))
	})
	t.Run("123.", func(t *testing.T) {
		assert.False(t, IsNumber("123."))
	})
	t.Run("2024-10-23", func(t *testing.T) {
		assert.False(t, IsNumber("2024-10-23"))
	})
	t.Run("ABC", func(t *testing.T) {
		assert.False(t, IsNumber("ABC"))
	})
}

func TestIsDateNumber(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.False(t, IsDateNumber(""))
	})
	t.Run("123", func(t *testing.T) {
		assert.True(t, IsDateNumber("123"))
	})
	t.Run("123.", func(t *testing.T) {
		assert.False(t, IsDateNumber("123."))
	})
	t.Run("2024-10-23", func(t *testing.T) {
		assert.True(t, IsDateNumber("2024-10-23"))
	})
	t.Run("20200102-204030", func(t *testing.T) {
		assert.True(t, IsDateNumber("20200102-204030"))
	})
	t.Run("ABC", func(t *testing.T) {
		assert.False(t, IsDateNumber("ABC"))
	})
}

func TestIsLatin(t *testing.T) {
	t.Run("The quick brown fox.", func(t *testing.T) {
		assert.False(t, IsLatin("The quick brown fox."))
	})
	t.Run("bridge", func(t *testing.T) {
		assert.True(t, IsLatin("bridge"))
	})
	t.Run("桥", func(t *testing.T) {
		assert.False(t, IsLatin("桥"))
	})
	t.Run("桥船", func(t *testing.T) {
		assert.False(t, IsLatin("桥船"))
	})
	t.Run("स्थान", func(t *testing.T) {
		assert.False(t, IsLatin("स्थान"))
	})
	t.Run("réseau", func(t *testing.T) {
		assert.True(t, IsLatin("réseau"))
	})
	t.Run("empty", func(t *testing.T) {
		assert.False(t, IsLatin(""))
	})
}
