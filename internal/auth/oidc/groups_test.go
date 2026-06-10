package oidc

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/auth/acl"
)

func TestGroupsFromClaims(t *testing.T) {
	claims := map[string]any{
		"groups": []any{"ABC-123", "def-456", 7},
	}

	groups, overage := GroupsFromClaims(claims, "groups")

	assert.False(t, overage)
	assert.Equal(t, []string{"abc-123", "def-456"}, groups)
}

func TestGroupsFromClaimsOverage(t *testing.T) {
	claims := map[string]any{
		"_claim_names": map[string]any{
			"groups": "src1",
		},
	}

	groups, overage := GroupsFromClaims(claims, "groups")

	assert.True(t, overage)
	assert.Nil(t, groups)
}

func TestMapGroupsToRole(t *testing.T) {
	mapping := map[string]acl.Role{
		"abc-123": acl.RoleAdmin,
		"def-456": acl.RoleGuest,
	}

	role, ok := MapGroupsToRole([]string{"zzz", "DEF-456"}, mapping)

	assert.True(t, ok)
	assert.Equal(t, acl.RoleGuest, role)
}

func TestHasAnyGroup(t *testing.T) {
	required := []string{"abc-123", "def-456"}

	assert.True(t, HasAnyGroup([]string{"ABC-123"}, required))
	assert.False(t, HasAnyGroup([]string{"zzz"}, required))
	assert.True(t, HasAnyGroup([]string{"zzz"}, nil))
}

func TestPortalGrantedRole(t *testing.T) {
	portal := func(role any) map[string]any {
		return map[string]any{"pp_issuer_kind": "portal", "pp_role": role}
	}
	t.Run("Admin", func(t *testing.T) {
		role, ok := PortalGrantedRole(portal("admin"))
		assert.True(t, ok)
		assert.Equal(t, acl.RoleAdmin, role)
	})
	t.Run("Guest", func(t *testing.T) {
		role, ok := PortalGrantedRole(portal("guest"))
		assert.True(t, ok)
		assert.Equal(t, acl.RoleGuest, role)
	})
	t.Run("RuntimeRegisteredInstanceRole", func(t *testing.T) {
		// Pro/portal builds register extra instance roles (e.g. viewer) into
		// acl.UserRoles at startup; the helper must honor whatever is registered.
		key := acl.RoleViewer.String()
		if _, had := acl.UserRoles[key]; !had {
			acl.UserRoles[key] = acl.RoleViewer
			defer delete(acl.UserRoles, key)
		}
		role, ok := PortalGrantedRole(portal("viewer"))
		assert.True(t, ok)
		assert.Equal(t, acl.RoleViewer, role)
	})
	t.Run("WhitespaceTrimmed", func(t *testing.T) {
		role, ok := PortalGrantedRole(portal("  admin  "))
		assert.True(t, ok)
		assert.Equal(t, acl.RoleAdmin, role)
	})
	t.Run("ClusterAdminRejected", func(t *testing.T) {
		role, ok := PortalGrantedRole(portal("cluster_admin"))
		assert.False(t, ok)
		assert.Equal(t, acl.RoleNone, role)
	})
	t.Run("VisitorRejected", func(t *testing.T) {
		role, ok := PortalGrantedRole(portal("visitor"))
		assert.False(t, ok)
		assert.Equal(t, acl.RoleNone, role)
	})
	t.Run("EmptyRole", func(t *testing.T) {
		role, ok := PortalGrantedRole(portal(""))
		assert.False(t, ok)
		assert.Equal(t, acl.RoleNone, role)
	})
	t.Run("UnknownRole", func(t *testing.T) {
		role, ok := PortalGrantedRole(portal("wizard"))
		assert.False(t, ok)
		assert.Equal(t, acl.RoleNone, role)
	})
	t.Run("NonPortalIssuerKind", func(t *testing.T) {
		role, ok := PortalGrantedRole(map[string]any{"pp_issuer_kind": "upstream", "pp_role": "admin"})
		assert.False(t, ok)
		assert.Equal(t, acl.RoleNone, role)
	})
	t.Run("MissingIssuerKind", func(t *testing.T) {
		role, ok := PortalGrantedRole(map[string]any{"pp_role": "admin"})
		assert.False(t, ok)
		assert.Equal(t, acl.RoleNone, role)
	})
	t.Run("EmptyClaims", func(t *testing.T) {
		role, ok := PortalGrantedRole(nil)
		assert.False(t, ok)
		assert.Equal(t, acl.RoleNone, role)
	})
}
