package sanitize

import (
	"fmt"
	"strings"
	"unicode"
)

// Log sanitizes strings created from user input in response to the log4j debacle.
func Log(s string) string {
	if reject(s, 512) {
		return "?"
	}

	// Trim quotes, tabs, and newline characters.
	s = strings.Trim(s, " '\"“`\t\n\r")

	// Remove non-printable and other potentially problematic characters.
	s = strings.Map(func(r rune) rune {
		if !unicode.IsPrint(r) {
			return -1
		}

		switch r {
		case '`', '"':
			return '\''
		case '\\', '$', '<', '>', '{', '}':
			return '?'
		default:
			return r
		}
	}, s)

	// Empty?
	if s == "" || strings.ContainsAny(s, " ") {
		return fmt.Sprintf("'%s'", s)
	}

	return s
}

// LogLower sanitizes strings created from user input and converts them to lowercase.
func LogLower(s string) string {
	return Log(strings.ToLower(s))
}
