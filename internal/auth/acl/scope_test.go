package acl

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/list"
)

func TestGrantScopeRead(t *testing.T) {
	t.Run("ActionView", func(t *testing.T) {
		assert.True(t, GrantScopeRead.Allow(ActionView))
		assert.False(t, GrantScopeRead.DenyAny(Permissions{ActionView}))
	})
	t.Run("ActionUpdate", func(t *testing.T) {
		assert.False(t, GrantScopeRead.Allow(ActionUpdate))
		assert.True(t, GrantScopeRead.DenyAny(Permissions{ActionUpdate}))
	})
	t.Run("AccessAll", func(t *testing.T) {
		assert.True(t, GrantScopeRead.Allow(AccessAll))
		assert.False(t, GrantScopeRead.DenyAny(Permissions{AccessAll}))
	})
}

func TestGrantScopeWrite(t *testing.T) {
	t.Run("ActionView", func(t *testing.T) {
		assert.False(t, GrantScopeWrite.Allow(ActionView))
		assert.True(t, GrantScopeWrite.DenyAny(Permissions{ActionView}))
	})
	t.Run("ActionUpdate", func(t *testing.T) {
		assert.True(t, GrantScopeWrite.Allow(ActionUpdate))
		assert.False(t, GrantScopeWrite.DenyAny(Permissions{ActionUpdate}))
	})
	t.Run("AccessAll", func(t *testing.T) {
		assert.True(t, GrantScopeWrite.Allow(AccessAll))
		assert.False(t, GrantScopeWrite.DenyAny(Permissions{AccessAll}))
	})
}

func TestScopePermits(t *testing.T) {
	t.Run("AnyScope", func(t *testing.T) {
		assert.True(t, ScopePermits("*", "", nil))
	})
	t.Run("ReadScope", func(t *testing.T) {
		assert.True(t, ScopePermits("read", "metrics", nil))
		assert.True(t, ScopePermits("read", "sessions", nil))
		assert.True(t, ScopePermits("read", "metrics", Permissions{ActionView, AccessAll}))
		assert.False(t, ScopePermits("read", "metrics", Permissions{ActionUpdate}))
		assert.False(t, ScopePermits("read", "metrics", Permissions{ActionUpdate}))
		assert.False(t, ScopePermits("read", "settings", Permissions{ActionUpdate}))
		assert.False(t, ScopePermits("read", "settings", Permissions{ActionCreate}))
		assert.False(t, ScopePermits("read", "sessions", Permissions{ActionDelete}))
	})
	t.Run("ReadAny", func(t *testing.T) {
		assert.True(t, ScopePermits("read *", "metrics", nil))
		assert.True(t, ScopePermits("read *", "sessions", nil))
		assert.True(t, ScopePermits("read *", "metrics", Permissions{ActionView, AccessAll}))
		assert.False(t, ScopePermits("read *", "metrics", Permissions{ActionUpdate}))
		assert.False(t, ScopePermits("read *", "metrics", Permissions{ActionUpdate}))
		assert.False(t, ScopePermits("read *", "settings", Permissions{ActionUpdate}))
		assert.False(t, ScopePermits("read *", "settings", Permissions{ActionCreate}))
		assert.False(t, ScopePermits("read *", "sessions", Permissions{ActionDelete}))
	})
	t.Run("ReadSettings", func(t *testing.T) {
		assert.True(t, ScopePermits("read settings", "settings", Permissions{ActionView}))
		assert.False(t, ScopePermits("read settings", "metrics", nil))
		assert.False(t, ScopePermits("read settings", "sessions", nil))
		assert.False(t, ScopePermits("read settings", "metrics", Permissions{ActionView, AccessAll}))
		assert.False(t, ScopePermits("read settings", "metrics", Permissions{ActionUpdate}))
		assert.False(t, ScopePermits("read settings", "metrics", Permissions{ActionUpdate}))
		assert.False(t, ScopePermits("read settings", "settings", Permissions{ActionUpdate}))
		assert.False(t, ScopePermits("read settings", "sessions", Permissions{ActionDelete}))
		assert.False(t, ScopePermits("read settings", "sessions", Permissions{ActionDelete}))
	})
	t.Run("MCPScope", func(t *testing.T) {
		// "mcp" is the canonical scope token for ResourceMCP and is accepted both
		// standalone and combined with other scopes; unrelated scopes are denied.
		assert.True(t, ScopePermits("mcp", ResourceMCP, Permissions{ActionView}))
		assert.True(t, ScopePermits("mcp", ResourceMCP, Permissions{ActionSearch}))
		assert.True(t, ScopePermits("MCP", ResourceMCP, Permissions{ActionView}))
		assert.True(t, ScopePermits("mcp metrics", ResourceMCP, Permissions{ActionView}))
		assert.True(t, ScopePermits("read mcp", ResourceMCP, Permissions{ActionView}))
		assert.False(t, ScopePermits("read mcp", ResourceMCP, Permissions{ActionUpdate}))
		assert.False(t, ScopePermits("metrics", ResourceMCP, Permissions{ActionView}))
		// "mcp:read" is NOT an alias for "mcp read" - KeyValue.Parse treats
		// it as {key=mcp, value=read}, which Attr.Find("mcp") never matches
		// because the resource lookup expects a {mcp,true} entry. This
		// regression assertion locks in the documented behavior so the
		// parser does not silently start accepting the colon form.
		assert.False(t, ScopePermits("mcp:read", ResourceMCP, Permissions{ActionView}))
		assert.False(t, ScopePermits("mcp:read", ResourceMCP, Permissions{ActionUpdate}))
	})
}

