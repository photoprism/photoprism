package acl

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
