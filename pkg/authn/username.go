package authn

import (
	"errors"
	"strings"

	"github.com/gosimple/unidecode"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/txt"
)

var (
	// ErrEmpty indicates the username was not provided.
	ErrEmpty = errors.New("empty")
	// ErrTooLong indicates the username exceeds the allowed length.
	ErrTooLong = errors.New("too long")
	// ErrInvalid indicates the username failed basic validation.
	ErrInvalid = errors.New("invalid")
	// ErrReserved indicates the username is reserved for system use.
	ErrReserved = errors.New("reserved")
	// ErrNotASCII indicates a new username contains characters other than ASCII.
	ErrNotASCII = errors.New("username may only contain ASCII characters")
)

// Username checks if the name provided is invalid or reserved.
func Username(name string) (sanitized string, err error) {
	if name == "" {
		return "", ErrEmpty
	} else if len(name) > txt.ClipEmail {
		return "", ErrTooLong
	}

	s := clean.Username(name)

	switch {
	case s == "" || !strings.EqualFold(s, name):
		return s, ErrInvalid
	case rnd.IsUID(s, 'c'):
		return s, ErrReserved
	}

	return s, err
}

// IsASCIIUsername reports whether name contains only ASCII characters, as new and renamed accounts require.
func IsASCIIUsername(name string) bool {
	return txt.IsASCII(name)
}

// NewUsername checks a name for a new or renamed account like Username and returns ErrNotASCII if it
// contains characters other than ASCII; sanitizing ASCII input yields ASCII. Existing accounts keep
// their names, so lookups use Username.
func NewUsername(name string) (sanitized string, err error) {
	if sanitized, err = Username(name); err != nil {
		return sanitized, err
	} else if !IsASCIIUsername(name) {
		return sanitized, ErrNotASCII
	}

	return sanitized, nil
}

// FoldUsername returns the ASCII transliteration of name if it is a valid new username, or an empty string.
// It only suggests a name; nothing stores it unless an operator chooses it.
func FoldUsername(name string) string {
	if name == "" || len(name) > txt.ClipEmail {
		return ""
	}

	if s, err := NewUsername(clean.Username(unidecode.Unidecode(clean.Username(name)))); err == nil {
		return s
	}

	return ""
}
