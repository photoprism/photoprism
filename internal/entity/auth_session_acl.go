package entity

import (
	"github.com/photoprism/photoprism/internal/auth/acl"
)

// Grants reports whether the session is granted perm on the resource. A client session must be
// permitted by the client role and, when a user is present, by the user role as well, since a client
// acting for a user is limited by both. A nil session is internal or CLI use and is not restricted.
func (m *Session) Grants(resource acl.Resource, perm acl.Permission) bool {
	if m == nil {
		return true
	}

	if m.IsClient() {
		if !acl.Rules.Allow(resource, m.GetClientRole(), perm) {
			return false
		} else if m.NoUser() {
			return true
		}
	}

	return acl.Rules.Allow(resource, m.GetUserRole(), perm)
}

// GrantsAny reports whether the session is granted at least one of the perms on the resource.
func (m *Session) GrantsAny(resource acl.Resource, perms acl.Permissions) bool {
	for i := range perms {
		if m.Grants(resource, perms[i]) {
			return true
		}
	}

	return false
}

// Denies reports whether the session is not granted perm on the resource.
func (m *Session) Denies(resource acl.Resource, perm acl.Permission) bool {
	return !m.Grants(resource, perm)
}

// DeniesAll reports whether the session is granted none of the perms on the resource.
func (m *Session) DeniesAll(resource acl.Resource, perms acl.Permissions) bool {
	return !m.GrantsAny(resource, perms)
}

// HasSharedAccessOnly reports whether the session's effective access to the resource is limited to
// shared content, evaluating the client role alongside the user role. A nil session is unrestricted.
func (m *Session) HasSharedAccessOnly(resource acl.Resource) bool {
	if m == nil {
		return false
	}

	if m.Denies(resource, acl.AccessShared) {
		return false
	}

	return m.DeniesAll(resource, acl.Permissions{acl.AccessAll, acl.AccessLibrary})
}
