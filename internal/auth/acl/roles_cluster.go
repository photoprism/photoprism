package acl

import (
	"github.com/photoprism/photoprism/pkg/clean"
)

// ClusterInstanceRoles lists, in descending privilege order, the roles a cluster
// group→role mapping or grant may assign as an instance login role. Shared by the
// Portal resolver, the CE node handlers, and CLI usage help; it excludes the
// Portal-only cluster_admin and the anonymous visitor role.
var ClusterInstanceRoles = []Role{
	RoleAdmin,
	RoleManager,
	RoleUser,
	RoleContributor,
	RoleViewer,
	RoleGuest,
}

// clusterInstanceRoles indexes ClusterInstanceRoles for O(1) membership checks.
var clusterInstanceRoles = func() map[Role]struct{} {
	m := make(map[Role]struct{}, len(ClusterInstanceRoles))
	for _, role := range ClusterInstanceRoles {
		m[role] = struct{}{}
	}
	return m
}()

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

// ClusterInstanceRolesCliUsageString returns ClusterInstanceRoles formatted for CLI usage help.
// It is the edition-independent source for the --oidc-group-role and --cluster-allow-group-roles
// flag help, so the listed roles always match what ClusterInstanceRole accepts.
func ClusterInstanceRolesCliUsageString() string {
	return RolesCliUsageString(ClusterInstanceRoles)
}
