package clean

import (
	"regexp"
	"strings"

	"github.com/photoprism/photoprism/pkg/list"
	"github.com/photoprism/photoprism/pkg/txt"
)

var EmailRegexp = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

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
	if s == "" || reject(s, txt.ClipUserName) {
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

// Password returns the sanitized password string with trimmed whitespace.
func Password(s string) string {
	s = strings.TrimSpace(s)

	if s == "" || reject(s, txt.ClipPassword) {
		return ""
	}

	return s
}
