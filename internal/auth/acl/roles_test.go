package acl

import (
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

	// Expect deterministic, sorted output and no empty entries.
	assert.Equal(t, []string{"admin", "guest", "visitor"}, got)
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
		m := RoleStrings{"visitor": RoleVisitor, "guest": RoleGuest, "admin": RoleAdmin}
		assert.Equal(t, "admin, guest, or visitor", m.CliUsageString())
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
	t.Run("ClientRolesStringsIncludeAliasNoneExcludeEmpty", func(t *testing.T) {
		got := ClientRoles.Strings()
		// Contains exactly the expected elements, order not enforced.
		assert.ElementsMatch(t, []string{"admin", "client", "instance", "none", "portal", "service"}, got)
		// Does not include empty string
		for _, s := range got {
			assert.NotEqual(t, "", s)
		}
	})
	t.Run("UserRolesStringsIncludeAliasNoneExcludeEmpty", func(t *testing.T) {
		got := UserRoles.Strings()
		assert.ElementsMatch(t, []string{"admin", "guest", "none", "visitor"}, got)
		for _, s := range got {
			assert.NotEqual(t, "", s)
		}
	})
	t.Run("ClientRolesCliUsageStringIncludesNoneAndOrBeforeLast", func(t *testing.T) {
		u := ClientRoles.CliUsageString()
		// Should list known roles and end with "or none" (alias present).
		for _, s := range []string{"admin", "client", "instance", "portal", "service", "none"} {
			assert.Contains(t, u, s)
		}
		assert.Regexp(t, `, or none$`, u)
	})
	t.Run("UserRolesCliUsageStringIncludesNoneAndOrBeforeLast", func(t *testing.T) {
		u := UserRoles.CliUsageString()
		for _, s := range []string{"admin", "guest", "visitor", "none"} {
			assert.Contains(t, u, s)
		}
		assert.Regexp(t, `, or none$`, u)
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
		found := false
		for _, have := range ResourceNames {
			if have == w {
				found = true
				break
			}
		}
		assert.Truef(t, found, "resource %s not found in ResourceNames", w)
	}
}
