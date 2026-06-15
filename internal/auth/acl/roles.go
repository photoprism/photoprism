package acl

import (
	"sort"
	"strings"

	"github.com/photoprism/photoprism/pkg/clean"
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

// AdminRoles maps the roles that grant administrative privileges. The
// Portal-only cluster_admin is treated as an admin-tier role everywhere admin
// privileges are checked (e.g. user-management self-lockout protection), so a
// cluster_admin owner is not forced or downgraded to the plain admin role.
var AdminRoles = RoleStrings{
	string(RoleAdmin):        RoleAdmin,
	string(RoleClusterAdmin): RoleClusterAdmin,
}

// IsAdminRole reports whether role is an administrative role (admin or cluster_admin).
func IsAdminRole(role Role) bool {
	_, ok := AdminRoles[string(role)]
	return ok
}

// IsFederatedRole reports whether role may be assigned to a user account through
// an external identity provider (OIDC/LDAP). cluster_admin, visitor, and the empty
// role are never federatable, so a compromised IdP can neither escalate an account
// to operator access nor clear its role.
func IsFederatedRole(role Role) bool {
	switch role {
	case RoleNone, RoleClusterAdmin, RoleVisitor:
		return false
	default:
		return true
	}
}

// clusterInstanceRoles is the shared source of truth for the roles a cluster
// group→role mapping or grant may assign as an instance login role, used by the
// Portal resolver and the CE-shared node handlers. It excludes cluster_admin and visitor.
var clusterInstanceRoles = map[Role]struct{}{
	RoleAdmin:       {},
	RoleManager:     {},
	RoleUser:        {},
	RoleContributor: {},
	RoleViewer:      {},
	RoleGuest:       {},
}

// IsClusterInstanceRole reports whether role may be assigned to a user on a cluster instance.
func IsClusterInstanceRole(role Role) bool {
	_, ok := clusterInstanceRoles[role]
	return ok
}

// ClusterInstanceRole normalizes s and returns the matching cluster instance role, or false.
func ClusterInstanceRole(s string) (Role, bool) {
	role := Role(clean.Role(s))
	if IsClusterInstanceRole(role) {
		return role, true
	}
	return RoleNone, false
}

// FederatedRoleUpdate reports the account role an external identity provider may
// apply to an existing user, and whether to apply it. Returns ok=false when the
// current or mapped role is non-federatable, or when the role is unchanged, so a
// directory sync can neither escalate nor clear a non-federatable role.
func FederatedRoleUpdate(current, mapped Role) (Role, bool) {
	if !IsFederatedRole(current) || !IsFederatedRole(mapped) || current == mapped {
		return RoleNone, false
	}

	return mapped, true
}

// Strings returns the roles as string slice for display, e.g. CLI help.
func (m RoleStrings) Strings() []string {
	result := make([]string, 0, len(m))

	for r := range m {
		if r == "" || r == RoleAliasNone || r == "app" || r == RoleVisitor.String() {
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
