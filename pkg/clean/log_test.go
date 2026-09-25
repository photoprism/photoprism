package clean

import (
	"errors"
	"strings"
	"testing"

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
		assert.Equal(t, "'  The ?quick? ''''brown \"fox.   '", Log("  The <quick>\n\r ''brown \"fox. \t  "))
	})
	t.Run("LoremIpsum", func(t *testing.T) {
		assert.Equal(t, "'It is a long established fact that a reader will be distracted by the readable "+
			"content of a pagewhen looking at its layout. The point of using Lorem Ipsum is that it has a "+
			"more-or-less normal distribution of letters,as opposed to using ''Content here, content here'', making it "+
			"look like readable English.Many desktop publishing packages and web page editors now use Lorem Ipsum as "+
			"their default model text, and a search for''lorem ipsum'' will uncover many web sites still in their "+
			"infancy. Various versions…'", Log(clip.LoremIpsum))
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

func TestLogText(t *testing.T) {
	t.Run("Plain", func(t *testing.T) {
		s, quote := logText("filename.txt")
		assert.Equal(t, "filename.txt", s)
		assert.False(t, quote)
	})
	t.Run("Space", func(t *testing.T) {
		s, quote := logText("two words")
		assert.Equal(t, "two words", s)
		assert.True(t, quote)
	})
	t.Run("FieldSep", func(t *testing.T) {
		s, quote := logText("a" + string(FieldSep) + "b")
		assert.Equal(t, "a"+string(FieldSep)+"b", s)
		assert.True(t, quote)
	})
	t.Run("Empty", func(t *testing.T) {
		// Reported as needing quotes so that both callers render the empty pair.
		s, quote := logText("")
		assert.Equal(t, "", s)
		assert.True(t, quote)
	})
	t.Run("Rejected", func(t *testing.T) {
		// The marker stands for the whole value, so quoting it would misreport its extent.
		s, quote := logText("ldap://evil/a")
		assert.Equal(t, "?", s)
		assert.False(t, quote)
	})
}
