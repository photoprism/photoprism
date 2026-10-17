package oidc

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

func TestMergeGroups(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		merged := MergeGroups([]string{"ABC-123", "def-456"}, []string{"def-456", "GHI-789"})
		assert.Equal(t, []string{"abc-123", "def-456", "ghi-789"}, merged)
	})
	t.Run("DropsEmptyEntries", func(t *testing.T) {
		merged := MergeGroups([]string{"", "abc-123"}, []string{"   "})
		assert.Equal(t, []string{"abc-123"}, merged)
	})
	t.Run("NoSources", func(t *testing.T) {
		assert.Nil(t, MergeGroups())
		assert.Nil(t, MergeGroups(nil, []string{}))
	})
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

func TestNormalizeGroupID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		assert.Equal(t, "photo admins", NormalizeGroupID(" Photo Admins "))
		assert.Equal(t, strings.Repeat("a", 255), NormalizeGroupID(strings.Repeat("A", 255)))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", NormalizeGroupID(""))
		assert.Equal(t, "", NormalizeGroupID("\x00<>"))
	})
	t.Run("TooLong", func(t *testing.T) {
		assert.Equal(t, "", NormalizeGroupID(strings.Repeat("a", 256)))
		assert.Equal(t, "", NormalizeGroupID("CN="+strings.Repeat("A", 251)+" Admins,OU=Groups"))
		assert.Equal(t, "", NormalizeGroupID(strings.Repeat("a", 253)+"  users"))
	})
	t.Run("Idempotent", func(t *testing.T) {
		for _, s := range []string{"Photo Admins", strings.Repeat("\u023a", 170), strings.Repeat("A", 255)} {
			id := NormalizeGroupID(s)
			assert.NotEmpty(t, id, s)
			assert.Equal(t, id, NormalizeGroupID(id), s)
		}

		assert.Equal(t, "", NormalizeGroupID(strings.Repeat("\u023a", 171)))
	})
}

func TestGroupsFromClaims_TooLong(t *testing.T) {
	groups, _ := GroupsFromClaims(map[string]any{"groups": []any{"Admins", strings.Repeat("a", 253) + "  users"}}, "")
	assert.Equal(t, []string{"admins"}, groups)
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
	t.Run("ClusterAdminGuest", func(t *testing.T) {
		role, ok := PortalGrantedRole(portal("cluster_admin"))
		assert.True(t, ok)
		assert.Equal(t, acl.RoleGuest, role)
	})
	t.Run("VisitorGuest", func(t *testing.T) {
		role, ok := PortalGrantedRole(portal("visitor"))
		assert.True(t, ok)
		assert.Equal(t, acl.RoleGuest, role)
	})
	t.Run("EmptyRole", func(t *testing.T) {
		role, ok := PortalGrantedRole(portal(""))
		assert.False(t, ok)
		assert.Equal(t, acl.RoleNone, role)
	})
	t.Run("UnknownRole", func(t *testing.T) {
		role, ok := PortalGrantedRole(portal("wizard"))
		assert.True(t, ok)
		assert.Equal(t, acl.RoleGuest, role)
	})
	t.Run("NotString", func(t *testing.T) {
		role, ok := PortalGrantedRole(portal(42))
		assert.True(t, ok)
		assert.Equal(t, acl.RoleGuest, role)
	})
	t.Run("NoRole", func(t *testing.T) {
		role, ok := PortalGrantedRole(map[string]any{"pp_issuer_kind": "portal"})
		assert.False(t, ok)
		assert.Equal(t, acl.RoleNone, role)
		role, ok = PortalGrantedRole(portal(" "))
		assert.False(t, ok)
		assert.Equal(t, acl.RoleNone, role)
	})
	t.Run("RoleNotInEdition", func(t *testing.T) {
		orig := acl.UserRoles
		t.Cleanup(func() { acl.UserRoles = orig })

		acl.UserRoles = acl.RoleStrings{"admin": acl.RoleAdmin, "guest": acl.RoleGuest, "visitor": acl.RoleVisitor}

		for _, name := range []string{"manager", "user", "contributor", "viewer"} {
			role, ok := PortalGrantedRole(portal(name))
			assert.True(t, ok, name)
			assert.Equal(t, acl.RoleGuest, role, name)
		}

		acl.UserRoles = acl.RoleStrings{"admin": acl.RoleAdmin, "user": acl.RoleUser, "viewer": acl.RoleViewer, "guest": acl.RoleGuest}

		role, ok := PortalGrantedRole(portal("manager"))
		assert.True(t, ok)
		assert.Equal(t, acl.RoleUser, role)
		role, ok = PortalGrantedRole(portal("contributor"))
		assert.True(t, ok)
		assert.Equal(t, acl.RoleGuest, role, "viewer can see the library, which a contributor cannot")

		acl.UserRoles = acl.RoleStrings{"admin": acl.RoleAdmin, "contributor": acl.RoleContributor, "uploader": acl.RoleContributor, "guest": acl.RoleGuest}

		role, ok = PortalGrantedRole(portal("uploader"))
		assert.True(t, ok)
		assert.Equal(t, acl.RoleContributor, role)
		role, ok = PortalGrantedRole(portal("manager"))
		assert.True(t, ok)
		assert.Equal(t, acl.RoleGuest, role)
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

func TestPortalRoleFallback(t *testing.T) {
	for _, role := range acl.ClusterInstanceRoles {
		fallback, ok := portalRoleFallback[role]
		require.True(t, ok, role)
		require.NotEmpty(t, fallback, role)
		assert.Equal(t, role, fallback[0], role)
		assert.Equal(t, acl.RoleGuest, fallback[len(fallback)-1], role)

		for _, r := range fallback {
			assert.True(t, acl.IsClusterInstanceRole(r), "%s: %s", role, r)
		}
	}

	assert.Len(t, portalRoleFallback, len(acl.ClusterInstanceRoles))
}

func TestPortalRoleClaim(t *testing.T) {
	assert.Equal(t, "manager", PortalRoleClaim(map[string]any{"pp_issuer_kind": "portal", "pp_role": " Manager "}))
	assert.Equal(t, "", PortalRoleClaim(map[string]any{"pp_issuer_kind": "portal", "pp_role": 42}))
	assert.Equal(t, "", PortalRoleClaim(map[string]any{"pp_issuer_kind": "upstream", "pp_role": "admin"}))
	assert.Equal(t, "", PortalRoleClaim(nil))
}

func TestEditionRole(t *testing.T) {
	orig := acl.UserRoles
	t.Cleanup(func() { acl.UserRoles = orig })

	acl.UserRoles = acl.RoleStrings{"admin": acl.RoleAdmin, "user": acl.RoleUser, "viewer": acl.RoleViewer, "guest": acl.RoleGuest}

	t.Run("Known", func(t *testing.T) {
		assert.Equal(t, acl.RoleUser, editionRole(acl.RoleUser))
		assert.Equal(t, acl.RoleAdmin, editionRole(acl.RoleAdmin))
	})
	t.Run("Fallback", func(t *testing.T) {
		assert.Equal(t, acl.RoleUser, editionRole(acl.RoleManager))
		assert.Equal(t, acl.RoleGuest, editionRole(acl.RoleContributor))
	})
	t.Run("NotClusterInstanceRole", func(t *testing.T) {
		assert.Equal(t, acl.RoleGuest, editionRole(acl.RoleVisitor))
	})
}
