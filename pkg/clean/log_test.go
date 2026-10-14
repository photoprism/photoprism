package clean

import (
	"errors"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/txt/clip"
)

func TestLog(t *testing.T) {
	t.Run("TheQuickBrownFox", func(t *testing.T) {
		assert.Equal(t, "'The quick brown fox.'", Log("The quick brown fox."))
	})
	t.Run("FilenameTxt", func(t *testing.T) {
		assert.Equal(t, "filename.txt", Log("filename.txt"))
	})
	t.Run("EmptyString", func(t *testing.T) {
		assert.Equal(t, "''", Log(""))
	})
	t.Run("Replace", func(t *testing.T) {
		assert.Equal(t, "?", Log("${https://<host>:<port>/<path>}"))
	})
	t.Run("Ldap", func(t *testing.T) {
		assert.Equal(t, "?", Log("User-Agent: {jndi:ldap://<host>:<port>/<path>}"))
	})
	t.Run("SpecialChars", func(t *testing.T) {
		// The quotes delimit the value, so any it already holds are doubled: the two here
		// render as four, and the value cannot close the pair and continue outside it.
		assert.Equal(t, "'  The ?quick?   ''''brown \"fox.    '", Log("  The <quick>\n\r ''brown \"fox. \t  "))
	})
	t.Run("LoremIpsum", func(t *testing.T) {
		assert.Equal(t, "'It is a long established fact that a reader will be distracted by the readable "+
			"content of a page when looking at its layout. The point of using Lorem Ipsum is that it has a "+
			"more-or-less normal distribution of letters, as opposed to using ''Content here, content here'', making it "+
			"look like readable English. Many desktop publishing packages and web page editors now use Lorem Ipsum as "+
			"their default model text, and a search for ''lorem ipsum'' will uncover many web sites still in their "+
			"infancy. Various ver…'", Log(clip.LoremIpsum))
	})
}

func TestLogQuote(t *testing.T) {
	t.Run("TheQuickBrownFox", func(t *testing.T) {
		assert.Equal(t, "'The quick brown fox.'", LogQuote("The quick brown fox."))
	})
	t.Run("SpecialChars", func(t *testing.T) {
		assert.Equal(t, "'?The quick brown fox'", LogQuote("$The quick brown fox"))
	})
}

func TestLogLower(t *testing.T) {
	t.Run("TheQuickBrownFox", func(t *testing.T) {
		assert.Equal(t, "'the quick brown fox.'", LogLower("The quick brown fox."))
	})
	t.Run("FilenameTxt", func(t *testing.T) {
		assert.Equal(t, "filename.txt", LogLower("filename.TXT"))
	})
	t.Run("EmptyString", func(t *testing.T) {
		assert.Equal(t, "''", LogLower(""))
	})
	t.Run("Replace", func(t *testing.T) {
		assert.Equal(t, "?", LogLower("${https://<host>:<port>/<path>}"))
	})
	t.Run("Ldap", func(t *testing.T) {
		assert.Equal(t, "?", LogLower("User-Agent: ${jndi:ldap://<host>:<port>/<path>}"))
	})
}