func TestScopeAttr(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{name: "Empty", input: "", expected: nil},
		{name: "Lowercase", input: "read metrics", expected: []string{"metrics", "read"}},
		{name: "Uppercase", input: "READ SETTINGS", expected: []string{"read", "settings"}},
		{name: "WithNoise", input: "  Read\tSessions\nmetrics", expected: []string{"metrics", "read", "sessions"}},
		{name: "Deduplicates", input: "metrics metrics", expected: []string{"metrics"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			attr := ScopeAttr(tc.input)
			if len(tc.expected) == 0 {
				assert.Len(t, attr, 0)
				return
			}
			assert.ElementsMatch(t, tc.expected, attr.Strings())
		})
	}
}

func TestScopePermitsEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		scope    string
		resource Resource
		perms    Permissions
		want     bool
	}{
		{name: "EmptyScope", scope: "", resource: "metrics", perms: nil, want: false},
		{name: "OnlyInvalidChars", scope: "()", resource: "metrics", perms: nil, want: false},
		{name: "WildcardMixedOrder", scope: "* read metrics", resource: "metrics", perms: Permissions{ActionUpdate}, want: false},
		{name: "WildcardOverridesReadRestrictions", scope: "read metrics *", resource: "metrics", perms: Permissions{ActionDelete}, want: false},
		{name: "WildcardWithFalseValueIgnored", scope: "*:false read", resource: "metrics", perms: Permissions{ActionUpdate}, want: false},
		{name: "ExplicitFalseResource", scope: "metrics:false", resource: "metrics", perms: nil, want: false},
		{name: "ExplicitTrueResource", scope: "metrics:true", resource: "metrics", perms: nil, want: true},
		{name: "CaseInsensitiveScopeAndResource", scope: "READ SETTINGS", resource: Resource("Settings"), perms: Permissions{ActionView}, want: true},
		{name: "WhitespaceAndTabs", scope: "\tread\tsettings\n", resource: "settings", perms: Permissions{ActionView}, want: true},
		{name: "DefaultResourceRead", scope: "read default", resource: "", perms: Permissions{ActionView}, want: true},
		{name: "DefaultResourceUpdateDenied", scope: "read default", resource: "", perms: Permissions{ActionUpdate}, want: false},
		{name: "WriteAllowsMutation", scope: "write settings", resource: "settings", perms: Permissions{ActionUpdate}, want: true},
		{name: "WriteBlocksReadOnly", scope: "write settings", resource: "settings", perms: Permissions{ActionView}, want: false},
		{name: "ReadGrantAllowsAccessAll", scope: "read", resource: "metrics", perms: Permissions{AccessAll}, want: true},
		{name: "ReadGrantDeniesManage", scope: "read metrics", resource: "metrics", perms: Permissions{ActionManage}, want: false},
		{name: "WriteGrantAllowsManage", scope: "write metrics", resource: "metrics", perms: Permissions{ActionManage}, want: true},
		{name: "ResourceWildcard", scope: "metrics:*", resource: "metrics", perms: Permissions{ActionDelete}, want: true},
		{name: "GlobalWildcardWithoutRead", scope: "* metrics", resource: "metrics", perms: Permissions{ActionDelete}, want: true},
		{name: "ResourceWildcardWithRead", scope: "read metrics:*", resource: "metrics", perms: Permissions{ActionView}, want: true},
		{name: "ResourceWildcardWriteDenied", scope: "read metrics:*", resource: "metrics", perms: Permissions{ActionUpdate}, want: false},
		{name: "DuplicateAndNoise", scope: "  read   metrics metrics   ", resource: "metrics", perms: nil, want: true},
		{name: "FalseOverridesTrue", scope: "metrics metrics:false", resource: "metrics", perms: nil, want: false},
		{name: "CaseInsensitiveResourceLookup", scope: "read metrics", resource: Resource("METRICS"), perms: Permissions{ActionView}, want: true},
		{name: "MixedReadWriteConflict", scope: "read write settings", resource: "settings", perms: Permissions{ActionUpdate}, want: false},
		{name: "PermissionsEmptySlice", scope: "read metrics", resource: "metrics", perms: Permissions{}, want: true},
		{name: "SimpleNonReadScopeAllows", scope: "cluster vision", resource: "cluster", perms: nil, want: true},
		{name: "SimpleNonReadScopeRejectsMissing", scope: "cluster vision", resource: "portal", perms: nil, want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ScopePermits(tc.scope, tc.resource, tc.perms)
			assert.Equalf(t, tc.want, got, "scope %q resource %q perms %v", tc.scope, tc.resource, tc.perms)
		})
	}
}

