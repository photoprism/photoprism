package acl

// IsAdminRole reports whether role is an administrative role (admin or cluster_admin).
func IsAdminRole(role Role) bool {
	_, ok := AdminRoles[string(role)]
	return ok
}

// IsFederatedRole reports whether role may be assigned through an external identity
// provider (OIDC/LDAP). cluster_admin, visitor, and the empty role are never federatable,
// so a compromised IdP can neither escalate to operator access nor clear a role.
func IsFederatedRole(role Role) bool {
	switch role {
	case RoleNone, RoleClusterAdmin, RoleVisitor:
		return false
	default:
		return true
	}
}

// FederatedRoleUpdate reports the account role an external identity provider may apply to an
// existing user, and whether to apply it. Returns ok=false when the current or mapped role is
// non-federatable or unchanged, so a directory sync can neither escalate nor clear such a role.
func FederatedRoleUpdate(current, mapped Role) (Role, bool) {
	if !IsFederatedRole(current) || !IsFederatedRole(mapped) || current == mapped {
		return RoleNone, false
	}

	return mapped, true
}