func TestLogNames(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "''", LogNames(nil))
		assert.Equal(t, "''", LogNames([]string{}))
	})
	t.Run("One", func(t *testing.T) {
		assert.Equal(t, "filename.txt", LogNames([]string{"filename.txt"}))
	})
	t.Run("Several", func(t *testing.T) {
		assert.Equal(t, "a.jpg, b.jpg, c.jpg", LogNames([]string{"a.jpg", "b.jpg", "c.jpg"}))
	})
	t.Run("AtLimit", func(t *testing.T) {
		names := make([]string, LogNamesLimit)
		for i := range names {
			names[i] = "a.jpg"
		}
		result := LogNames(names)
		assert.NotContains(t, result, "more")
		assert.Equal(t, LogNamesLimit, strings.Count(result, "a.jpg"))
	})
	t.Run("BeyondLimitIsCounted", func(t *testing.T) {
		names := make([]string, LogNamesLimit+5)
		for i := range names {
			names[i] = "a.jpg"
		}
		result := LogNames(names)
		assert.Equal(t, LogNamesLimit, strings.Count(result, "a.jpg"))
		assert.Contains(t, result, "and 5 more")
	})
	t.Run("LengthFollowsTheLimits", func(t *testing.T) {
		// The rendered length must follow the two limits, not the length of the list or of a name.
		bound := LogNamesLimit*(LogNamesBytes+8) + 32
		for _, fill := range []string{"a", "\u00e4", "\u4e2d", "\U0001F600"} {
			names := make([]string, 10000)
			for i := range names {
				names[i] = strings.Repeat(fill, 600) + ".jpg"
			}
			assert.Less(t, len(LogNames(names)), bound, "fill %q", fill)
		}
	})
	t.Run("NameIsBoundedInBytes", func(t *testing.T) {
		// A multi-byte name gets the same budget as an ASCII one.
		for _, fill := range []string{"a", "\u4e2d", "\U0001F600"} {
			assert.LessOrEqual(t, len(LogNames([]string{strings.Repeat(fill, 600)})), LogNamesBytes+8, "fill %q", fill)
		}
	})
	t.Run("EachNameIsSanitized", func(t *testing.T) {
		result := LogNames([]string{"ok.jpg", "report\nsummary.jpg", "x\ry.jpg"})
		assert.NotContains(t, result, "\n")
		assert.NotContains(t, result, "\r")
		assert.Contains(t, result, "ok.jpg")
	})
	t.Run("InjectionIsRejected", func(t *testing.T) {
		assert.Equal(t, "?", LogNames([]string{"${jndi:ldap://host:1389/a}"}))
	})
}

func TestFieldSep(t *testing.T) {
	sep := string(FieldSep)

	t.Run("LogQuotes", func(t *testing.T) {
		// A value holding this character is quoted, so it cannot read as more than one field.
		assert.Equal(t, "'ok"+sep+"tail'", Log("ok"+sep+"tail"))
		assert.Equal(t, "'ok"+sep+"tail'", LogQuote("ok"+sep+"tail"))
	})
	t.Run("SpacedFormAlreadyQuoted", func(t *testing.T) {
		// The rendered separator carries spaces, so the space rule covers that form too.
		assert.Equal(t, "'ok "+sep+" tail'", Log("ok "+sep+" tail"))
	})
	t.Run("NotAUsername", func(t *testing.T) {
		// Asserted exactly: a rewrite to another character also satisfies NotContains, and the
		// two helpers dispose of it differently on purpose.
		assert.Equal(t, "oktail", Username("ok"+sep+"tail"))
		assert.Equal(t, "ok.tail", Handle("ok"+sep+"tail"))
	})
	t.Run("NotAnErrorMessage", func(t *testing.T) {
		// Asserted exactly: a message is a sentence, so it is folded rather than quoted, and the
		// three helpers dispose of the character in three different ways on purpose.
		assert.Equal(t, "read?failed", ErrorFull(errors.New("read"+sep+"failed")))
		assert.Equal(t, "read?failed", Error(errors.New("read"+sep+"failed")))
	})
}

func TestLogQuoteEscaping(t *testing.T) {
	sep := string(FieldSep)

	t.Run("ValueCannotCloseItsOwnQuote", func(t *testing.T) {
		// Doubling is what makes the pair delimit. Without it the value below renders with a
		// balanced pair around each half, so it reads as two values rather than one.
		assert.Equal(t, `'x'' `+sep+` ''tail'`, Log(`x' `+sep+` 'tail`))
	})
	t.Run("AlwaysBalanced", func(t *testing.T) {
		for _, in := range []string{`'tail`, `tail'`, `'`, `''`, `O'Brien`, "plain"} {
			got := LogQuote(in)

			require.GreaterOrEqual(t, len(got), 2, in)
			assert.Equal(t, byte('\''), got[0], in)
			assert.Equal(t, byte('\''), got[len(got)-1], in)
			// Every quote between the outer pair is doubled, so the count inside is even.
			assert.Zero(t, strings.Count(got[1:len(got)-1], `'`)%2, in)
		}
	})
	t.Run("KeepsAnOrdinaryApostrophe", func(t *testing.T) {
		// A value is not quoted when nothing requires it, so nothing is doubled either.
		assert.Equal(t, `O'Brien`, Log(`O'Brien`))
	})
}

func TestLogQuoted(t *testing.T) {
	t.Run("Wraps", func(t *testing.T) {
		assert.Equal(t, `'a'`, logQuoted("a"))
		assert.Equal(t, `''`, logQuoted(""))
	})
	t.Run("Doubles", func(t *testing.T) {
		assert.Equal(t, `'a''b'`, logQuoted(`a'b`))
		assert.Equal(t, `''''`, logQuoted(`'`))
	})
}

