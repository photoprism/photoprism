package util

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

func TestIsSeparator(t *testing.T) {
	t.Run("rune A", func(t *testing.T) {
		assert.Equal(t, false, isSeparator('A'))
	})
	t.Run("rune 99", func(t *testing.T) {
		assert.Equal(t, false, isSeparator('9'))
	})
	t.Run("rune /", func(t *testing.T) {
		assert.Equal(t, true, isSeparator('/'))
	})
	t.Run("rune \\", func(t *testing.T) {
		assert.Equal(t, true, isSeparator('\\'))
	})
	t.Run("rune ♥ ", func(t *testing.T) {
		assert.Equal(t, false, isSeparator('♥'))
	})
	t.Run("rune  space", func(t *testing.T) {
		assert.Equal(t, true, isSeparator(' '))
	})
	t.Run("rune '", func(t *testing.T) {
		assert.Equal(t, false, isSeparator('\''))
	})
	t.Run("rune ý", func(t *testing.T) {
		assert.Equal(t, false, isSeparator('ý'))
	})
}

func TestUcFirst(t *testing.T) {
	t.Run("photo-lover", func(t *testing.T) {
		assert.Equal(t, "Photo-lover", UcFirst("photo-lover"))
	})
	t.Run("cat", func(t *testing.T) {
		assert.Equal(t, "Cat", UcFirst("Cat"))
	})
	t.Run("empty string", func(t *testing.T) {
		assert.Equal(t, "", UcFirst(""))
	})
}

func TestTitle(t *testing.T) {
	t.Run("photo-lover", func(t *testing.T) {
		assert.Equal(t, "Browse Your Life In Pictures", Title("Browse your life in pictures"))
	})
	t.Run("cat", func(t *testing.T) {
		assert.Equal(t, "Photo-Lover", Title("photo-lover"))
	})
}
