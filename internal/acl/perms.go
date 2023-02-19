package acl

import "strings"

// Permissions that can be granted to roles.
const (
	FullAccess      Permission = "full_access"
	AccessShared    Permission = "access_shared"
	AccessLibrary   Permission = "access_library"
	AccessPrivate   Permission = "access_private"
	AccessOwn       Permission = "access_own"
	AccessAll       Permission = "access_all"
	ActionSearch    Permission = "search"
	ActionView      Permission = "view"
	ActionUpload    Permission = "upload"
	ActionCreate    Permission = "create"
	ActionUpdate    Permission = "update"
	ActionDownload  Permission = "download"
	ActionShare     Permission = "share"
	ActionDelete    Permission = "delete"
	ActionRate      Permission = "rate"
	ActionReact     Permission = "react"
	ActionManage    Permission = "manage"
	ActionSubscribe Permission = "subscribe"
)

// Permission represents a single ability.
type Permission string

// String returns the type as string.
func (p Permission) String() string {
	return strings.ReplaceAll(string(p), "_", " ")
}

// LogId returns an identifier string for use in log messages.
func (p Permission) LogId() string {
	return p.String()
}

// Equal checks if the type matches.
func (p Permission) Equal(s string) bool {
	return strings.EqualFold(s, p.String())
}

// NotEqual checks if the type is different.
func (p Permission) NotEqual(s string) bool {
	return !p.Equal(s)
}

// Permissions is a list of permissions.
type Permissions []Permission

// String returns the permissions as a comma-separated string.
func (perm Permissions) String() string {
	s := make([]string, len(perm))

	for i := range perm {
		s[i] = perm[i].String()
	}

	return strings.Join(s, ", ")
}