func TestLogBytes(t *testing.T) {
	t.Run("Short", func(t *testing.T) {
		assert.Equal(t, "photo.jpg", LogBytes("photo.jpg", 16))
	})
	t.Run("Replaced", func(t *testing.T) {
		// Each invalid byte becomes a 3-byte character, so the text is shortened after the replacement.
		result := LogBytes(strings.Repeat("\xff", 10), 16)
		assert.LessOrEqual(t, len(result), 16)
		assert.True(t, strings.HasSuffix(result, "…"))
	})
	t.Run("Quoted", func(t *testing.T) {
		assert.Equal(t, "'a b'", LogBytes("a b", 16))
	})
}

func TestLogQuotedBytes(t *testing.T) {
	// unquote removes the enclosing quotes and every doubled pair, leaving any quote the value could close the pair with.
	unquote := func(s string) string {
		return strings.ReplaceAll(s[1:len(s)-1], "''", "")
	}

	t.Run("Fits", func(t *testing.T) {
		assert.Equal(t, "'a b'", logQuotedBytes("a b", 16))
	})
	t.Run("Shortened", func(t *testing.T) {
		result := logQuotedBytes(strings.Repeat("a'b c'd", 10), 32)
		assert.LessOrEqual(t, len(result), 32)
		assert.Greater(t, len(result), 32-4)
		assert.True(t, strings.HasSuffix(result, "…'"))
		assert.NotContains(t, unquote(result), "'")
	})
	t.Run("OnlyQuotes", func(t *testing.T) {
		result := logQuotedBytes(strings.Repeat("'", 600), LengthLog)
		assert.LessOrEqual(t, len(result), LengthLog)
		assert.Greater(t, len(result), LengthLog-4)
		assert.NotContains(t, unquote(result), "'")
	})
	t.Run("TinyBudget", func(t *testing.T) {
		assert.Equal(t, "''", logQuotedBytes("abcdef", 4))
	})
}

func TestShortenBytes(t *testing.T) {
	t.Run("Fits", func(t *testing.T) {
		assert.Equal(t, "abc ", shortenBytes("abc ", 4, clip.Ellipsis))
	})
	t.Run("Suffix", func(t *testing.T) {
		assert.Equal(t, "abcd…", shortenBytes("abcdefgh", 7, clip.Ellipsis))
	})
	t.Run("NoSuffix", func(t *testing.T) {
		assert.Equal(t, "abcdefg", shortenBytes("abcdefgh", 7, ""))
	})
	t.Run("MultiByte", func(t *testing.T) {
		assert.Equal(t, "\u4e2d\u4e2d", shortenBytes(strings.Repeat("\u4e2d", 5), 8, ""))
	})
	t.Run("TinyBudget", func(t *testing.T) {
		assert.Equal(t, "", shortenBytes("abcdef", 2, clip.Ellipsis))
	})
}

func TestLogBounds(t *testing.T) {
	// Replacing an invalid byte or keeping a multi-byte character must not grow a value past its budget.
	for _, fill := range []string{"a", "\xff", "\u4e2d", "\U0001F600", "'", "a '"} {
		value := strings.Repeat(fill, 1000)
		assert.LessOrEqual(t, len(Log(value)), LengthLog, "fill %q", fill)
		assert.LessOrEqual(t, len(LogQuote(value)), LengthLog, "fill %q", fill)
		assert.True(t, utf8.ValidString(Log(value)), "fill %q", fill)
		assert.True(t, utf8.ValidString(LogQuote(value)), "fill %q", fill)
		assert.True(t, utf8.ValidString(LogNames([]string{value})), "fill %q", fill)
		assert.LessOrEqual(t, len(LogNames([]string{value})), LogNamesBytes, "fill %q", fill)
	}
}

func TestLogExactBudget(t *testing.T) {
	// A value that fits its budget exactly is not marked as shortened.
	assert.Equal(t, strings.Repeat("a", LogNamesBytes), LogNames([]string{strings.Repeat("a", LogNamesBytes)}))
	assert.Equal(t, strings.Repeat("a", LengthLog), Log(strings.Repeat("a", LengthLog)))
	assert.Equal(t, "'a b'", logQuotedBytes("a b", 5))
	assert.Equal(t, "'…'", logQuotedBytes("abcdef", 5))
}

