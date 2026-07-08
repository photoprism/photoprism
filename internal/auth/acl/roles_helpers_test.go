package acl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsAdminRole(t *testing.T) {
	assert.True(t, IsAdminRole(RoleAdmin))
	assert.True(t, IsAdminRole(RoleClusterAdmin))
	assert.False(t, IsAdminRole(RoleUser))
	assert.False(t, IsAdminRole(RoleViewer))
	assert.False(t, IsAdminRole(RoleGuest))
	assert.False(t, IsAdminRole(RoleVisitor))
	assert.False(t, IsAdminRole(RoleNone))
	assert.False(t, IsAdminRole(RoleManager))
}

func TestIsFederatedRole(t *testing.T) {
	t.Run("Federatable", func(t *testing.T) {
		assert.True(t, IsFederatedRole(RoleAdmin))
		assert.True(t, IsFederatedRole(RoleUser))
		assert.True(t, IsFederatedRole(RoleViewer))
		assert.True(t, IsFederatedRole(RoleGuest))
		assert.True(t, IsFederatedRole(RoleManager))
		assert.True(t, IsFederatedRole(RoleContributor))
	})
	t.Run("NotFederatable", func(t *testing.T) {
		// cluster_admin is the Portal operator role and visitor is anonymous;
		// neither may be granted or revoked via an external IdP/AD, and an empty
		// role must never be applied by a sync.
		assert.False(t, IsFederatedRole(RoleClusterAdmin))
		assert.False(t, IsFederatedRole(RoleVisitor))
		assert.False(t, IsFederatedRole(RoleNone))
	})
}

func TestFederatedRoleUpdate(t *testing.T) {
	t.Run("AppliesChangedFederatableRole", func(t *testing.T) {
		role, ok := FederatedRoleUpdate(RoleUser, RoleViewer)
		assert.True(t, ok)
		assert.Equal(t, RoleViewer, role)
	})
	t.Run("UnchangedRoleNotApplied", func(t *testing.T) {
		_, ok := FederatedRoleUpdate(RoleUser, RoleUser)
		assert.False(t, ok)
	})
	t.Run("NeverDowngradesClusterAdmin", func(t *testing.T) {
		// An existing cluster_admin/visitor account must not be touched.
		_, ok := FederatedRoleUpdate(RoleClusterAdmin, RoleAdmin)
		assert.False(t, ok)
		_, ok = FederatedRoleUpdate(RoleVisitor, RoleGuest)
		assert.False(t, ok)
	})
	t.Run("NeverEscalatesToNonFederatable", func(t *testing.T) {
		// The directory must not promote to cluster_admin/visitor.
		_, ok := FederatedRoleUpdate(RoleUser, RoleClusterAdmin)
		assert.False(t, ok)
		_, ok = FederatedRoleUpdate(RoleAdmin, RoleVisitor)
		assert.False(t, ok)
		_, ok = FederatedRoleUpdate(RoleUser, RoleNone)
		assert.False(t, ok)
	})
}
