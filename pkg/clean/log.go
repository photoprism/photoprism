package clean

import (
	"fmt"
	"strings"

	"github.com/photoprism/photoprism/pkg/txt/clip"
)

// Log sanitizes strings created from user input in response to the log4j debacle.
func Log(s string) string {
	if s == "" {
		return "''"
	}

	s = clip.Shorten(s, LengthLog, clip.Ellipsis)

	if reject(s, LengthLimit) {
		return "?"
	}

	spaces := false

	// Remove non-printable and other potentially problematic characters.
	s = strings.Map(func(r rune) rune {
		if r < 32 || r == 127 {
			return -1
		}

		switch r {
		case ' ':
			spaces = true
			return r
		case '`', '"':
			return '\''
		case '\\', '$', '<', '>', '{', '}':
			return '?'
		default:
			return r
		}
	}, s)

	// Contains spaces?
	if spaces {
		return fmt.Sprintf("'%s'", s)
	}

	return s
}

// LogQuote sanitizes a string and puts it in single quotes for logging.
func LogQuote(s string) string {
	if s = Log(s); s[0] != '\'' {
		return fmt.Sprintf("'%s'", s)
	} else {
		return s
	}
}

// LogLower sanitizes strings created from user input and converts them to lowercase.
func LogLower(s string) string {
	return Log(strings.ToLower(s))
}
