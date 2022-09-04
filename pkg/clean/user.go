package clean

import (
	"regexp"
	"strings"

	"github.com/photoprism/photoprism/pkg/txt"
)

var EmailRegexp = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

// Login returns the username of an account in lowercase, with spaces being trimmed.
func Login(s string) string {
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
	if s == "" || reject(s, txt.ClipUserName) {
		return ""
	}

	return strings.ToLower(s)
}

// Email returns the username of an account in lowercase, with spaces being trimmed.
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

// Role returns the password with spaces being trimmed.
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

// Password returns the password with spaces being trimmed.
func Password(s string) string {
	s = strings.TrimSpace(s)

	if s == "" || reject(s, txt.ClipPassword) {
		return ""
	}

	return s
}
