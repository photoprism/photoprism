package clean

import (
	"fmt"
	"strings"
)

// Log sanitizes strings created from user input in response to the log4j debacle.
func Log(s string) string {
	if s == "" {
		return "''"
	} else if reject(s, 512) {
		return "?"
	}

	spaces := false

	// Remove non-printable and other potentially problematic characters.
	s = strings.Map(func(r rune) rune {
		if r < 32 {
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

// LogLower sanitizes strings created from user input and converts them to lowercase.
func LogLower(s string) string {
	return Log(strings.ToLower(s))
}
