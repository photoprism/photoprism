package clean

import (
	"fmt"
	"strings"

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
		out = append(out, Log(clip.Bytes(name, LogNamesBytes)))
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
	s, quote := logText(s)

	if quote {
		return logQuoted(s)
	}

	return s
}

// logText sanitizes a value and reports whether it has to be quoted for its extent to be clear.
func logText(s string) (string, bool) {
	if s == "" {
		return "", true
	}

	s = clip.Shorten(s, LengthLog, clip.Ellipsis)

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

	return s, quote
}

// LogQuote sanitizes a string and puts it in single quotes for logging. It quotes whether or not
// the value needs it, so that a caller placing several values side by side can tell them apart.
func LogQuote(s string) string {
	s, _ = logText(s)

	return logQuoted(s)
}

// LogLower sanitizes strings created from user input and converts them to lowercase.
func LogLower(s string) string {
	return Log(strings.ToLower(s))
}
