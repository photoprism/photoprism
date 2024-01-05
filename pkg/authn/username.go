package authn

import (
	"errors"
	"strings"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/txt"
)

var (
	ErrEmpty    = errors.New("empty")
	ErrTooLong  = errors.New("too long")
	ErrInvalid  = errors.New("invalid")
	ErrReserved = errors.New("reserved")
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
	case rnd.IsUID(s, 'c'), s == "visitor":
		return s, ErrReserved
	}

	return s, err
}
