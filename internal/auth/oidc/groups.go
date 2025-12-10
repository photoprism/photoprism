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
