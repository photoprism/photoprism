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

// IsFederatedUserRole reports whether role may be assigned to a user account by
// an external identity provider or directory — i.e. through an OIDC/LDAP group
// or attribute mapping, or a configured federation default role. The Portal-only
// cluster_admin (Admin UI access) and the anonymous-only visitor role are never
// federatable, so a compromised IdP/AD cannot escalate a federated login to the
// Portal Admin UI or to the login-disabled visitor role. RoleNone is rejected
// because an account with no role cannot meaningfully be provisioned this way.
// Roles outside UserRoles (e.g. the machine roles) are rejected implicitly.
func IsFederatedUserRole(role Role) bool {
	switch role {
	case RoleClusterAdmin, RoleVisitor, RoleNone:
		return false
	}

	_, ok := UserRoles[string(role)]

	return ok
}

// Strings returns the roles as string slice for display, e.g. CLI help.
// Alias keys ("app" for instance, "uploader" for contributor) and the
// non-assignable none/visitor roles are omitted so the list stays canonical.
func (m RoleStrings) Strings() []string {
	result := make([]string, 0, len(m))

	for r := range m {
		if r == "" || r == RoleAliasNone || r == "app" || r == "uploader" || r == RoleVisitor.String() {
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