func TestScopeAttrPermits(t *testing.T) {
	tests := []struct {
		name     string
		scope    string
		resource Resource
		perms    Permissions
		want     bool
	}{
		{name: "EmptyAttr", scope: "", resource: "metrics", perms: nil, want: false},
		{name: "Wildcard", scope: "*", resource: "metrics", perms: Permissions{ActionUpdate}, want: true},
		{name: "ReadAllowsView", scope: "read", resource: "settings", perms: Permissions{ActionView}, want: true},
		{name: "ReadBlocksUpdate", scope: "read", resource: "settings", perms: Permissions{ActionUpdate}, want: false},
		{name: "ResourceMismatch", scope: "read metrics", resource: "settings", perms: nil, want: false},
		{name: "WriteAllowsManage", scope: "write metrics", resource: "metrics", perms: Permissions{ActionManage}, want: true},
		{name: "WriteBlocksView", scope: "write metrics", resource: "metrics", perms: Permissions{ActionView}, want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			attr := ScopeAttr(tc.scope)
			got := ScopeAttrPermits(attr, tc.resource, tc.perms)
			assert.Equalf(t, tc.want, got, "scope %q resource %q perms %v", tc.scope, tc.resource, tc.perms)
		})
	}
}

func TestScopeAttrWithin(t *testing.T) {
	within := func(scope, allowed string) list.Attr {
		return ScopeAttrWithin(ScopeAttr(scope), ScopeAttr(allowed))
	}

	t.Run("AllowedWildcard", func(t *testing.T) {
		assert.Equal(t, "mcp photos", within("mcp photos", "*").String())
	})
	t.Run("Reduced", func(t *testing.T) {
		result := within("mcp photos albums", "config cluster vision metrics mcp users")
		assert.Equal(t, "mcp", result.String())
		assert.True(t, ScopeAttrPermits(result, ResourceMCP, Permissions{ActionView}))
		assert.False(t, ScopeAttrPermits(result, ResourcePhotos, Permissions{ActionSearch}))
	})
	t.Run("ScopeWildcard", func(t *testing.T) {
		result := within("*", "cluster mcp")
		assert.Equal(t, "cluster mcp", result.String())
		assert.False(t, ScopeAttrPermits(result, ResourcePhotos, Permissions{ActionView}))
	})
	t.Run("ReadOnlyScope", func(t *testing.T) {
		result := within("read", "mcp photos")
		assert.True(t, ScopeAttrPermits(result, ResourcePhotos, Permissions{ActionView}))
		assert.False(t, ScopeAttrPermits(result, ResourcePhotos, Permissions{ActionUpdate}))
		assert.False(t, ScopeAttrPermits(result, ResourceAlbums, Permissions{ActionView}))
	})
	t.Run("ScopeModifierKept", func(t *testing.T) {
		result := within("mcp read", "mcp")
		assert.True(t, ScopeAttrPermits(result, ResourceMCP, Permissions{ActionView}))
		assert.False(t, ScopeAttrPermits(result, ResourceMCP, Permissions{ActionUpdate}))
	})
	t.Run("AllowedModifierNarrows", func(t *testing.T) {
		result := within("mcp", "mcp read")
		assert.True(t, ScopeAttrPermits(result, ResourceMCP, Permissions{ActionView}))
		assert.False(t, ScopeAttrPermits(result, ResourceMCP, Permissions{ActionUpdate}))
	})
	t.Run("NegatedEntryKept", func(t *testing.T) {
		result := within("mcp photos:false", "mcp photos")
		assert.True(t, ScopeAttrPermits(result, ResourceMCP, Permissions{ActionView}))
		assert.False(t, ScopeAttrPermits(result, ResourcePhotos, Permissions{ActionView}))
	})
	t.Run("AllowedNegation", func(t *testing.T) {
		result := within("mcp photos", "mcp photos:false")
		assert.False(t, ScopeAttrPermits(result, ResourcePhotos, Permissions{ActionView}))
	})
	t.Run("NoOverlap", func(t *testing.T) {
		result := within("photos", "mcp")
		assert.Empty(t, result)
		assert.False(t, ScopeAttrPermits(result, ResourcePhotos, Permissions{ActionView}))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Empty(t, within("", "mcp"))
		assert.Empty(t, within("mcp", ""))
	})
}

