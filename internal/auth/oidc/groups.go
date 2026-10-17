package oidc

import (
	"strings"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/pkg/clean"
)

// NormalizeGroupID lowercases and sanitizes a group identifier (GUID or name).
func NormalizeGroupID(id string) string {
	return strings.ToLower(clean.Auth(id))
}

// GroupsFromClaims extracts group identifiers from token or userinfo claims and detects Entra-style overage markers.
func GroupsFromClaims(claims map[string]any, claimName string) (groups []string, overage bool) {
	if len(claims) == 0 {
		return nil, false
	}

	if claimName == "" {
		claimName = "groups"
	}

	if raw, ok := claims[claimName]; ok {
		groups = append(groups, normalizeGroupValues(raw)...)
	}

	if raw, ok := claims["_claim_names"]; ok {
		if names, ok := raw.(map[string]any); ok {
			if _, ok := names[claimName]; ok {
				overage = true
			}
		}
	}

	return uniqueGroups(groups), overage
}

// MergeGroups returns a deduplicated, normalized list of group identifiers
// drawn from one or more source slices (typically the ID token and userinfo
// groups claims, plus any LDAP-derived virtual groups). Empty entries are
// dropped; remaining entries keep first-occurrence order after normalization
// through NormalizeGroupID. Returns nil when no groups remain.
func MergeGroups(sources ...[]string) []string {
	total := 0

	for _, src := range sources {
		total += len(src)
	}

	if total == 0 {
		return nil
	}

	merged := make([]string, 0, total)

	for _, src := range sources {
		merged = append(merged, src...)
	}

	return uniqueGroups(merged)
}

// MapGroupsToRole returns the first matching role for the provided groups using the supplied mapping.
func MapGroupsToRole(groups []string, mapping map[string]acl.Role) (acl.Role, bool) {
	if len(groups) == 0 || len(mapping) == 0 {
		return acl.RoleNone, false
	}

	for _, g := range uniqueGroups(groups) {
		if role, ok := mapping[g]; ok && role != acl.RoleNone {
			return role, true
		}
	}

	return acl.RoleNone, false
}

// portalRoleFallback lists, for each cluster instance role, the roles an edition may assign
// in its place, strongest first; none grants access the Portal role does not include.
var portalRoleFallback = map[acl.Role][]acl.Role{
	acl.RoleAdmin:       {acl.RoleAdmin, acl.RoleGuest},
	acl.RoleManager:     {acl.RoleManager, acl.RoleUser, acl.RoleViewer, acl.RoleGuest},
	acl.RoleUser:        {acl.RoleUser, acl.RoleViewer, acl.RoleGuest},
	acl.RoleContributor: {acl.RoleContributor, acl.RoleGuest},
	acl.RoleViewer:      {acl.RoleViewer, acl.RoleGuest},
	acl.RoleGuest:       {acl.RoleGuest},
}

// PortalGrantedRole returns the instance login role from the pp_role claim, and true
// when pp_issuer_kind marks the token as Portal-issued and pp_role is set. A role this
// edition does not have maps to the strongest fallback it has, and any other value to guest.
func PortalGrantedRole(claims map[string]any) (acl.Role, bool) {
	if len(claims) == 0 {
		return acl.RoleNone, false
	}

	if kind, _ := claims["pp_issuer_kind"].(string); kind != acl.RolePortal.String() {
		return acl.RoleNone, false
	}

	raw, found := claims["pp_role"]
	roleName, isString := raw.(string)

	if !found || isString && strings.TrimSpace(roleName) == "" {
		return acl.RoleNone, false
	} else if role := acl.ParseRole(roleName); acl.IsFederatedRole(role) {
		return role, true
	} else if granted, ok := acl.ClusterInstanceRole(roleName); ok {
		return editionRole(granted), true
	}

	return acl.RoleGuest, true
}

// editionRole returns the strongest role in the fallback list of role that this edition has,
// or guest when it has none of them.
func editionRole(role acl.Role) acl.Role {
	for _, r := range portalRoleFallback[role] {
		if acl.ParseRole(r.String()) == r {
			return r
		}
	}

	return acl.RoleGuest
}

// HasAnyGroup returns true when at least one of the user's groups matches a required group.
func HasAnyGroup(groups []string, required []string) bool {
	if len(required) == 0 {
		return true
	}

	normalized := make(map[string]struct{}, len(uniqueGroups(groups)))

	for _, g := range uniqueGroups(groups) {
		normalized[g] = struct{}{}
	}

	for _, r := range required {
		if _, ok := normalized[NormalizeGroupID(r)]; ok {
			return true
		}
	}

	return false
}

func normalizeGroupValues(raw any) []string {
	switch v := raw.(type) {
	case []string:
		return normalizeGroupSlice(v)
	case []any:
		result := make([]string, 0, len(v))

		for _, s := range v {
			if val, ok := s.(string); ok {
				result = append(result, val)
			}
		}

		return normalizeGroupSlice(result)
	case string:
		return normalizeGroupSlice([]string{v})
	default:
		return nil
	}
}

// normalizeGroupSlice sanitizes and lowercases each group identifier in the provided slice.
func normalizeGroupSlice(values []string) []string {
	result := make([]string, 0, len(values))

	for _, v := range values {
		if n := NormalizeGroupID(v); n != "" {
			result = append(result, n)
		}
	}

	return result
}

// uniqueGroups returns a deduplicated, normalized list of group identifiers.
func uniqueGroups(values []string) []string {
	if len(values) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))

	for _, v := range normalizeGroupSlice(values) {
		if _, ok := seen[v]; ok {
			continue
		}

		seen[v] = struct{}{}
		result = append(result, v)
	}

	return result
}
