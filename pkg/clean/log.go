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
		out = append(out, logBytes(name, LogNamesBytes))
	}

	if omitted := len(names) - kept; omitted > 0 {
		return fmt.Sprintf("%s and %d more", strings.Join(out, ", "), omitted)
	}

	return strings.Join(out, ", ")
}

// logQuoted wraps a value in the quotes that show where it begins and ends, doubling any it
// already contains so that the value cannot close the pair and continue outside it.
func logQuoted(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

// Log sanitizes strings created from user input in response to the log4j debacle.
func Log(s string) string {
	return logBytes(s, LengthLog)
}

// logBytes sanitizes a value like Log, with the result bounded to maxBytes.
func logBytes(s string, maxBytes int) string {
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
		r, w := utf8.DecodeRuneInString(s[cut:])
		n := w

		if r == '\'' {
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

// shortenBytes shortens s to at most maxBytes without splitting a character, ending it with suffix if it cuts.
func shortenBytes(s string, maxBytes int, suffix string) string {
	if len(s) <= maxBytes {
		return s
	}

	limit := maxBytes - len(suffix)

	if limit < 0 {
		return ""
	}

	cut := 0

	for i := range s {
		if i > limit {
			break
		}

		cut = i
	}

	return s[:cut] + suffix
}

// logText sanitizes a value and reports whether it has to be quoted for its extent to be clear.
// The result has at most maxBytes, since replacing an invalid byte can triple its size.
func logText(s string, maxBytes int) (string, bool) {
	if s == "" {
		return "", true
	}

	s = shortenBytes(s, maxBytes, clip.Ellipsis)

	if reject(s, LengthLimit) {
		return "?", false
	}

	quote := false
	dropped := false

	// Remove non-printable and other potentially problematic characters.
	s = strings.Map(func(r rune) rune {
		switch {
		case r == ' ' || unsafeSpaceRune(r):
			quote = true
			return unsafeSpace
		case r == FieldSep:
			quote = true
			return r
		case unsafeDropRune(r):
			dropped = true
			return -1
		case unsafeRune(r):
			return unsafeMarker
		}

		switch r {
		case '`':
			return '\''
		case '"':
			return '"'
		case '\\', '$', '<', '>', '{', '}':
			return '?'
		default:
			return r
		}
	}, s)

	// A value of only dropped characters empties here, and both callers render the pair.
	if s == "" {
		return "", true
	}

	// Dropping a character joins what it separated, so re-check what the guard refused.
	if dropped && reject(s, LengthLimit) {
		return "?", false
	}

	return shortenBytes(s, maxBytes, clip.Ellipsis), quote
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
