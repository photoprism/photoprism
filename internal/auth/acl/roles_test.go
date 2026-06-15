package acl

import (
	"slices"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoleStrings_Strings_SortedAndNoEmpty(t *testing.T) {
	m := RoleStrings{
		"visitor": RoleVisitor,
		"":        RoleNone,
		"guest":   RoleGuest,
		"admin":   RoleAdmin,
	}

	got := m.Strings()

	// Expect deterministic, sorted output, no empty entries, and visitor
	// excluded (reserved for anonymous/link-share access, never offered).
	assert.Equal(t, []string{"admin", "guest"}, got)
	assert.NotContains(t, got, "visitor")
	assert.True(t, sort.StringsAreSorted(got))
}

func TestRoleStrings_String_Join(t *testing.T) {
	m := RoleStrings{
		"b": RoleUser,
		"a": RoleAdmin,
	}

	// Sorted keys joined by ", ".
	assert.Equal(t, "a, b", m.String())
}

func TestRoleStrings_CliUsageString(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", (RoleStrings{}).CliUsageString())
	})
	t.Run("Single", func(t *testing.T) {
		m := RoleStrings{"admin": RoleAdmin}
		assert.Equal(t, "admin", m.CliUsageString())
	})
	t.Run("Two", func(t *testing.T) {
		m := RoleStrings{"guest": RoleGuest, "admin": RoleAdmin}
		// Note the comma before "or" matches current implementation.
		assert.Equal(t, "admin, or guest", m.CliUsageString())
	})
	t.Run("Three", func(t *testing.T) {
		m := RoleStrings{"user": RoleUser, "guest": RoleGuest, "admin": RoleAdmin}
		assert.Equal(t, "admin, guest, or user", m.CliUsageString())
	})
	t.Run("ExcludesVisitor", func(t *testing.T) {
		m := RoleStrings{"visitor": RoleVisitor, "guest": RoleGuest, "admin": RoleAdmin}
		assert.Equal(t, "admin, or guest", m.CliUsageString())
	})
}

func TestRoles_Allow(t *testing.T) {
	t.Run("SpecificRoleGrant", func(t *testing.T) {
		roles := Roles{
			RoleVisitor: GrantViewShared, // denies delete
		}
		assert.True(t, roles.Allow(RoleVisitor, ActionView))
		assert.True(t, roles.Allow(RoleVisitor, ActionDownload))
		assert.False(t, roles.Allow(RoleVisitor, ActionDelete))
	})
	t.Run("DefaultFallbackUsed", func(t *testing.T) {
		roles := Roles{
			RoleDefault: GrantViewAll, // allows view, denies delete
		}
		assert.True(t, roles.Allow(RoleUser, ActionView))
		assert.False(t, roles.Allow(RoleUser, ActionDelete))
	})
	t.Run("SpecificOverridesDefaultNoFallback", func(t *testing.T) {
		roles := Roles{
			RoleVisitor: GrantViewShared, // denies delete
			RoleDefault: GrantFullAccess, // would allow delete, must NOT be used
		}
		assert.False(t, roles.Allow(RoleVisitor, ActionDelete))
	})
	t.Run("NoMatchAndNoDefault", func(t *testing.T) {
		roles := Roles{
			RoleVisitor: GrantViewShared,
		}
		assert.False(t, roles.Allow(RoleUser, ActionView))
	})
}

