package acl

import (
	"sort"
	"strings"
)

// RoleStrings represents user role names mapped to roles.
type RoleStrings map[string]Role

// UserRoles maps valid user account roles.
var UserRoles = RoleStrings{
	string(RoleAdmin):   RoleAdmin,
	string(RoleGuest):   RoleGuest,
	string(RoleVisitor): RoleVisitor,
	string(RoleNone):    RoleNone,
	RoleAliasNone:       RoleNone,
}

// ClientRoles maps valid API client roles.
var ClientRoles = RoleStrings{
	string(RoleAdmin):    RoleAdmin,
	string(RoleInstance): RoleInstance,
	"app":                RoleInstance,
	string(RoleService):  RoleService,
	string(RolePortal):   RolePortal,
	string(RoleClient):   RoleClient,
	string(RoleNone):     RoleNone,
	RoleAliasNone:        RoleNone,
}

// Strings returns the roles as string slice.
func (m RoleStrings) Strings() []string {
	result := make([]string, 0, len(m))
	includesNone := false

	for r := range m {
		if r == "app" {
			continue
		}

		if r == RoleAliasNone {
			includesNone = true
		} else if r != string(RoleNone) {
			result = append(result, r)
		}
	}

	sort.Strings(result)

	if includesNone {
		result = append(result, RoleAliasNone)
	}

	return result
}

// String returns the comma separated roles as string.
func (m RoleStrings) String() string {
	return strings.Join(m.Strings(), ", ")
}

// CliUsageString returns the roles as string for use in CLI usage descriptions.
func (m RoleStrings) CliUsageString() string {
	s := m.Strings()

	if l := len(s); l > 1 {
		s[l-1] = "or " + s[l-1]
	}

	return strings.Join(s, ", ")
}

// Roles grants permissions to roles.
type Roles map[Role]Grant

// Allow checks whether the permission is granted based on the role.
func (roles Roles) Allow(role Role, grant Permission) bool {
	if a, ok := roles[role]; ok {
		return a.Allow(grant)
	} else if a, ok = roles[RoleDefault]; ok {
		return a.Allow(grant)
	}

	return false
}

// sidebarRestrictedRoles lists user roles that receive the reduced
// photo viewer sidebar payload instead of the full entity. "contributor"
// is a Pro-edition role that has no constant in the open-source build,
// so it is matched by literal string here.
var sidebarRestrictedRoles = map[Role]struct{}{
	RoleGuest:           {},
	RoleVisitor:         {},
	Role("contributor"): {},
}

// SidebarFullAccess reports whether the role is allowed to see the full
// photo viewer sidebar; it is the single source of truth for sidebar
// access checks.
func SidebarFullAccess(r Role) bool {
	_, restricted := sidebarRestrictedRoles[r]
	return !restricted
}
