package clean

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToASCII(t *testing.T) {
	result := ASCII("幸福 = Happiness.")
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
		result := Clip(ASCII(strings.ToLower(" 幸福 Hanzi are logograms developed for the writing of Chinese! Expressions in an index may not ...!")), ClipType)
		assert.Equal(t, "hanzi are logograms developed for the writing of chinese! expres", result)
		assert.Equal(t, 64, len(result))
	})
	t.Run("Empty", func(t *testing.T) {
		result := Clip("", 999)
		assert.Equal(t, "", result)
		assert.Equal(t, 0, len(result))
	})
}

func TestType(t *testing.T) {
	result := Type(" 幸福 Hanzi are logograms developed for the writing of Chinese! Expressions in an index may not ...!")
	assert.Equal(t, "Hanzi are logograms developed for the writing of Chinese! Expres", result)
	assert.Equal(t, ClipType, len(result))
}

func TestTypeLower(t *testing.T) {
	result := TypeLower(" 幸福 Hanzi are logograms developed for the writing of Chinese! Expressions in an index may not ...!")
	assert.Equal(t, "hanzi are logograms developed for the writing of chinese! expres", result)
	assert.Equal(t, ClipType, len(result))
}

func TestTypeLowerUnderscore(t *testing.T) {
	t.Run("Undefined", func(t *testing.T) {
		assert.Equal(t, "", TypeLowerUnderscore("    \t "))
	})
	t.Run("ClientCredentials", func(t *testing.T) {
		assert.Equal(t, "client_credentials", TypeLowerUnderscore(" Client Credentials幸"))
	})
	t.Run("Clip", func(t *testing.T) {
		assert.Equal(
			t,
			"hanzi_are_logograms_developed_for_the_writing_of_chinese!_expres",
			TypeLowerUnderscore(" 幸福 Hanzi are logograms developed for the writing of Chinese! Expressions in an index may not ...!"))
	})
}

func TestShortType(t *testing.T) {
	result := ShortType(" 幸福 Hanzi are logograms developed for the writing of Chinese! Expressions in an index may not ...!")
	assert.Equal(t, "Hanzi ar", result)
	assert.Equal(t, ClipShortType, len(result))
}

func TestShortTypeLower(t *testing.T) {
	result := ShortTypeLower(" 幸福 Hanzi are logograms developed for the writing of Chinese! Expressions in an index may not ...!")
	assert.Equal(t, "hanzi ar", result)
	assert.Equal(t, ClipShortType, len(result))
}

func TestShortTypeLowerUnderscore(t *testing.T) {
	t.Run("Undefined", func(t *testing.T) {
		assert.Equal(t, "", ShortTypeLowerUnderscore("    \t "))
	})
	t.Run("ClientCredentials", func(t *testing.T) {
		assert.Equal(t, "client_c", ShortTypeLowerUnderscore(" Client Credentials幸"))
	})
	t.Run("Clip", func(t *testing.T) {
		assert.Equal(t,
			"hanzi_ar",
			ShortTypeLowerUnderscore(" 幸福 Hanzi are logograms developed for the writing of Chinese! Expressions in an index may not ...!"))
	})
}
