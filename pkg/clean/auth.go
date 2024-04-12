package clean

import (
	"regexp"
	"strings"

	"github.com/photoprism/photoprism/pkg/list"
	"github.com/photoprism/photoprism/pkg/txt"
)

var EmailRegexp = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

// Auth returns the sanitized authentication identifier trimmed to a maximum length of 255 characters.
func Auth(s string) string {
	if s == "" || len(s) > 2048 {
		return ""
	}

	i := 0

	// Remove unwanted characters and limit string length
	s = strings.Map(func(r rune) rune {
		if i == 0 && r == 32 || r < 32 || r == 127 {
			return -1
		}

		switch r {
		case '<', '>':
			return -1
		}

		i++

		if i > 255 {
			return -1
		}

		return r
	}, s)

	s = strings.TrimRight(s, " ")

	return s
}

// Handle returns the sanitized username with trimmed whitespace and in lowercase.
func Handle(s string) string {
	s, _, _ = strings.Cut(s, "@")

	if d, u, found := strings.Cut(s, "\\"); found && u != "" {
		s = u
	} else {
		s = d
	}

	s = strings.TrimSpace(s)

	// Remove unwanted characters.
	s = strings.Map(func(r rune) rune {
		if r <= 31 || r == 127 {
			return -1
		}
		switch r {
		case ' ', '"', '\'', '(', ')', '#', '&', '$', ',', '+', '=', '`', '~', '?', '|', '*', '/', '\\', ':', ';', '<', '>', '{', '}':
			return '.'
		}
		return r
	}, s)

	// Empty or too long?
	if s == "" || reject(s, txt.ClipUsername) {
		return ""
	}

	return strings.ToLower(s)
}

// Username returns the sanitized distinguished name (Username) with trimmed whitespace and in lowercase.
func Username(s string) string {
	s = strings.TrimSpace(s)

	// Remove unwanted characters.
	s = strings.Map(func(r rune) rune {
		if r <= 31 || r == 127 {
			return -1
		}
		switch r {
		case '"', '\'', '(', ')', '#', '&', '$', ',', '+', '=', '`', '~', '?', '|', '*', '/', ':', ';', '<', '>', '{', '}':
			return -1
		}
		return r
	}, s)

	// Empty or too long?
	if s == "" || reject(s, txt.ClipEmail) {
		return ""
	}

	return strings.ToLower(s)
}

// Email returns the sanitized email with trimmed whitespace and in lowercase.
func Email(s string) string {
	// Empty or too long?
	if s == "" || reject(s, txt.ClipEmail) {
		return ""
	}

	s = strings.ToLower(strings.TrimSpace(s))

	if EmailRegexp.MatchString(s) {
		return s
	}

	return ""
}

// Role returns the sanitized role with trimmed whitespace and in lowercase.
func Role(s string) string {
	// Remove unwanted characters.
	s = strings.Map(func(r rune) rune {
		if r <= 42 || r == 127 {
			return -1
		}
		switch r {
		case '`', '~', '?', '|', '*', '\\', '%', '$', '@', ':', ';', '<', '>', '{', '}':
			return -1
		}
		return r
	}, s)

	// Empty or too long?
	if s == "" || reject(s, txt.ClipRole) {
		return ""
	}

	return strings.ToLower(s)
}

// Attr returns the sanitized attributes.
func Attr(s string) string {
	return list.ParseAttr(s).String()
}

// Password returns the password string with all leading and trailing white space removed.
func Password(s string) string {
	return strings.TrimSpace(s)
}

// Passcode sanitizes a passcode and returns it in lowercase with all whitespace removed.
func Passcode(s string) string {
	if s == "" || reject(s, txt.ClipPasscode) {
		return ""
	} else if s = strings.ToLower(strings.TrimSpace(s)); s == "" {
		return ""
	}

	// Remove unwanted characters.
	s = strings.Map(func(r rune) rune {
		if (r < '0' || r > '9') && (r < 'a' || r > 'z') {
			return -1
		}

		return r
	}, s)

	return s
}
