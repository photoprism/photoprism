package clean

import (
	"errors"
	"fmt"
	"maps"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

// sanitizers are the renderings that must not emit an unsafe rune, whatever they are given.
var sanitizers = map[string]func(string) string{
	"Log":       Log,
	"LogQuote":  LogQuote,
	"LogLower":  LogLower,
	"LogNames":  func(s string) string { return LogNames([]string{s}) },
	"Error":     func(s string) string { return Error(errors.New(s)) },
	"ErrorFull": func(s string) string { return ErrorFull(errors.New(s)) },
}

// renderers are the sanitizers plus Display, which keeps spaces and may return an empty string.
var renderers = func() map[string]func(string) string {
	m := map[string]func(string) string{"Display": Display}

	for name, fn := range sanitizers {
		m[name] = fn
	}

	return m
}()

// endsLine reports whether a rune terminates a line for a pager, an editor or CSS white-space.
// The assertions hold to this property; the enumeration in unsafe.go is one way of satisfying it.
func endsLine(r rune) bool {
	switch r {
	case '\n', '\r', '\v', '\f', 0x85, 0x2028, 0x2029:
		return true
	}

	return false
}

// reordersText reports whether a rune changes the direction or order of the text after it.
func reordersText(r rune) bool {
	switch {
	case r == 0x061C, r == 0x200E, r == 0x200F:
		return true
	case r >= 0x202A && r <= 0x202E:
		return true
	case r >= 0x2066 && r <= 0x2069:
		return true
	}

	return false
}

// drivesTerminal reports whether a rune is read as a command by a terminal rather than printed.
func drivesTerminal(r rune) bool {
	return r < 0x20 || r == 0x7F || (r >= 0x80 && r <= 0x9F)
}

// eachRune runs fn for every rune the sanitizers could be asked to render. The whole range,
// because the categories the classifier rejects reach well above the BMP.
func eachRune(fn func(r rune)) {
	for r := rune(0); r <= unicode.MaxRune; r++ {
		if utf8.ValidRune(r) {
			fn(r)
		}
	}
}

// eachInput returns the renderings of r a sanitizer is asked for. The bare form matters on its
// own: a value built only from removed characters empties, and callers index the first byte.
func eachInput(r rune) []string {
	return []string{"a" + string(r) + "b", string(r)}
}

func TestSanitizersEmitNoLineTerminator(t *testing.T) {
	for name, sanitize := range renderers {
		t.Run(name, func(t *testing.T) {
			eachRune(func(r rune) {
				for _, in := range eachInput(r) {
					for _, out := range sanitize(in) {
						if endsLine(out) {
							t.Fatalf("input U+%04X produced the line terminator U+%04X", r, out)
						}
					}
				}
			})
		})
	}
}

func TestSanitizersEmitNoTerminalControl(t *testing.T) {
	for name, sanitize := range renderers {
		t.Run(name, func(t *testing.T) {
			eachRune(func(r rune) {
				for _, in := range eachInput(r) {
					for _, out := range sanitize(in) {
						if drivesTerminal(out) {
							t.Fatalf("input U+%04X produced the control U+%04X", r, out)
						}
					}
				}
			})
		})
	}
}

func TestSanitizersEmitNoBidiControl(t *testing.T) {
	for name, sanitize := range renderers {
		t.Run(name, func(t *testing.T) {
			eachRune(func(r rune) {
				for _, in := range eachInput(r) {
					for _, out := range sanitize(in) {
						if reordersText(out) {
							t.Fatalf("input U+%04X produced the bidi control U+%04X", r, out)
						}
					}
				}
			})
		})
	}
}

func TestSanitizersKeepLetterFormingJoiners(t *testing.T) {
	// U+200C and U+200D are the two the enumeration deliberately admits: the first separates
	// letters in Persian and several Indic scripts, the second joins an emoji sequence.
	keep := map[string]rune{"ZWNJ": 0x200C, "ZWJ": 0x200D}

	for name, sanitize := range renderers {
		for label, r := range keep {
			t.Run(name+"/"+label, func(t *testing.T) {
				assert.Contains(t, sanitize("a"+string(r)+"b"), string(r))
			})
		}
	}
}

func TestSanitizersKeepEmojiSequences(t *testing.T) {
	// Escaped where invisible: variation selectors, modifiers and joiners must pass every helper.
	emoji := map[string]string{
		"EmojiPresentation": "❤\uFE0F",
		"TextPresentation":  "☺\uFE0E",
		"SkinTone":          "👍\U0001F3FD",
		"Keycap":            "1\uFE0F\u20E3",
		"CountryFlag":       "\U0001F1E9\U0001F1EA",
		"Astronaut":         "🧑\u200D🚀",
		"AstronautSkinTone": "🧑\U0001F3FD\u200D🚀",
		"RainbowFlag":       "🏳\uFE0F\u200D🌈",
		"Kiss":              "👩\u200D❤\uFE0F\u200D💋\u200D👨",
	}

	helpers := map[string]func(string) string{
		"AttachmentName": AttachmentName,
		"DisplayLines":   DisplayLines,
		"Name":           Name,
		"SearchBindable": SearchBindable,
		"StripControl":   StripControl,
		"Text":           Text,
		"TypeUnicode":    TypeUnicode,
	}

	maps.Copy(helpers, renderers)

	for name, sanitize := range helpers {
		for label, s := range emoji {
			t.Run(name+"/"+label, func(t *testing.T) {
				assert.Contains(t, sanitize("a "+s+" b"), s)
			})
		}
	}
}

func TestSanitizersKeepOrdinaryText(t *testing.T) {
	// Nothing outside the enumerated set may be dropped, or ordinary names become unreadable.
	// Lowercase throughout, because LogLower converts case by design.
	keep := []string{
		"ünterlagen.jpg",
		"日本語のファイル.png",
		"δοκιμή.tiff",
		"صورة.heic",
		"family-2019.mp4",
	}

	for name, sanitize := range renderers {
		for _, s := range keep {
			t.Run(name+"/"+s, func(t *testing.T) {
				assert.Contains(t, sanitize(s), s)
			})
		}
	}
}

func TestSanitizersNormalizeLookalikeSpaces(t *testing.T) {
	// Selected from the standard library rather than from unsafeSpaceRune, so a set that
	// shrinks is caught here instead of silently narrowing what the test looks at.
	for name, sanitize := range sanitizers {
		t.Run(name, func(t *testing.T) {
			eachRune(func(r rune) {
				if r == ' ' || !unicode.Is(unicode.Zs, r) {
					return
				}
				out := sanitize("a" + string(r) + "b")
				assert.NotContains(t, out, string(r), "U+%04X survived", r)
				assert.Contains(t, out, "a b", "U+%04X was dropped instead of normalized", r)
			})
		})
	}
}

func TestLogQuotesLookalikeSpaces(t *testing.T) {
	// The whole output, because asserting only that "a b" appears is satisfied by a rendering
	// that normalized the character and then failed to quote the value.
	eachRune(func(r rune) {
		if r == ' ' || !unicode.Is(unicode.Zs, r) {
			return
		}
		assert.Equal(t, "'John Smith'", Log("John"+string(r)+"Smith"), "U+%04X", r)
		assert.Equal(t, "'John Smith'", LogQuote("John"+string(r)+"Smith"), "U+%04X", r)
	})
}

// hidesText reports whether a rune is invisible where a reader expects a glyph. Selected from
// the standard library's categories rather than from unsafeRune, so a classifier that stops
// rejecting a category is caught instead of narrowing what the test looks at.
func hidesText(r rune) bool {
	if r == 0x200C || r == 0x200D {
		return false
	}

	return unicode.Is(unicode.Cf, r) || unicode.Is(unicode.Co, r) || unicode.Is(unicode.Cn, r)
}

func TestSanitizersRemoveEveryHiddenRune(t *testing.T) {
	// The three properties above do not cover the invisible characters, which end no line,
	// drive no terminal and reorder nothing.
	for name, sanitize := range renderers {
		t.Run(name, func(t *testing.T) {
			eachRune(func(r rune) {
				if !hidesText(r) {
					return
				}
				for _, in := range eachInput(r) {
					assert.NotContains(t, sanitize(in), string(r), "U+%04X survived", r)
				}
			})
		})
	}
}

func TestLogDoesNotReassembleRefusedTokens(t *testing.T) {
	assert.Equal(t, "?", Log("ldap://evil/a"))
	// A marked character cannot join what it separated.
	assert.NotContains(t, Log("l\x00dap://evil/a"), "ldap://")
	assert.NotContains(t, Log("l\x01dap://evil/a"), "ldap://")
	assert.NotContains(t, Log("l\u200bdap://evil/a"), "ldap://")
	assert.NotContains(t, Log("$\u200b{jndi:x}"), "${")
}

func TestLogMarksRemovedCharacters(t *testing.T) {
	// Dropping an invisible character silently makes the result identical to a benign name, so
	// a reader cannot tell the two apart and acts on the wrong file.
	for _, r := range []rune{0x00AD, 0x200B, 0x202E, 0x2060, 0xFEFF, 0xE0041} {
		got := Log("rep" + string(r) + "ort.jpg")
		assert.Equal(t, fmt.Sprintf("rep[U+%04X]ort.jpg", r), got, "U+%04X", r)
	}
}

func TestSanitizersNeverReturnEmpty(t *testing.T) {
	// Callers index the first byte of a sanitized value, and a caller rendering an error needs
	// a cause to print.
	for name, sanitize := range sanitizers {
		t.Run(name, func(t *testing.T) {
			eachRune(func(r rune) {
				assert.NotEmpty(t, sanitize(string(r)), "U+%04X emptied the result", r)
			})
		})
	}
}

func TestUnsafeRune(t *testing.T) {
	t.Run("Unsafe", func(t *testing.T) {
		for _, r := range []rune{0x00, 0x1F, 0x7F, 0x80, 0x85, 0x9B, 0x9F, 0x2028, 0x2029,
			0x061C, 0x200E, 0x200F, 0x202A, 0x202E, 0x2066, 0x2069, 0x00AD, 0x180E, 0x200B,
			0x2060, 0x206A, 0x206F, 0xFEFF, 0xFFF9, 0x0600, 0x110BD, 0x1D173, 0xE0001,
			0xE0041, 0xE007F, 0xF8FF, 0x0378} {
			assert.True(t, unsafeRune(r), "U+%04X", r)
		}
	})
	t.Run("Safe", func(t *testing.T) {
		for _, r := range []rune{' ', 'a', '~', 0x200C, 0x200D, 0x00E4, 0x4E2D,
			0x0301, 0x2800, 0x1F1E6, 0x1F3F4, 0x1F600, 0xFE0E, 0xFE0F, 0x20E3, 0x1F3FB,
			0x1F3FF, 0xE0100, 0xE01EF} {
			assert.False(t, unsafeRune(r), "U+%04X", r)
		}
	})
	t.Run("SpaceTakesPrecedence", func(t *testing.T) {
		// A Zs character is not printable either, so the order of the two decides whether it
		// separates words or is marked.
		for _, r := range []rune{0x00A0, 0x1680, 0x2000, 0x202F, 0x3000} {
			assert.True(t, unsafeRune(r), "U+%04X", r)
			assert.True(t, unsafeSpaceRune(r), "U+%04X", r)
			assert.Equal(t, "'a b'", Log("a"+string(r)+"b"), "U+%04X", r)
		}
	})
}

func TestUnsafeDropRune(t *testing.T) {
	t.Run("Dropped", func(t *testing.T) {
		for _, r := range []rune{0x00, 0x09, 0x0A, 0x0D, 0x1F, 0x7F} {
			assert.True(t, unsafeDropRune(r), "U+%04X", r)
		}
	})
	t.Run("NotDropped", func(t *testing.T) {
		// Everything the classifier newly rejects is marked instead, so a removal is visible.
		for _, r := range []rune{' ', 'a', 0x80, 0x85, 0x9B, 0x00AD, 0x200B, 0x202E, 0xFEFF} {
			assert.False(t, unsafeDropRune(r), "U+%04X", r)
		}
	})
}

func TestUnsafeSpaceRune(t *testing.T) {
	t.Run("LooksLikeSpace", func(t *testing.T) {
		for _, r := range []rune{0x00A0, 0x2000, 0x200A, 0x202F, 0x205F, 0x3000} {
			assert.True(t, unsafeSpaceRune(r), "U+%04X", r)
		}
	})
	t.Run("IsNotSpace", func(t *testing.T) {
		for _, r := range []rune{' ', 'a', 0x0009, 0x000A, 0x0085, 0x2028, 0x2029, 0x200B,
			0x200C, 0x200D, 0xFEFF} {
			assert.False(t, unsafeSpaceRune(r), "U+%04X", r)
		}
	})
}

func TestErrorSanitizesWrappedCause(t *testing.T) {
	// Escaped rather than literal: a raw bidi control in this file would reorder the source a
	// reader sees, which is the same effect the assertion is about.
	inner := errors.New("inner\u202Egpj.exe")
	err := fmt.Errorf("outer\u0085next: %w", inner)

	for _, out := range Error(err) {
		assert.False(t, endsLine(out) || reordersText(out) || drivesTerminal(out),
			"U+%04X survived the chain", out)
	}
}

func TestUnsafeBreakRune(t *testing.T) {
	for _, r := range []rune{'\t', '\n', '\v', '\f', '\r', 0x85, 0x2028, 0x2029} {
		assert.True(t, unsafeBreakRune(r), "U+%04X", r)
	}
	for _, r := range []rune{' ', 'a', 0x00, 0x1B, 0x7F, 0x9B, 0x202E} {
		assert.False(t, unsafeBreakRune(r), "U+%04X", r)
	}
}

func TestMarkText(t *testing.T) {
	// upper marks lowercase letters and replaces digits with a hash.
	upper := func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
			return markRune
		case r >= '0' && r <= '9':
			return '#'
		}
		return r
	}

	t.Run("Unchanged", func(t *testing.T) {
		assert.Equal(t, "", markText("", upper))
		assert.Equal(t, "ABC DEF", markText("ABC DEF", upper))
		assert.Equal(t, "رحلة", markText("رحلة", upper))
	})
	t.Run("Mapped", func(t *testing.T) {
		assert.Equal(t, "A[U+0062]C#", markText("AbC7", upper))
	})
	t.Run("InvalidBytes", func(t *testing.T) {
		assert.Equal(t, "A[0xFF]B[0xC3]", markText("A\xffB\xc3", upper))
		assert.Equal(t, "\uFFFD", markText("\uFFFD", upper))
	})
}

