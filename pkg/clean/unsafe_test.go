package clean

import (
	"errors"
	"fmt"
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
	for name, sanitize := range sanitizers {
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
	for name, sanitize := range sanitizers {
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
	for name, sanitize := range sanitizers {
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

	for name, sanitize := range sanitizers {
		for label, r := range keep {
			t.Run(name+"/"+label, func(t *testing.T) {
				assert.Contains(t, sanitize("a"+string(r)+"b"), string(r))
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

	for name, sanitize := range sanitizers {
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
	for name, sanitize := range sanitizers {
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
	// A dropped character joins what it separated, which is why the guard runs again.
	assert.NotContains(t, Log("l\x01dap://evil/a"), "ldap://")
	// A marked one cannot join anything, so it never reaches that second guard.
	assert.NotContains(t, Log("l\u200bdap://evil/a"), "ldap://")
	assert.NotContains(t, Log("$\u200b{jndi:x}"), "${")
}

func TestLogMarksRemovedCharacters(t *testing.T) {
	// Dropping an invisible character silently makes the result identical to a benign name, so
	// a reader cannot tell the two apart and acts on the wrong file.
	for _, r := range []rune{0x00AD, 0x200B, 0x202E, 0x2060, 0xFEFF, 0xE0041} {
		got := Log("rep" + string(r) + "ort.jpg")
		assert.NotEqual(t, "report.jpg", got, "U+%04X", r)
		assert.Contains(t, got, string(unsafeMarker), "U+%04X", r)
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
			0x0301, 0x2800, 0x1F1E6, 0x1F3F4, 0x1F600} {
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
