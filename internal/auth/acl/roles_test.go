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
		// Two items read without a comma before "or" (see txt.JoinOr).
		assert.Equal(t, "admin or guest", m.CliUsageString())
	})
	t.Run("Three", func(t *testing.T) {
		m := RoleStrings{"user": RoleUser, "guest": RoleGuest, "admin": RoleAdmin}
		assert.Equal(t, "admin, guest, or user", m.CliUsageString())
	})
	t.Run("ExcludesVisitor", func(t *testing.T) {
		m := RoleStrings{"visitor": RoleVisitor, "guest": RoleGuest, "admin": RoleAdmin}
		assert.Equal(t, "admin or guest", m.CliUsageString())
	})
}

func TestRolesCliUsageString(t *testing.T) {
	assert.Equal(t, "", RolesCliUsageString(nil))
	assert.Equal(t, "admin", RolesCliUsageString([]Role{RoleAdmin}))
	assert.Equal(t, "admin or guest", RolesCliUsageString([]Role{RoleAdmin, RoleGuest}))
	assert.Equal(t, "admin, manager, or guest", RolesCliUsageString([]Role{RoleAdmin, RoleManager, RoleGuest}))
}

func TestRoleStrings_Strings_ExcludesAliases(t *testing.T) {
	// The app→instance and uploader→contributor aliases and the visitor/empty roles
	// are display-only and must not appear in role listings, even though the map
	// still validates them.
	m := RoleStrings{
		string(RoleAdmin):       RoleAdmin,
		string(RoleContributor): RoleContributor,
		"uploader":              RoleContributor,
		"app":                   RoleInstance,
		string(RoleVisitor):     RoleVisitor,
		"":                      RoleNone,
	}
	got := m.Strings()
	assert.ElementsMatch(t, []string{"admin", "contributor"}, got)
	assert.NotContains(t, got, "uploader")
	assert.NotContains(t, got, "app")
	assert.NotContains(t, got, "visitor")
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
		assert.Regexp(t, ` or guest$`, u)
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
