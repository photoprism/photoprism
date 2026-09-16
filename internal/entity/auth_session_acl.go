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

// SeesPrivatePeople reports whether the session may see the people Subject.NameWithheld covers,
// so every surface that resolves a person's name answers the same question about the same session.
//
// The scope is read alongside the role, because these names travel on other resources: a
// credential admitted on photos or files reaches them without people ever being authorized.
func (m *Session) SeesPrivatePeople() bool {
	if m == nil {
		return true
	} else if !m.Grants(acl.ResourcePeople, acl.AccessPrivate) {
		return false
	}

	return m.NoScope() || m.ValidateScope(acl.ResourcePeople, acl.Permissions{acl.AccessPrivate})
}

// SeesFullDetail reports whether the session may receive the full detail of a resource it has
// already been admitted to, rather than the reduced projection: the role must grant view and
// whole-library reach, and the credential's scope must permit viewing that resource. It reads the
// role and scope AuthAny reads, and nothing else, so the caller owes it the admission checks.
// A nil session is internal or CLI use and is unrestricted.
func (m *Session) SeesFullDetail(resource acl.Resource) bool {
	if m == nil {
		return true
	}

	// A principal with neither an account nor a credential is a share-link visitor at most.
	if !m.IsRegistered() && !m.IsClient() {
		return false
	}

	if !m.Grants(resource, acl.ActionView) ||
		!m.GrantsAny(resource, acl.Permissions{acl.AccessAll, acl.AccessLibrary}) {
		return false
	}

	// An empty scope is unrestricted for an account session and insufficient for a credential,
	// as it is on the handler.
	if m.IsClient() {
		return m.ValidateScope(resource, acl.Permissions{acl.ActionView})
	}

	return !m.HasScope() || m.ValidateScope(resource, acl.Permissions{acl.ActionView})
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
