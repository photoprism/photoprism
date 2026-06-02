package acl

import (
	"slices"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoleStrings_Strings_SortedAndExcludesNonAssignable(t *testing.T) {
	m := RoleStrings{
		"visitor":     RoleVisitor,
		"":            RoleNone,
		RoleAliasNone: RoleNone,
		"app":         RoleInstance,
		"guest":       RoleGuest,
		"admin":       RoleAdmin,
	}

	got := m.Strings()

	// Expect deterministic, sorted output; the empty string, the "none" alias,
	// the "app" alias, and the "visitor" share-link role are all excluded.
	assert.Equal(t, []string{"admin", "guest"}, got)
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
	t.Run("ExcludesNonAssignable", func(t *testing.T) {
		// The "visitor" role and the "none" alias are never listed for assignment.
		m := RoleStrings{"visitor": RoleVisitor, RoleAliasNone: RoleNone, "admin": RoleAdmin}
		assert.Equal(t, "admin", m.CliUsageString())
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
	t.Run("ClientRolesStringsExcludeNoneAndEmpty", func(t *testing.T) {
		got := ClientRoles.Strings()
		// Contains exactly the assignable roles, order not enforced; the empty
		// string and the "none" alias are excluded.
		assert.ElementsMatch(t, []string{"admin", "instance", "client", "portal", "service"}, got)
		for _, s := range got {
			assert.NotEqual(t, "", s)
			assert.NotEqual(t, "none", s)
		}
	})
	t.Run("UserRolesStringsExcludeNoneVisitorAndEmpty", func(t *testing.T) {
		got := UserRoles.Strings()
		// The "none" alias and the "visitor" share-link role are both excluded.
		assert.ElementsMatch(t, []string{"admin", "guest"}, got)
		for _, s := range got {
			assert.NotEqual(t, "", s)
			assert.NotEqual(t, "none", s)
			assert.NotEqual(t, "visitor", s)
		}
	})
	t.Run("ClientRolesCliUsageStringExcludesNone", func(t *testing.T) {
		u := ClientRoles.CliUsageString()
		// Lists the assignable roles and ends with "or service" (last alphabetically).
		for _, s := range []string{"admin", "client", "instance", "portal", "service"} {
			assert.Contains(t, u, s)
		}
		assert.NotContains(t, u, "none")
		assert.Regexp(t, `, or service$`, u)
	})
	t.Run("UserRolesCliUsageStringExcludesNoneAndVisitor", func(t *testing.T) {
		u := UserRoles.CliUsageString()
		for _, s := range []string{"admin", "guest"} {
			assert.Contains(t, u, s)
		}
		assert.NotContains(t, u, "none")
		assert.NotContains(t, u, "visitor")
		assert.Equal(t, "admin, or guest", u)
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
	t.Run("AdminTier", func(t *testing.T) {
		assert.True(t, IsAdminRole(RoleAdmin))
		assert.True(t, IsAdminRole(RoleClusterAdmin))
	})
	t.Run("NonAdmin", func(t *testing.T) {
		for _, r := range []Role{RoleUser, RoleViewer, RoleGuest, RoleVisitor, RoleInstance, RoleService, RolePortal, RoleClient, RoleNone} {
			assert.Falsef(t, IsAdminRole(r), "role %s must not be admin-tier", r)
		}
	})
}

func TestIsFederatedUserRole(t *testing.T) {
	t.Run("Federatable", func(t *testing.T) {
		// admin and guest are present in this package's UserRoles map (CE) and may
		// be assigned from an external IdP/directory.
		assert.True(t, IsFederatedUserRole(RoleAdmin))
		assert.True(t, IsFederatedUserRole(RoleGuest))
	})
	t.Run("NeverFederatable", func(t *testing.T) {
		// visitor is in UserRoles yet must still be rejected, so the switch has to
		// override map membership; cluster_admin and none are rejected too.
		for _, r := range []Role{RoleClusterAdmin, RoleVisitor, RoleNone} {
			assert.Falsef(t, IsFederatedUserRole(r), "role %s must never be federatable", r)
		}
	})
	t.Run("MachineRoles", func(t *testing.T) {
		for _, r := range []Role{RoleInstance, RoleService, RolePortal, RoleClient} {
			assert.Falsef(t, IsFederatedUserRole(r), "machine role %s must not be federatable", r)
		}
	})
	t.Run("UnknownRole", func(t *testing.T) {
		assert.False(t, IsFederatedUserRole(Role("does-not-exist")))
	})
}