func TestLogLongerThanLimit(t *testing.T) {
	// A value longer than the guard's limit is shortened rather than refused.
	value := strings.Repeat("a", LengthLimit+1)
	assert.Equal(t, strings.Repeat("a", LengthLog-len(clip.Ellipsis))+clip.Ellipsis, Log(value))
	budget := LogNamesBytes - len(clip.Ellipsis)
	assert.Equal(t, strings.Repeat("a", budget-budget/3)+clip.Ellipsis+strings.Repeat("a", budget/3), LogNames([]string{value}))
}

func TestLogText(t *testing.T) {
	t.Run("Plain", func(t *testing.T) {
		s, quote := logText("filename.txt", LengthLog)
		assert.Equal(t, "filename.txt", s)
		assert.False(t, quote)
	})
	t.Run("Space", func(t *testing.T) {
		s, quote := logText("two words", LengthLog)
		assert.Equal(t, "two words", s)
		assert.True(t, quote)
	})
	t.Run("FieldSep", func(t *testing.T) {
		s, quote := logText("a"+string(FieldSep)+"b", LengthLog)
		assert.Equal(t, "a"+string(FieldSep)+"b", s)
		assert.True(t, quote)
	})
	t.Run("Empty", func(t *testing.T) {
		// Reported as needing quotes so that both callers render the empty pair.
		s, quote := logText("", LengthLog)
		assert.Equal(t, "", s)
		assert.True(t, quote)
	})
	t.Run("Rejected", func(t *testing.T) {
		// The marker stands for the whole value, so quoting it would misreport its extent.
		s, quote := logText("ldap://evil/a", LengthLog)
		assert.Equal(t, "?", s)
		assert.False(t, quote)
	})
}

func TestLogMarkerBudget(t *testing.T) {
	// incomplete reports whether a value holds a part of a marker, since the values here contain no "[".
	incomplete := func(s string) bool {
		for i := range s {
			if s[i] == '[' && markerLen(s[i:]) == 0 {
				return true
			}
		}
		return false
	}

	for _, value := range []string{
		strings.Repeat("\u200B", 300),
		strings.Repeat("a\x07", 300),
		strings.Repeat("\xff", 600),
		strings.Repeat("a b\u2060", 200),
		" " + strings.Repeat("a", 501) + "\u200Bb",
		strings.Repeat("ab", 100) + strings.Repeat("\x01", 200) + ".jpg",
		strings.Repeat("ab", 100) + strings.Repeat("\x00", 200) + ".jpg",
	} {
		for s, bound := range map[string]int{
			Log(value):                 LengthLog,
			LogQuote(value):            LengthLog,
			LogBytes(value, 100):       100,
			LogNames([]string{value}):  LogNamesBytes,
			ClipMiddle(Log(value), 64): 64,
		} {
			assert.LessOrEqual(t, len(s), bound, "%q", s)
			assert.False(t, incomplete(s), "%q", s)
			assert.NotContains(t, s, clip.Ellipsis+clip.Ellipsis, "%q", s)
		}
	}
}

