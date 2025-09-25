package clean

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/txt/clip"
)

func TestToASCII(t *testing.T) {
	result := ASCII("幸福 = Happiness.")
	assert.Equal(t, " = Happiness.", result)
}

func TestClip(t *testing.T) {
	t.Run("Foo", func(t *testing.T) {
		result := clip.Chars("Foo", 16)
		assert.Equal(t, "Foo", result)
		assert.Equal(t, 3, len(result))
	})
	t.Run("TrimFoo", func(t *testing.T) {
		result := clip.Chars(" Foo ", 16)
		assert.Equal(t, "Foo", result)
		assert.Equal(t, 3, len(result))
	})
	t.Run("TooLong", func(t *testing.T) {
		result := clip.Chars(" 幸福 Hanzi are logograms developed for the writing of Chinese! ", 16)
		assert.Equal(t, "幸福 Hanzi are", result)
		assert.Equal(t, 16, len(result))
	})
	t.Run("ToASCII", func(t *testing.T) {
		result := clip.Chars(ASCII(strings.ToLower(" 幸福 Hanzi are logograms developed for the writing of Chinese! Expressions in an index may not ...!")), LengthType)
		assert.Equal(t, "hanzi are logograms developed for the writing of chinese! expres", result) // codespell:ignore
		assert.Equal(t, 64, len(result))
	})
	t.Run("Empty", func(t *testing.T) {
		result := clip.Chars("", 999)
		assert.Equal(t, "", result)
		assert.Equal(t, 0, len(result))
	})
}

func TestType(t *testing.T) {
	t.Run("Clip", func(t *testing.T) {
		result := Type(" 幸福 Hanzi are logograms developed for the writing of Chinese! Expressions in an index may not ...!")
		assert.Equal(t, "Hanzi are logograms developed for the writing of Chinese! Expres", result) // codespell:ignore
		assert.Equal(t, LengthType, len(result))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", Type(""))
	})
}

func TestTypeLower(t *testing.T) {
	t.Run("Clip", func(t *testing.T) {
		result := TypeLower(" 幸福 Hanzi are logograms developed for the writing of Chinese! Expressions in an index may not ...!")
		assert.Equal(t, "hanzi are logograms developed for the writing of chinese! expres", result) // codespell:ignore
		assert.Equal(t, LengthType, len(result))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", TypeLower(""))
	})
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
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", TypeLowerUnderscore(""))
	})
}

func TestTypeLowerDash(t *testing.T) {
	t.Run("Undefined", func(t *testing.T) {
		assert.Equal(t, "", TypeLowerDash("    \t "))
	})
	t.Run("ClientCredentials", func(t *testing.T) {
		assert.Equal(t, "client-credentials", TypeLowerDash(" Client Credentials幸"))
	})
	t.Run("OllamaModel", func(t *testing.T) {
		assert.Equal(
			t,
			"ollama-model:7b",
			TypeLowerDash("Ollama Model:7b"))
	})
	t.Run("OllamaModelWithSlash", func(t *testing.T) {
		assert.Equal(
			t,
			"ollama-model/7b",
			TypeLowerDash("Ollama Model/7b"))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", TypeLowerDash(""))
	})
}

func TestTypeUnderscore(t *testing.T) {
	t.Run("WhitespaceToUnderscore", func(t *testing.T) {
		in := "a b\tc\nd"
		out := TypeUnderscore(in)
		assert.Equal(t, "a_b_c_d", out)
	})
	t.Run("SpecialsToUnderscore", func(t *testing.T) {
		// Maps (colon and slash allowed): '-', '`', '~', '\\', '|', '"', '\'', '?', '*', '<', '>', '{', '}'
		in := "a-`~/\\:|\"'?*<>{}b"
		out := TypeUnderscore(in)
		assert.Equal(t, "a___/_:_________b", out)
	})
	t.Run("NonASCIIPreserved", func(t *testing.T) {
		assert.Equal(t, "äöü", TypeUnderscore("äöü"))
	})
}

func TestTypeDash(t *testing.T) {
	t.Run("WhitespaceToDash", func(t *testing.T) {
		in := "a b\tc\nd"
		out := TypeDash(in)
		assert.Equal(t, "a-b-c-d", out)
	})
	t.Run("SpecialsToDash", func(t *testing.T) {
		// Maps (colon and slash allowed): '_', '`', '~', '\\', '|', '"', '\'', '?', '*', '<', '>', '{', '}'
		in := "a_`~/\\:|\"'?*<>{}b"
		out := TypeDash(in)
		// 13 mapped; slash and colon preserved → 3 dashes, '/', '-', ':', then 9 dashes
		assert.Equal(t, "a---/-:---------b", out)
	})
	t.Run("NonASCIIPreserved", func(t *testing.T) {
		assert.Equal(t, "äöü", TypeDash("äöü"))
	})
}

func TestShortTypeLowerDash(t *testing.T) {
	t.Run("Undefined", func(t *testing.T) {
		assert.Equal(t, "", ShortTypeLowerDash("    \t "))
	})
	t.Run("ClientCredentials", func(t *testing.T) {
		assert.Equal(t, "client-c", ShortTypeLowerDash(" Client Credentials幸"))
	})
	t.Run("Clip", func(t *testing.T) {
		assert.Equal(t,
			"hanzi-ar",
			ShortTypeLowerDash(" 幸福 Hanzi are logograms developed for the writing of Chinese! Expressions in an index may not ...!"))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", ShortTypeLowerDash(""))
	})
}

func TestShortType(t *testing.T) {
	t.Run("Clip", func(t *testing.T) {
		result := ShortType(" 幸福 Hanzi are logograms developed for the writing of Chinese! Expressions in an index may not ...!")
		assert.Equal(t, "Hanzi ar", result)
		assert.Equal(t, LengthShortType, len(result))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", ShortType(""))
	})
}

func TestShortTypeLower(t *testing.T) {
	t.Run("Clip", func(t *testing.T) {
		result := ShortTypeLower(" 幸福 Hanzi are logograms developed for the writing of Chinese! Expressions in an index may not ...!")
		assert.Equal(t, "hanzi ar", result)
		assert.Equal(t, LengthShortType, len(result))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", ShortTypeLower(""))
	})
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
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", ShortTypeLowerUnderscore(""))
	})
}
