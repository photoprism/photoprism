package authn

import (
	"strings"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/txt"
)

// KeyType represents a multi-factor authentication key type.
type KeyType string

// Multi-factor authentication key types.
const (
	KeyTOTP    KeyType = "totp"
	KeyUnknown KeyType = ""
)

// String returns the authentication key type as a string.
func (t KeyType) String() string {
	switch t {
	case "totp", "otp":
		return string(KeyTOTP)
	default:
		return string(t)
	}
}

// Equal checks if the type matches.
func (t KeyType) Equal(s string) bool {
	return strings.EqualFold(s, t.String())
}

// NotEqual checks if the type is different.
func (t KeyType) NotEqual(s string) bool {
	return !t.Equal(s)
}

// Pretty returns the authentication key type in an easy-to-read format.
func (t KeyType) Pretty() string {
	switch t {
	case KeyTOTP:
		return "TOTP"
	case KeyUnknown:
		return "Unknown"
	default:
		return txt.UpperFirst(t.String())
	}
}

// Key casts a string to a normalized authentication key type.
func Key(s string) KeyType {
	s = clean.TypeLower(s)
	switch s {
	case "", "otp", "totp", "2fa":
		return KeyTOTP
	case "unknown", "null", "nil", "0", "false":
		return KeyUnknown
	default:
		return KeyType(s)
	}
}
