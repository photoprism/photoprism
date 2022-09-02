package acl

import (
	"strings"

	"github.com/photoprism/photoprism/pkg/clean"
)

type Role string
type Roles map[Role]Actions

const (
	RoleAdmin   Role = "admin"
	RoleEditor  Role = "editor"
	RoleViewer  Role = "viewer"
	RoleGuest   Role = "guest"
	RoleDefault Role = "*"
)

// String returns the type as string.
func (t Role) String() string {
	return clean.Role(string(t))
}

// Equal checks if the type matches.
func (t Role) Equal(s string) bool {
	return strings.EqualFold(s, t.String())
}

// NotEqual checks if the type is different.
func (t Role) NotEqual(s string) bool {
	return !t.Equal(s)
}
