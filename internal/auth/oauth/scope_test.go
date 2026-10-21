package oauth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/auth/acl"
)

func TestValidateScopes(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		result, err := ValidateScopes(nil, "mcp")
		require.Nil(t, err)
		assert.Equal(t, []string{"mcp"}, result)
	})
	t.Run("AddsMCP", func(t *testing.T) {
		result, err := ValidateScopes([]string{"photos", "photos"}, "*")
		require.Nil(t, err)
		assert.Equal(t, []string{"mcp", "photos"}, result)
	})
	t.Run("Unsupported", func(t *testing.T) {
		_, err := ValidateScopes([]string{"mcp", "password"}, "*")
		require.NotNil(t, err)
		assert.Equal(t, ErrCodeInvalidScope, err.Code)
		_, err = ValidateScopes([]string{"unknown"}, "*")
		require.NotNil(t, err)
	})
	t.Run("AnyClientScope", func(t *testing.T) {
		result, err := ValidateScopes(AppScopes, "*")
		require.Nil(t, err)
		assert.Equal(t, AppScopes, result)
		_, err = ValidateScopes([]string{"mcp", "users"}, "*")
		require.NotNil(t, err)
		assert.Equal(t, ErrCodeInvalidScope, err.Code)
	})
	t.Run("ClientManagement", func(t *testing.T) {
		assert.NotContains(t, AppScopes, acl.ResourceClients.String())
		_, err := ValidateScopes([]string{"mcp", acl.ResourceClients.String()}, "*")
		require.NotNil(t, err)
		assert.Equal(t, ErrCodeInvalidScope, err.Code)
	})
	t.Run("NotAllowedForClient", func(t *testing.T) {
		_, err := ValidateScopes([]string{"mcp", "photos"}, "mcp")
		require.NotNil(t, err)
		assert.Equal(t, ErrCodeInvalidScope, err.Code)
		_, err = ValidateScopes(nil, "photos")
		require.NotNil(t, err)
	})
}

func TestClientScopes(t *testing.T) {
	t.Run("AnyClientScope", func(t *testing.T) {
		assert.Equal(t, []string{"mcp", "photos"}, ClientScopes([]string{"mcp", "photos"}, "*"))
	})
	t.Run("Narrowed", func(t *testing.T) {
		assert.Equal(t, []string{"mcp"}, ClientScopes([]string{"mcp", "photos"}, "mcp"))
	})
	t.Run("ReadVerb", func(t *testing.T) {
		assert.Equal(t, []string{"mcp", "photos"}, ClientScopes([]string{"mcp", "photos"}, "mcp read photos"))
	})
	t.Run("None", func(t *testing.T) {
		assert.Empty(t, ClientScopes([]string{"mcp", "photos"}, "albums"))
	})
	t.Run("NoLongerAppScope", func(t *testing.T) {
		assert.Equal(t, []string{"mcp"}, ClientScopes([]string{"mcp", "users"}, "*"))
	})
}

func TestScopeUsable(t *testing.T) {
	t.Run("Admin", func(t *testing.T) {
		assert.True(t, ScopeUsable("mcp", acl.RoleAdmin))
		assert.True(t, ScopeUsable("photos", acl.RoleAdmin))
	})
	t.Run("Guest", func(t *testing.T) {
		assert.False(t, ScopeUsable("mcp", acl.RoleGuest))
		assert.True(t, ScopeUsable("photos", acl.RoleGuest))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.False(t, ScopeUsable("", acl.RoleAdmin))
	})
}

func TestUsableScopes(t *testing.T) {
	t.Run("Admin", func(t *testing.T) {
		assert.Equal(t, []string{"mcp", "photos"}, UsableScopes([]string{"mcp", "photos"}, acl.RoleAdmin))
	})
	t.Run("GuestWithoutMCP", func(t *testing.T) {
		assert.Empty(t, UsableScopes([]string{"mcp", "photos"}, acl.RoleGuest))
	})
}

func TestJoinScopes(t *testing.T) {
	assert.Equal(t, "mcp photos", JoinScopes([]string{"mcp", "photos"}))
	assert.Equal(t, "", JoinScopes(nil))
}