func TestRoleStrings_GlobalMaps_AliasNoneAndUsage(t *testing.T) {
	t.Run("ClientRolesStringsExcludeAliasNoneAndEmpty", func(t *testing.T) {
		got := ClientRoles.Strings()
		// Contains exactly the expected elements, order not enforced; the "none"
		// alias and the empty role are excluded from display.
		assert.ElementsMatch(t, []string{"admin", "instance", "client", "portal", "service"}, got)
		assert.NotContains(t, got, "none")
		// Does not include empty string.
		for _, s := range got {
			assert.NotEqual(t, "", s)
		}
	})
	t.Run("UserRolesStringsExcludeAliasNoneEmptyAndVisitor", func(t *testing.T) {
		got := UserRoles.Strings()
		assert.ElementsMatch(t, []string{"admin", "guest"}, got)
		assert.NotContains(t, got, "none")
		assert.NotContains(t, got, "visitor")
		for _, s := range got {
			assert.NotEqual(t, "", s)
		}
	})
	t.Run("ClientRolesCliUsageStringExcludesNoneAndOrBeforeLast", func(t *testing.T) {
		u := ClientRoles.CliUsageString()
		// Should list known roles and end with "or service"; the "none" alias is excluded.
		for _, s := range []string{"admin", "client", "instance", "portal", "service"} {
			assert.Contains(t, u, s)
		}
		assert.NotContains(t, u, "none")
		assert.Regexp(t, `, or service$`, u)
	})
	t.Run("UserRolesCliUsageStringExcludesNoneVisitorAndOrBeforeLast", func(t *testing.T) {
		u := UserRoles.CliUsageString()
		for _, s := range []string{"admin", "guest"} {
			assert.Contains(t, u, s)
		}
		assert.NotContains(t, u, "none")
		assert.NotContains(t, u, "visitor")
		assert.Regexp(t, `, or guest$`, u)
	})
	t.Run("AliasNoneMapsToRoleNone", func(t *testing.T) {
		assert.Equal(t, RoleNone, ClientRoles[RoleAliasNone])
		assert.Equal(t, RoleNone, UserRoles[RoleAliasNone])
	})
}

func TestRole_Pretty_And_ParseRole(t *testing.T) {
	t.Run("PrettyAdmin", func(t *testing.T) {
		r := Role("admin")
		assert.Equal(t, "Admin", r.Pretty())
	})
	t.Run("PrettyNoneEmpty", func(t *testing.T) {
		r := Role("")
		assert.Equal(t, "None", r.Pretty())
	})
	t.Run("PrettyNoneAlias", func(t *testing.T) {
		r := Role(RoleAliasNone)
		assert.Equal(t, "None", r.Pretty())
	})
	t.Run("ParseRoleTokensToNone", func(t *testing.T) {
		tokens := []string{"", "0", "false", "nil", "null", "nan"}
		for _, s := range tokens {
			assert.Equal(t, RoleNone, ParseRole(s))
		}
	})
	t.Run("ParseRoleAliasNone", func(t *testing.T) {
		assert.Equal(t, RoleNone, ParseRole("none"))
	})
	t.Run("ParseRoleAdmin", func(t *testing.T) {
		assert.Equal(t, RoleAdmin, ParseRole("admin"))
	})
}

func TestPermission_String_And_Compare(t *testing.T) {
	p := Permission("action_update_own")
	assert.Equal(t, "action update own", p.String())
	assert.True(t, p.Equal("Action Update Own"))
	assert.True(t, p.NotEqual("delete"))
}

func TestPermissions_String_Join(t *testing.T) {
	perms := Permissions{ActionView, ActionUpdateOwn, AccessAll}
	s := perms.String()
	assert.Contains(t, s, "view")
	assert.Contains(t, s, "update own")
	assert.Contains(t, s, "access all")
}

func TestResource_Default_String_And_Compare(t *testing.T) {
	var r Resource
	assert.Equal(t, "default", r.String())
	assert.True(t, r.Equal("DEFAULT"))
	assert.True(t, r.NotEqual("photos"))
}

func TestResourceNames_ContainsCore(t *testing.T) {
	want := []Resource{ResourceDefault, ResourcePhotos, ResourceAlbums, ResourceWebDAV, ResourceApi}
	for _, w := range want {
		found := slices.Contains(ResourceNames, w)
		assert.Truef(t, found, "resource %s not found in ResourceNames", w)
	}
}

func TestIsAdminRole(t *testing.T) {
	assert.True(t, IsAdminRole(RoleAdmin))
	assert.True(t, IsAdminRole(RoleClusterAdmin))
	assert.False(t, IsAdminRole(RoleUser))
	assert.False(t, IsAdminRole(RoleViewer))
	assert.False(t, IsAdminRole(RoleGuest))
	assert.False(t, IsAdminRole(RoleVisitor))
	assert.False(t, IsAdminRole(RoleNone))
	assert.False(t, IsAdminRole(Role("manager")))
}

func TestIsFederatedRole(t *testing.T) {
	t.Run("Federatable", func(t *testing.T) {
		assert.True(t, IsFederatedRole(RoleAdmin))
		assert.True(t, IsFederatedRole(RoleUser))
		assert.True(t, IsFederatedRole(RoleViewer))
		assert.True(t, IsFederatedRole(RoleGuest))
		assert.True(t, IsFederatedRole(Role("manager")))
		assert.True(t, IsFederatedRole(Role("contributor")))
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
