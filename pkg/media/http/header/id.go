package header

import "strings"

// ID sanitizes identifier tokens, for example, a session ID, a UUID, or some other string ID.
func ID(s string) string {
	if s == "" || len(s) > 4096 {
		return ""
	}

	var prev rune

	// Remove all invalid characters.
	s = strings.Map(func(r rune) rune {
		if r == ' ' && (prev == 0 || prev == ' ') {
			return -1
		}

		prev = r

		switch r {
		case ' ', '"', '-', '+', '/', '=', '#', '$', '@', ':', ';', '_':
			return r
		}

		if (r < '0' || r > '9') && (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') {
			return -1
		}

		return r
	}, s)

	return s
}
