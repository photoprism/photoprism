package entity

import (
	"strings"
	"testing"

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
		result := Clip(ToASCII(strings.ToLower(" 幸福 Hanzi are logograms developed for the writing of Chinese! Expressions in an index may not ...!")), ClipStringType)
		assert.Equal(t, "hanzi are logograms developed for the writing of chinese! expres", result)
		assert.Equal(t, 64, len(result))
	})
	t.Run("Empty", func(t *testing.T) {
		result := Clip("", 999)
		assert.Equal(t, "", result)
		assert.Equal(t, 0, len(result))
	})
}

func TestSanitizeStringType(t *testing.T) {
	result := SanitizeStringType(" 幸福 Hanzi are logograms developed for the writing of Chinese! Expressions in an index may not ...!")
	assert.Equal(t, "Hanzi are logograms developed for the writing of Chinese! Expres", result)
	assert.Equal(t, ClipStringType, len(result))
}

func TestSanitizeStringTypeLower(t *testing.T) {
	result := SanitizeStringTypeLower(" 幸福 Hanzi are logograms developed for the writing of Chinese! Expressions in an index may not ...!")
	assert.Equal(t, "hanzi are logograms developed for the writing of chinese! expres", result)
	assert.Equal(t, ClipStringType, len(result))
}
