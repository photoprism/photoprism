package oauth

import (
	"slices"
	"strings"

	"github.com/photoprism/photoprism/internal/auth/acl"
)

// DefaultScope is requested when an authorization request has no scope, and included in every request.
var DefaultScope = acl.ResourceMCP.String()

// AppScopes lists the scopes that apps can request: the MCP endpoint and the library resources its tools read.
var AppScopes = []string{
	acl.ResourceMCP.String(),
	acl.ResourcePhotos.String(),
	acl.ResourceVideos.String(),
	acl.ResourceAlbums.String(),
	acl.ResourceFavorites.String(),
	acl.ResourceMoments.String(),
	acl.ResourceCalendar.String(),
	acl.ResourceFolders.String(),
	acl.ResourcePlaces.String(),
	acl.ResourceLabels.String(),
	acl.ResourcePeople.String(),
	acl.ResourceCameras.String(),
	acl.ResourceLenses.String(),
	acl.ResourceFiles.String(),
}

// scopePerms lists the permissions of which a role needs at least one to use a scope.
var scopePerms = acl.Permissions{acl.ActionView, acl.ActionSearch, acl.AccessAll, acl.AccessLibrary, acl.AccessShared, acl.AccessOwn}

// ValidateScopes checks the requested scopes against AppScopes and the scopes the client may request,
// and returns them without duplicates, starting with DefaultScope, which every request includes.
func ValidateScopes(requested []string, clientScope string) ([]string, *RequestError) {
	if !slices.Contains(requested, DefaultScope) {
		requested = append([]string{DefaultScope}, requested...)
	}

	result := make([]string, 0, len(requested))

	for _, s := range requested {
		switch {
		case slices.Contains(result, s):
			continue
		case !slices.Contains(AppScopes, s):
			return nil, NewRequestError(ErrCodeInvalidScope, "unsupported scope")
		case !acl.ScopePermits(clientScope, acl.Resource(s), acl.Permissions{acl.ActionView}):
			return nil, NewRequestError(ErrCodeInvalidScope, "scope not allowed for this client")
		}

		result = append(result, s)
	}

	return result, nil
}

// ClientScopes returns the scopes that are still app scopes and that the client scope still permits.
func ClientScopes(scopes []string, clientScope string) []string {
	result := make([]string, 0, len(scopes))

	for _, s := range scopes {
		if slices.Contains(AppScopes, s) && acl.ScopePermits(clientScope, acl.Resource(s), acl.Permissions{acl.ActionView}) {
			result = append(result, s)
		}
	}

	return result
}

// ScopeUsable reports whether the ACL rules allow a role to use the resource named by a scope.
func ScopeUsable(scope string, role acl.Role) bool {
	if scope == "" {
		return false
	}

	return acl.Rules.AllowAny(acl.Resource(scope), role, scopePerms)
}

// UsableScopes returns the scopes that the ACL rules allow a role to use. Since tokens are only accepted by
// the MCP endpoint, no scope is usable without DefaultScope.
func UsableScopes(scopes []string, role acl.Role) []string {
	if !ScopeUsable(DefaultScope, role) {
		return []string{}
	}

	result := make([]string, 0, len(scopes))

	for _, s := range scopes {
		if ScopeUsable(s, role) {
			result = append(result, s)
		}
	}

	return result
}

// JoinScopes returns the scopes as a space-separated string.
func JoinScopes(scopes []string) string {
	return strings.Join(scopes, " ")
}
