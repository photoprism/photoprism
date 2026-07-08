package cluster

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeNodeRole(t *testing.T) {
	t.Run("Instance", func(t *testing.T) {
		assert.Equal(t, RoleInstance, NormalizeNodeRole("instance"))
	})
	t.Run("LegacyAliasAppToInstance", func(t *testing.T) {
		assert.Equal(t, RoleInstance, NormalizeNodeRole(" app "))
	})
	t.Run("Portal", func(t *testing.T) {
		assert.Equal(t, RolePortal, NormalizeNodeRole("portal"))
	})
	t.Run("Service", func(t *testing.T) {
		assert.Equal(t, RoleService, NormalizeNodeRole("service"))
	})
	t.Run("Invalid", func(t *testing.T) {
		assert.Equal(t, NodeRole(""), NormalizeNodeRole("unknown"))
	})
}
