package clean

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/photoprism/photoprism/pkg/txt/clip"
)

const (
	// FieldSep is the character an event message separates its fields with. A value holding one
	// is quoted, so a single value cannot read as several fields.
	FieldSep = '\u203a'
	// LogNamesLimit is the number of names LogNames renders before it counts the rest.
	LogNamesLimit = 10
	// LogNamesBytes is the budget LogNames gives each name it renders.
	LogNamesBytes = 128
)

// LogNames sanitizes a list of names for logging and counts those beyond LogNamesLimit, so that
// the length of the message follows the two limits and not the length of the list. Each name is
// bounded in bytes rather than characters, which keeps a multi-byte name inside the same budget.
func LogNames(names []string) string {
	if len(names) == 0 {
		return "''"
	}

	kept := min(len(names), LogNamesLimit)
	out := make([]string, 0, kept)

	for _, name := range names[:kept] {
		out = append(out, logName(name))
	}

	if omitted := len(names) - kept; omitted > 0 {
		return fmt.Sprintf("%s and %d more", strings.Join(out, ", "), omitted)
	}

	return strings.Join(out, ", ")
}

// logName sanitizes a name like LogBytes with LogNamesBytes, shortening a long one in the middle so that
// its end, such as a file extension, stays visible.
func logName(s string) string {
	// The whole name is checked, since only its ends are mapped. They render far more than the result
	// shows, so the cut between them is never visible.
	if reject(s, 0) {
		return "?"
	}

	s, quote := logMap(clip.Ends(s, 16*LogNamesBytes))

	if !quote {
		return ClipMiddle(s, LogNamesBytes)
	}

	// Shortened before the apostrophes are doubled, so a cut cannot split a doubled one, with the largest
	// budget whose doubled result fits.
	budget := LogNamesBytes - 2

	if text := strings.ReplaceAll(s, "'", "''"); len(text) <= budget {
		return "'" + text + "'"
	}

	for maxBytes := budget; maxBytes > 0; maxBytes-- {
		if text := strings.ReplaceAll(ClipMiddle(s, maxBytes), "'", "''"); len(text) <= budget {
			return "'" + text + "'"
		}
	}

	return "''"
}

// ClipMiddle shortens sanitized text to at most maxBytes like clip.Middle, without splitting a marker, so
// that the end of a stored message, such as its outcome, stays visible.
func ClipMiddle(s string, maxBytes int) string {
	return clip.MiddleFunc(s, maxBytes, 0, markedUnitLen)
}

// ClipRunes shortens sanitized text to at most size characters like txt.Clip, without splitting a marker,
// for a column whose limit is measured in characters. Leading and trailing spaces are removed.
func ClipRunes(s string, size int) string {
	s = strings.TrimSpace(s)
	cut, runes := 0, 0

	for cut < len(s) {
		// A marker is ASCII, so its length in bytes is its length in characters.
		n, chars := markerLen(s[cut:]), 0

		if n > 0 {
			chars = n
		} else {
			_, n = utf8.DecodeRuneInString(s[cut:])
			chars = 1
		}

		if runes+chars > size {
			break
		}

		runes += chars
		cut += n
	}

	return strings.TrimSpace(s[:cut])
}

// logQuoted wraps a value in the quotes that show where it begins and ends, doubling any it
// already contains so that the value cannot close the pair and continue outside it.
func logQuoted(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

// Log sanitizes strings created from user input in response to the log4j debacle.
func Log(s string) string {
	return LogBytes(s, LengthLog)
}

// LogBytes sanitizes a value like Log, with the result bounded to maxBytes.
func LogBytes(s string, maxBytes int) string {
	s, quote := logText(s, maxBytes)

	if quote {
		return logQuotedBytes(s, maxBytes)
	}

	return s
}

// logQuotedBytes quotes a sanitized value, shortening it first so the doubled quotes still fit maxBytes.
// A budget without room for any text renders the empty pair.
func logQuotedBytes(s string, maxBytes int) string {
	if len(s)+strings.Count(s, "'")+2 <= maxBytes {
		return logQuoted(s)
	}

	budget := maxBytes - 2 - len(clip.Ellipsis)

	if budget < 0 {
		return logQuoted("")
	}

	used, cut := 0, 0

	for cut < len(s) {
		w := markedUnitLen(s[cut:])
		n := w

		if s[cut] == '\'' {
			n = 2
		}

		if used+n > budget {
			break
		}

		used += n
		cut += w
	}

	return logQuoted(s[:cut] + clip.Ellipsis)
}

// shortenBytes shortens s to at most maxBytes without splitting a character or marker, ending it with suffix
// if it cuts.
func shortenBytes(s string, maxBytes int, suffix string) string {
	if len(s) <= maxBytes {
		return s
	}

	limit := maxBytes - len(suffix)

	if limit < 0 {
		return ""
	}

	cut := 0

	for cut < len(s) {
		n := markedUnitLen(s[cut:])

		if cut+n > limit {
			break
		}

		cut += n
	}

	return s[:cut] + suffix
}

// logText sanitizes a value and reports whether it has to be quoted for its extent to be clear.
// The result has at most maxBytes, since a marker is longer than the character it replaces.
func logText(s string, maxBytes int) (string, bool) {
	if s == "" {
		return "", true
	}

	if s = shortenBytes(s, maxBytes, clip.Ellipsis); reject(s, LengthLimit) {
		return "?", false
	}

	s, quote := logMap(s)

	return shortenBytes(s, maxBytes, clip.Ellipsis), quote
}

// logMap replaces the characters of a value like logText and reports whether it has to be quoted, without
// shortening or checking it.
func logMap(s string) (string, bool) {
	if s == "" {
		return "", true
	}

	quote := false

	// Replace non-printable and other potentially problematic characters.
	s = markText(s, func(r rune) rune {
		switch {
		case r == ' ' || unsafeSpaceRune(r) || unsafeBreakRune(r):
			quote = true
			return unsafeSpace
		case r == FieldSep:
			quote = true
			return r
		case unsafeRune(r):
			return markRune
		}

		switch r {
		case '`':
			return '\''
		case '\\', '$', '<', '>', '{', '}':
			return '?'
		default:
			return r
		}
	})

	return s, quote
}

// LogQuote sanitizes a string and puts it in single quotes for logging. It quotes whether or not
// the value needs it, so that a caller placing several values side by side can tell them apart.
func LogQuote(s string) string {
	s, _ = logText(s, LengthLog)

	return logQuotedBytes(s, LengthLog)
}

// LogLower sanitizes strings created from user input and converts them to lowercase.
func LogLower(s string) string {
	return Log(strings.ToLower(s))
}
