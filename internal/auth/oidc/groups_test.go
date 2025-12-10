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
