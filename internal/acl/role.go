package acl

import (
	"strings"

	"github.com/photoprism/photoprism/pkg/txt"
)

// Role represents a user role.
type Role string

// String returns the type as string.
func (r Role) String() string {
	return string(r)
}

// Pretty returns the type in an easy-to-read format.
func (r Role) Pretty() string {
	if r == RoleNone {
		return "None"
	}

	return txt.UpperFirst(string(r))
}

// LogId returns an identifier string for use in log messages.
func (r Role) LogId() string {
	return "role " + r.String()
}

// Equal checks if the type matches.
func (r Role) Equal(s string) bool {
	return strings.EqualFold(s, r.String())
}

// NotEqual checks if the type is different.
func (r Role) NotEqual(s string) bool {
	return !r.Equal(s)
}

// Valid checks if the role is valid.
func (r Role) Valid(s string) bool {
	return UserRoles[s] != ""
}

// Invalid checks if the role is invalid.
func (r Role) Invalid(s string) bool {
	return !r.Valid(s)
}
