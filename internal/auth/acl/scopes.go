package acl

import (
	"strings"

	"github.com/photoprism/photoprism/pkg/list"
)

// Permission scopes to Grant multiple Permissions for a Resource.
const (
	ScopeRead  Permission = "read"
	ScopeWrite Permission = "write"
)

var (
	GrantScopeRead = Grant{
		AccessShared:    true,
		AccessLibrary:   true,
		AccessPrivate:   true,
		AccessOwn:       true,
		AccessAll:       true,
		ActionSearch:    true,
		ActionView:      true,
		ActionDownload:  true,
		ActionSubscribe: true,
	}
	GrantScopeWrite = Grant{
		AccessShared:    true,
		AccessLibrary:   true,
		AccessPrivate:   true,
		AccessOwn:       true,
		AccessAll:       true,
		ActionUpload:    true,
		ActionCreate:    true,
		ActionUpdate:    true,
		ActionShare:     true,
		ActionDelete:    true,
		ActionRate:      true,
		ActionReact:     true,
		ActionManage:    true,
		ActionManageOwn: true,
	}
)

// ScopeAttr parses an auth scope string and returns a normalized Attr
// with duplicate and invalid entries removed.
func ScopeAttr(s string) list.Attr {
	if s == "" {
		return list.Attr{}
	}

	return list.ParseAttr(strings.ToLower(s))
}

// ScopePermits sanitizes the raw scope string and then calls ScopeAttrPermits for
// the actual authorization check.
func ScopePermits(scope string, resource Resource, perms Permissions) bool {
	if scope == "" {
		return false
	}

	// Parse scope to check for resources and permissions.
	return ScopeAttrPermits(ScopeAttr(scope), resource, perms)
}

// ScopeAttrPermits evaluates an already-parsed scope attribute list against a
// resource and permission set, enforcing wildcard/read/write semantics.
func ScopeAttrPermits(attr list.Attr, resource Resource, perms Permissions) bool {
	if len(attr) == 0 {
		return false
	}

	scope := attr.String()

	// Skip detailed check and allow all if scope is "*".
	if scope == list.Any {
		return true
	}

	// Skip resource check if scope includes all read operations.
	if scope == ScopeRead.String() {
		return !GrantScopeRead.DenyAny(perms)
	}

	// Check if resource is within scope.
	if granted := attr.Contains(resource.String()); !granted {
		return false
	}

	// Check if permission is within scope.
	if len(perms) == 0 {
		return true
	}

	// Check if scope is limited to read or write operations.
	if a := attr.Find(ScopeRead.String()); a.Value == list.True && GrantScopeRead.DenyAny(perms) {
		return false
	} else if a = attr.Find(ScopeWrite.String()); a.Value == list.True && GrantScopeWrite.DenyAny(perms) {
		return false
	}

	return true
}