func TestScopeAttrWithin_Shapes(t *testing.T) {
	within := func(scope, allowed string) list.Attr {
		return ScopeAttrWithin(list.ParseAttr(scope), list.ParseAttr(allowed))
	}

	t.Run("AllowedReadOnly", func(t *testing.T) {
		result := within("mcp", "read")
		assert.True(t, ScopeAttrPermits(result, ResourceMCP, Permissions{ActionView}))
		assert.False(t, ScopeAttrPermits(result, ResourceMCP, Permissions{ActionUpdate}))
		assert.False(t, ScopeAttrPermits(result, ResourcePhotos, Permissions{ActionView}))
	})
	t.Run("ConflictingModifiers", func(t *testing.T) {
		assert.False(t, ScopeAttrPermits(within("mcp read:false", "mcp read"), ResourceMCP, Permissions{ActionUpdate}))
		assert.False(t, ScopeAttrPermits(within("mcp write:false", "mcp write"), ResourceMCP, Permissions{ActionView}))
		assert.False(t, ScopeAttrPermits(within("mcp", "mcp read read:*"), ResourceMCP, Permissions{ActionUpdate}))
	})
	t.Run("DuplicateKeys", func(t *testing.T) {
		assert.False(t, ScopeAttrPermits(within("mcp photos:* photos:false", "mcp photos"), ResourcePhotos, Permissions{ActionView}))
		assert.False(t, ScopeAttrPermits(within("*", "mcp photos:false photos:*"), ResourcePhotos, Permissions{ActionView}))
	})
	t.Run("LongValue", func(t *testing.T) {
		long := strings.Repeat("a", 770)
		result := within("mcp mcp:"+long, "config cluster vision metrics mcp users")
		assert.Equal(t, "mcp", result.String())
		result = within("mcp read:"+long, "mcp write")
		assert.False(t, ScopeAttrPermits(ScopeAttr(result.String()), ResourceMCP, Permissions{ActionView}))
	})
	t.Run("LongKey", func(t *testing.T) {
		stored := func(scope, allowed string) string { return within(scope, allowed).String() }
		result := stored("mcp photos "+strings.Repeat("a", 756), "read")
		assert.False(t, ScopePermits(result, ResourcePhotos, Permissions{ActionUpdate}))
		assert.False(t, ScopePermits(result, ResourceMCP, Permissions{ActionUpdate}))
		result = stored("* mcp read "+strings.Repeat("s", 758)+":false", "* users:false")
		assert.False(t, ScopePermits(result, ResourceUsers, Permissions{ActionView}))
		result = stored("users write "+strings.Repeat("k", 755), "read")
		assert.False(t, ScopePermits(result, ResourceUsers, Permissions{ActionView}))
	})
	t.Run("ModifierOnlyScope", func(t *testing.T) {
		assert.Empty(t, within("write", "mcp"))
		assert.Empty(t, within("read:false", "mcp"))
		assert.Empty(t, within("photos read", "mcp"))
		assert.Empty(t, within("photos", "mcp read"))
	})
	t.Run("UnsortedInput", func(t *testing.T) {
		scope, allowed := list.ParseAttr("mcp photos photos:false"), list.ParseAttr("mcp photos")
		result := ScopeAttrWithin(scope, allowed)
		assert.False(t, ScopeAttrPermits(result, ResourcePhotos, Permissions{ActionView}))
		assert.Equal(t, "photos:false", scope[2].Key+":"+scope[2].Value)
	})
}

