package acl

import (
	"strings"

	"github.com/photoprism/photoprism/pkg/enum"
	"github.com/photoprism/photoprism/pkg/list"
)

// Permission scopes to Grant multiple Permissions for a Resource.
const (
	ScopeRead  Permission = "read"
	ScopeWrite Permission = "write"
)

var (
	// GrantScopeRead grants read-level permissions for scopes.
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
	// GrantScopeWrite grants write-level permissions for scopes.
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

// ScopeAttrWithin returns a scope that permits a resource and action only where both lists do, derived from each
// list's own decision for every name in ResourceNames, as ScopePermits makes it for lowercase input from ScopeAttr.
// It is empty if the lists share no resource.
func ScopeAttrWithin(scope, allowed list.Attr) list.Attr {
	if len(scope) == 0 || len(allowed) == 0 {
		return list.Attr{}
	}

	a, b := scopeAttrCopy(scope), scopeAttrCopy(allowed)
	wild := scopeAttrWild(a) && scopeAttrWild(b)
	result := make(list.Attr, 0, len(ResourceNames)+3)

	for _, resource := range ResourceNames {
		// a and b are private sorted copies, so the in-place sort of ScopeAttrPermits leaves them unchanged.
		permitted := ScopeAttrPermits(a, resource, nil) && ScopeAttrPermits(b, resource, nil)

		switch {
		case permitted && !wild:
			result = append(result, &list.KeyValue{Key: resource.String(), Value: enum.True})
		case !permitted && wild:
			result = append(result, &list.KeyValue{Key: resource.String(), Value: enum.False})
		}
	}

	if wild {
		result = append(result, &list.KeyValue{Key: list.Any})
	} else if len(result) == 0 {
		return list.Attr{}
	}

	for _, mod := range []Permission{ScopeRead, ScopeWrite} {
		if scopeAttrModifier(a, mod) || scopeAttrModifier(b, mod) {
			result = append(result, &list.KeyValue{Key: mod.String(), Value: enum.True})
		}
	}

	return result
}

// scopeAttrCopy returns a sorted copy of the list, as ScopeAttrPermits sorts the list it checks in place.
func scopeAttrCopy(attr list.Attr) list.Attr {
	result := make(list.Attr, 0, len(attr))

	for _, kv := range attr {
		if kv != nil {
			result = append(result, &list.KeyValue{Key: strings.ToLower(kv.Key), Value: kv.Value})
		}
	}

	return result.Sort()
}

// scopeAttrWild reports whether the list is a wildcard: exactly "*" or "read", or with a "*" entry.
func scopeAttrWild(attr list.Attr) bool {
	if s := scopeAttrCopy(attr).String(); s == list.Any || s == ScopeRead.String() {
		return true
	}

	for _, kv := range attr {
		if kv != nil && kv.Key == list.Any && kv.Value != enum.False {
			return true
		}
	}

	return false
}

// scopeAttrModifier reports whether the list limits its resources to the read or write operations.
func scopeAttrModifier(attr list.Attr, mod Permission) bool {
	return scopeAttrCopy(attr).Find(mod.String()).Value == enum.True
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
	if a := attr.Find(ScopeRead.String()); a.Value == enum.True && GrantScopeRead.DenyAny(perms) {
		return false
	} else if a = attr.Find(ScopeWrite.String()); a.Value == enum.True && GrantScopeWrite.DenyAny(perms) {
		return false
	}

	return true
}
