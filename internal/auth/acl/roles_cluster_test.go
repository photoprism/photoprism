package acl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClusterInstanceRole(t *testing.T) {
	t.Run("Assignable", func(t *testing.T) {
		for _, s := range []string{"admin", "manager", "user", "contributor", "viewer", "guest", "  Admin  ", "VIEWER"} {
			role, ok := ClusterInstanceRole(s)
			assert.True(t, ok, "role %q must be assignable", s)
			assert.True(t, IsClusterInstanceRole(role))
		}
	})
	t.Run("Rejected", func(t *testing.T) {
		for _, s := range []string{"cluster_admin", "visitor", "instance", "service", "portal", "client", "none", "", "bogus"} {
			role, ok := ClusterInstanceRole(s)
			assert.False(t, ok, "role %q must be rejected", s)
			assert.Equal(t, RoleNone, role)
		}
	})
}

func TestClusterInstanceRolesCliUsageString(t *testing.T) {
	u := ClusterInstanceRolesCliUsageString()

	// Privilege order, comma-separated, with "or" before the last role.
	assert.Equal(t, "admin, manager, user, contributor, viewer, or guest", u)

	// The derived membership set and the slice source of truth must agree, and
	// every listed role must be accepted by ClusterInstanceRole.
	for _, role := range ClusterInstanceRoles {
		assert.True(t, IsClusterInstanceRole(role), "%s missing from membership set", role)
		assert.Contains(t, u, role.String())
	}
	assert.NotContains(t, u, "cluster_admin")
	assert.NotContains(t, u, "visitor")
}
