package clean

import (
	"math/rand/v2"
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
	t.Run("ControlChars", func(t *testing.T) {
		assert.Equal(t, "ab c", Type("\x00a\x1fb\tc\x7f"))
		assert.Equal(t, "", Type("\x00\x01"))
		assert.Equal(t, "abcd", Type("a\u0085b\u2028c\u202ed"))
	})
	t.Run("FilterBeforeClip", func(t *testing.T) {
		assert.Equal(t, "a", Type("a \x00"))
		assert.Equal(t, "abc", Type(strings.Repeat("\x00", 64)+"abc"))
	})
}

func TestTypeString(t *testing.T) {
	t.Run("Unchanged", func(t *testing.T) {
		result, changed, clipped := typeString(" image/jpeg ", LengthType)
		assert.Equal(t, "image/jpeg", result)
		assert.False(t, changed)
		assert.False(t, clipped)
		assert.Zero(t, testing.AllocsPerRun(10, func() { _, _, _ = typeString("Europe/Berlin", LengthType) }))
	})
	t.Run("Changed", func(t *testing.T) {
		result, changed, clipped := typeString("a\tb\x00c\xffd", LengthType)
		assert.Equal(t, "a bcd", result)
		assert.True(t, changed)
		assert.False(t, clipped)
	})
	t.Run("Clipped", func(t *testing.T) {
		result, changed, clipped := typeString("abc def", 4)
		assert.Equal(t, "abc", result)
		assert.False(t, changed)
		assert.True(t, clipped)
	})
	t.Run("SameAsComposition", func(t *testing.T) {
		// The loop must give the same result as filtering with ASCII and StripControl, then clipping.
		alphabet := []string{"a", "Z", "0", " ", "\t", "\n", "\r", "\x00", "\x01", "\x1b", "\x7f", "ä", "幸", "\u0085", "\u00a0", "\u202e", "\xff", ":", "/", "~", "\v", "\f", "\x1f", "\x80"}
		rng := rand.New(rand.NewPCG(1, 2)) //nolint:gosec // a fixed seed keeps a failing input reproducible.

		for range 20000 {
			var sb strings.Builder

			for range rng.IntN(80) {
				sb.WriteString(alphabet[rng.IntN(len(alphabet))])
			}

			in := sb.String()
			assert.Equal(t, clip.Chars(StripControl(ASCII(in)), LengthType), Type(in), "%q", in)
			assert.Equal(t, clip.Chars(StripControl(ASCII(in)), LengthShortType), ShortType(in), "%q", in)
		}
	})
}

func TestTypeUnicode(t *testing.T) {
	t.Run("UnsafeRunes", func(t *testing.T) {
		assert.Equal(t, "ab", TypeUnicode("a\x7fb"))
		assert.Equal(t, "ab", TypeUnicode("a\u009bb"))
		assert.Equal(t, "ab", TypeUnicode("a\u202Eb"))
		assert.Equal(t, "ab", TypeUnicode("a\u200Bb"))
		assert.Equal(t, "ab", TypeUnicode("a\uFEFFb"))
		assert.Equal(t, "a b", TypeUnicode("a\u2028b"))
	})
	t.Run("JoinersKept", func(t *testing.T) {
		assert.Equal(t, "👩\u200D👩\u200D👧", TypeUnicode("👩\u200D👩\u200D👧"))
		assert.Equal(t, "می\u200Cخواهم", TypeUnicode("می\u200Cخواهم"))
		assert.Equal(t, "Büro 東京 Node", TypeUnicode("Büro 東京 Node"))
	})
	tests := []struct {
		name       string
		input      string
		want       string
		maxRunes64 bool
	}{
		{
			name:       "Clip",
			input:      " 幸福 Hanzi are logograms developed for the writing of Chinese! Expressions in an index may not ...!",
			want:       "幸福 Hanzi are logograms developed for the writing of Chinese! Exp",
			maxRunes64: true,
		},
		{
			name:  "WhitespaceCollapsed",
			input: "a b\tc\nd",
			want:  "a b c d",
		},
		{
			name:  "SpecialCharacters",
			input: "a-`~/\\:|\"'?*<>{}b",
			want:  "a-~/:b",
		},
		{
			name:  "NonASCII",
			input: "äöü",
			want:  "äöü",
		},
		{
			name:  "About",
			input: "PhotoPrism® Pro",
			want:  "PhotoPrism® Pro",
		},
		{
			name:  "Empty",
			input: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TypeUnicode(tt.input)
			assert.Equal(t, tt.want, got)
			if tt.maxRunes64 {
				assert.LessOrEqual(t, len([]rune(got)), LengthType)
			}
		})
	}
}

func TestTypeLower(t *testing.T) {
	t.Run("Clip", func(t *testing.T) {
		result := TypeLower(" 幸福 Hanzi are logograms developed for the writing of Chinese! Expressions in an index may not ...!")
		assert.Equal(t, "hanzi are logograms developed for the writing of chinese! expres", result) // codespell:ignore
		assert.Equal(t, LengthType, len(result))
	})
	t.Run("Ollama", func(t *testing.T) {
		assert.Equal(t, "redule26/huihui_ai_qwen2.5-vl-7b-abliterated:latest", TypeLower("redule26/huihui_ai_qwen2.5-vl-7b-abliterated:latest"))
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
	t.Run("ControlChars", func(t *testing.T) {
		assert.Equal(t, "ab c", ShortType("\x00a\x1fb\tc\x7f"))
	})
	t.Run("FilterBeforeClip", func(t *testing.T) {
		assert.Equal(t, "image", ShortType(strings.Repeat("\x00", 8)+"image"))
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
