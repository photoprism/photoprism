package acl

import (
	"sort"
	"strings"

	"github.com/photoprism/photoprism/pkg/txt"
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

// AdminRoles maps the roles that grant administrative privileges. The Portal-only
// cluster_admin counts as admin-tier wherever admin checks run (e.g. self-lockout
// protection), so a cluster_admin owner is never downgraded to the plain admin role.
var AdminRoles = RoleStrings{
	string(RoleAdmin):        RoleAdmin,
	string(RoleClusterAdmin): RoleClusterAdmin,
}

// Strings returns the roles as string slice for display, e.g. CLI help.
func (m RoleStrings) Strings() []string {
	result := make([]string, 0, len(m))

	for r := range m {
		// Skip empty/none, the anonymous visitor role, and display aliases
		// (app→instance, uploader→contributor).
		if r == "" || r == RoleAliasNone || r == RoleVisitor.String() || r == "app" || r == "uploader" {
			continue
		}

		result = append(result, r)
	}

	sort.Strings(result)

	return result
}

// String returns the comma separated roles as string.
func (m RoleStrings) String() string {
	return strings.Join(m.Strings(), ", ")
}

// CliUsageString returns the roles as string for use in CLI usage descriptions.
func (m RoleStrings) CliUsageString() string {
	return txt.JoinOr(m.Strings())
}

// RolesCliUsageString formats roles for CLI usage help, preserving the given order
// and placing "or" before the last entry (see txt.JoinOr).
func RolesCliUsageString(roles []Role) string {
	s := make([]string, len(roles))
	for i, role := range roles {
		s[i] = role.String()
	}
	return txt.JoinOr(s)
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
