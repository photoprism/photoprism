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