// TestScopeAttrWithin_Property checks that the result permits a resource and permissions only where both lists do,
// and exactly there when neither list repeats a key.
func TestScopeAttrWithin_Property(t *testing.T) {
	atoms := []string{"*", "read", "write", "mcp", "photos", "albums", "photos:false", "mcp:false", "read:false"}
	resources := []Resource{ResourceMCP, ResourcePhotos, ResourceAlbums, ResourceUsers}
	perms := []Permissions{nil, {ActionView}, {ActionUpdate}, {ActionView, ActionUpdate}}

	var lists []string
	for i := range atoms {
		lists = append(lists, atoms[i])
		for j := i + 1; j < len(atoms); j++ {
			lists = append(lists, atoms[i]+" "+atoms[j])
			for k := j + 1; k < len(atoms); k++ {
				lists = append(lists, atoms[i]+" "+atoms[j]+" "+atoms[k])
			}
		}
	}

	repeats := func(s string) bool {
		seen := map[string]bool{}
		for _, f := range strings.Fields(s) {
			k, _, _ := strings.Cut(f, ":")
			if seen[k] {
				return true
			}
			seen[k] = true
		}
		return false
	}

	for _, scope := range lists {
		for _, allowed := range lists {
			result := ScopeAttrWithin(ScopeAttr(scope), ScopeAttr(allowed)).String()
			for _, r := range resources {
				for _, p := range perms {
					got := ScopePermits(result, r, p)
					want := ScopePermits(scope, r, p) && ScopePermits(allowed, r, p)
					if got && !want {
						t.Fatalf("widened: scope=%q allowed=%q result=%q resource=%s perms=%v", scope, allowed, result, r, p)
					} else if want && !got && !repeats(scope) && !repeats(allowed) {
						t.Fatalf("narrowed: scope=%q allowed=%q result=%q resource=%s perms=%v", scope, allowed, result, r, p)
					}
				}
			}
		}
	}
}
