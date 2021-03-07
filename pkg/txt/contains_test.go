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
	t.Run("123", func(t *testing.T) {
		assert.False(t, ContainsSymbols("123"))
	})
	t.Run("The quick brown fox.", func(t *testing.T) {
		assert.False(t, ContainsSymbols("The quick brown fox."))
	})
	t.Run("bridge", func(t *testing.T) {
		assert.False(t, ContainsSymbols("bridge"))
	})
	t.Run("桥", func(t *testing.T) {
		assert.False(t, ContainsSymbols("桥"))
	})
	t.Run("桥船", func(t *testing.T) {
		assert.False(t, ContainsSymbols("桥船"))
	})
	t.Run("स्थान", func(t *testing.T) {
		assert.False(t, ContainsSymbols("स्थान"))
	})
	t.Run("réseau", func(t *testing.T) {
		assert.False(t, ContainsSymbols("réseau"))
	})
	t.Run("empty", func(t *testing.T) {
		assert.False(t, ContainsSymbols(""))
	})
	t.Run("...", func(t *testing.T) {
		assert.True(t, ContainsSymbols("😉"))
	})
}

func TestContainsLetters(t *testing.T) {
	t.Run("123", func(t *testing.T) {
		assert.False(t, ContainsLetters("123"))
	})
	t.Run("The quick brown fox.", func(t *testing.T) {
		assert.False(t, ContainsLetters("The quick brown fox."))
	})
	t.Run("bridge", func(t *testing.T) {
		assert.True(t, ContainsLetters("bridge"))
	})
	t.Run("桥", func(t *testing.T) {
		assert.True(t, ContainsLetters("桥"))
	})
	t.Run("桥船", func(t *testing.T) {
		assert.True(t, ContainsLetters("桥船"))
	})
	t.Run("स्थान", func(t *testing.T) {
		assert.False(t, ContainsLetters("स्थान"))
	})
	t.Run("réseau", func(t *testing.T) {
		assert.True(t, ContainsLetters("réseau"))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, false, ContainsLetters(""))
	})
}

func TestContainsASCIILetters(t *testing.T) {
	t.Run("123", func(t *testing.T) {
		assert.False(t, ContainsASCIILetters("123"))
	})
	t.Run("The quick brown fox.", func(t *testing.T) {
		assert.False(t, ContainsASCIILetters("The quick brown fox."))
	})
	t.Run("bridge", func(t *testing.T) {
		assert.True(t, ContainsASCIILetters("bridge"))
	})
	t.Run("桥", func(t *testing.T) {
		assert.False(t, ContainsASCIILetters("桥"))
	})
	t.Run("桥船", func(t *testing.T) {
		assert.False(t, ContainsASCIILetters("桥船"))
	})
	t.Run("स्थान", func(t *testing.T) {
		assert.False(t, ContainsASCIILetters("स्थान"))
	})
	t.Run("réseau", func(t *testing.T) {
		assert.False(t, ContainsASCIILetters("réseau"))
	})
}