func TestWriteMarker(t *testing.T) {
	for r, expected := range map[rune]string{
		0x00:     "[U+0000]",
		0x7F:     "[U+007F]",
		0x202E:   "[U+202E]",
		0xFFFF:   "[U+FFFF]",
		0x1F600:  "[U+1F600]",
		0xE0041:  "[U+E0041]",
		0x10FFFD: "[U+10FFFD]",
	} {
		var b strings.Builder
		writeMarker(&b, r)
		assert.Equal(t, expected, b.String())
		assert.Equal(t, len(expected), markerLen(expected+"x"))
	}
}

func TestMarkerLen(t *testing.T) {
	t.Run("Markers", func(t *testing.T) {
		assert.Equal(t, 8, markerLen("[U+202E]abc"))
		assert.Equal(t, 9, markerLen("[U+1F600]"))
		assert.Equal(t, 10, markerLen("[U+10FFFD]"))
		assert.Equal(t, 6, markerLen("[0x9B]"))
	})
	t.Run("NoMarker", func(t *testing.T) {
		for _, s := range []string{"", "abc", "[U+20]", "[U+202]", "[U+202e]", "[U+1234567]", "[U+202E", "[0x9]", "[0x9b]", "[::1]", "U+202E"} {
			assert.Equal(t, 0, markerLen(s), s)
		}
	})
}

func TestMarkedUnitLen(t *testing.T) {
	assert.Equal(t, 8, markedUnitLen("[U+202E]b"))
	assert.Equal(t, 1, markedUnitLen("[::1]"))
	assert.Equal(t, 2, markedUnitLen("رحلة"))
	assert.Equal(t, 1, markedUnitLen("\xff"))
	assert.Equal(t, 0, markedUnitLen(""))
}
