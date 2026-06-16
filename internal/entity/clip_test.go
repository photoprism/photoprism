package entity

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

func TestToASCII(t *testing.T) {
	result := ToASCII("幸福 = Happiness.")
	assert.Equal(t, " = Happiness.", result)
}

func TestClip(t *testing.T) {
	t.Run("Foo", func(t *testing.T) {
		result := Clip("Foo", 16)
		assert.Equal(t, "Foo", result)
		assert.Equal(t, 3, len(result))
	})
	t.Run("TrimFoo", func(t *testing.T) {
		result := Clip(" Foo ", 16)
		assert.Equal(t, "Foo", result)
		assert.Equal(t, 3, len(result))
	})
	t.Run("TooLong", func(t *testing.T) {
		result := Clip(" 幸福 Hanzi are logograms developed for the writing of Chinese! ", 16)
		assert.Equal(t, "幸福 Hanzi are", result)
		assert.Equal(t, 16, len(result))
	})
	t.Run("ToASCII", func(t *testing.T) {
		result := Clip(ToASCII(strings.ToLower(" 幸福 Hanzi are logograms developed for the writing of Chinese! Expressions in an index may not ...!")), TypeBytes)
		assert.Equal(t, "hanzi are logograms developed for the writing of chinese! expres", result)
		assert.Equal(t, 64, len(result))
	})
	t.Run("Empty", func(t *testing.T) {
		result := Clip("", 999)
		assert.Equal(t, "", result)
		assert.Equal(t, 0, len(result))
	})
}

// TestClipPath verifies that paths are limited to the PathBytes byte budget
// without splitting a multi-byte rune.
func TestClipPath(t *testing.T) {
	t.Run("ShortPath", func(t *testing.T) {
		assert.Equal(t, "2024/Vacation", ClipPath("2024/Vacation"))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", ClipPath(""))
	})
	t.Run("MultiByteOverflow", func(t *testing.T) {
		// 400 four-byte emoji = 1600 bytes, over the 1024-byte budget.
		path := strings.Repeat("😀", 400)
		result := ClipPath(path)
		assert.LessOrEqual(t, len(result), PathBytes)
		assert.True(t, utf8.ValidString(result))
		assert.Equal(t, 256, utf8.RuneCountInString(result))
	})
}

func TestClipType(t *testing.T) {
	result := ClipType(" 幸福 Hanzi are logograms developed for the writing of Chinese! Expressions in an index may not ...!")
	assert.Equal(t, "Hanzi are logograms developed for the writing of Chinese! Expres", result) // codespell:ignore
	assert.Equal(t, TypeBytes, len(result))
}

func TestClipTypeLower(t *testing.T) {
	result := ClipTypeLower(" 幸福 Hanzi are logograms developed for the writing of Chinese! Expressions in an index may not ...!")
	assert.Equal(t, "hanzi are logograms developed for the writing of chinese! expres", result) // codespell:ignore
	assert.Equal(t, TypeBytes, len(result))
}