func TestLogName(t *testing.T) {
	t.Run("Short", func(t *testing.T) {
		assert.Equal(t, "IMG_1234.jpg", logName("IMG_1234.jpg"))
		assert.Equal(t, "'Summer 2026.jpg'", logName("Summer 2026.jpg"))
		assert.Equal(t, "''", logName(""))
	})
	t.Run("KeepsEnd", func(t *testing.T) {
		name := logName(strings.Repeat("vacation-", 20) + "IMG_1234.jpg")
		assert.LessOrEqual(t, len(name), LogNamesBytes)
		assert.True(t, strings.HasPrefix(name, "vacation-"))
		assert.True(t, strings.HasSuffix(name, "IMG_1234.jpg"))
		assert.Contains(t, name, clip.Ellipsis)
	})
	t.Run("Quoted", func(t *testing.T) {
		name := logName(strings.Repeat("it's ", 40) + "end.jpg")
		assert.LessOrEqual(t, len(name), LogNamesBytes)
		assert.True(t, strings.HasSuffix(name, "end.jpg'"))
	})
	t.Run("Apostrophes", func(t *testing.T) {
		name := logName(strings.Repeat("'", 130) + " x.jpg")
		assert.LessOrEqual(t, len(name), LogNamesBytes)
		assert.True(t, strings.HasSuffix(name, "x.jpg'"), name)
		// A doubled apostrophe is kept in one piece, so the quotes still pair up on both sides of the cut.
		for _, value := range []string{strings.Repeat("'", 130) + " x.jpg", "a b" + strings.Repeat("'", 300) + ".jpg", "a b" + strings.Repeat("'", 301)} {
			quoted := logName(value)
			assert.LessOrEqual(t, len(quoted), LogNamesBytes, quoted)
			for _, part := range strings.Split(quoted[1:len(quoted)-1], clip.Ellipsis) {
				assert.Equal(t, 0, strings.Count(part, "'")%2, quoted)
			}
		}
		assert.True(t, strings.HasSuffix(logName(strings.Repeat("'", 125)+" x"), " x'"))
	})
	t.Run("MarkedEnd", func(t *testing.T) {
		name := logName(strings.Repeat("ab", 100) + strings.Repeat("\x01", 20) + ".jpg")
		assert.LessOrEqual(t, len(name), LogNamesBytes)
		assert.True(t, strings.HasSuffix(name, "[U+0001].jpg"), name)
		assert.Equal(t, 1, strings.Count(name, clip.Ellipsis), name)
	})
	t.Run("WholeNameChecked", func(t *testing.T) {
		assert.Equal(t, "?", logName(strings.Repeat("a", 3000)+"${x}"+strings.Repeat("b", 3000)))
		name := logName(strings.Repeat("a", 1023) + "$" + strings.Repeat("b", 3000) + "{" + strings.Repeat("c", 1019) + ".jpg")
		assert.True(t, strings.HasSuffix(name, "c.jpg"), name)
	})
	t.Run("MarksNul", func(t *testing.T) {
		name := logName(strings.Repeat("\x00", 5000) + "IMG_1234.jpg")
		assert.LessOrEqual(t, len(name), LogNamesBytes)
		assert.True(t, strings.HasPrefix(name, "[U+0000]"), name)
		assert.True(t, strings.HasSuffix(name, "IMG_1234.jpg"), name)
		assert.Equal(t, "ld[U+0000]ap://host", logName("ld\x00ap://host"))
	})
	t.Run("Marked", func(t *testing.T) {
		assert.Equal(t, "rep[U+202E]ort.jpg", logName("rep\u202Eort.jpg"))
	})
}

func TestClipMiddle(t *testing.T) {
	assert.Equal(t, "short", ClipMiddle("short", 64))
	assert.Equal(t, "ab[U+202E]c…gh", ClipMiddle("ab[U+202E]cdef[U+200B]gh", 16))
	clipped := ClipMiddle("user admin › "+strings.Repeat("[U+200B]", 40)+" › failed", 64)
	assert.LessOrEqual(t, len(clipped), 64)
	assert.True(t, strings.HasPrefix(clipped, "user admin › "))
	assert.True(t, strings.HasSuffix(clipped, "› failed"))
}

func TestClipRunes(t *testing.T) {
	assert.Equal(t, "", ClipRunes("", 8))
	assert.Equal(t, "", ClipRunes("abc", 0))
	assert.Equal(t, "abc", ClipRunes("  abc  ", 8))
	assert.Equal(t, "رحلة", ClipRunes("رحلة 2026", 5))
	assert.Equal(t, "ab", ClipRunes("ab[U+200B]cd", 9))
	assert.Equal(t, "ab[U+200B]c", ClipRunes("ab[U+200B]cd", 11))
	assert.Equal(t, "[0xFF][0xFF]", ClipRunes("[0xFF][0xFF][0xFF]", 17))
}

func TestLogMap(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		s, quote := logMap("")
		assert.Equal(t, "", s)
		assert.True(t, quote)
	})
	t.Run("Plain", func(t *testing.T) {
		s, quote := logMap("IMG_1234.jpg")
		assert.Equal(t, "IMG_1234.jpg", s)
		assert.False(t, quote)
	})
	t.Run("Mapped", func(t *testing.T) {
		s, quote := logMap("a\tb`c<d>\x00e\u202E")
		assert.Equal(t, "a b'c?d?[U+0000]e[U+202E]", s)
		assert.True(t, quote)
	})
	t.Run("NotChecked", func(t *testing.T) {
		s, _ := logMap("ldap://host")
		assert.Equal(t, "ldap://host", s)
	})
}
