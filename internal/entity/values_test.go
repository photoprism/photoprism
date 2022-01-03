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

func TestTrim(t *testing.T) {
	t.Run("Foo", func(t *testing.T) {
		result := Trim("Foo", 16)
		assert.Equal(t, "Foo", result)
		assert.Equal(t, 3, len(result))
	})
	t.Run("TrimFoo", func(t *testing.T) {
		result := Trim(" Foo ", 16)
		assert.Equal(t, "Foo", result)
		assert.Equal(t, 3, len(result))
	})
	t.Run("TooLong", func(t *testing.T) {
		result := Trim(" 幸福 Hanzi are logograms developed for the writing of Chinese! ", 16)
		assert.Equal(t, "幸福 Hanzi are", result)
		assert.Equal(t, 16, len(result))
	})
	t.Run("ToASCII", func(t *testing.T) {
		result := Trim(ToASCII(strings.ToLower(" 幸福 Hanzi are logograms developed for the writing of Chinese! ")), TrimTypeString)
		assert.Equal(t, "hanzi are logograms developed for the wr", result)
		assert.Equal(t, 40, len(result))
	})
	t.Run("Empty", func(t *testing.T) {
		result := Trim("", 999)
		assert.Equal(t, "", result)
		assert.Equal(t, 0, len(result))
	})
}

func TestSanitizeTypeString(t *testing.T) {
	result := SanitizeTypeString(" 幸福 Hanzi are logograms developed for the writing of Chinese! ")
	assert.Equal(t, "hanzi are logograms developed for the wr", result)
	assert.Equal(t, TrimTypeString, len(result))
}
