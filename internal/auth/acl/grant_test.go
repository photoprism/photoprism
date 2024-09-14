package acl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGrant_Allow(t *testing.T) {
	t.Run("GrantFullAccessAll", func(t *testing.T) {
		assert.True(t, GrantFullAccess.Allow(AccessAll))
	})
	t.Run("GrantFullAccessDownload", func(t *testing.T) {
		assert.True(t, GrantFullAccess.Allow(ActionDownload))
	})
	t.Run("GrantFullAccessDelete", func(t *testing.T) {
		assert.True(t, GrantFullAccess.Allow(ActionDelete))
	})
	t.Run("GrantFullAccessLibrary", func(t *testing.T) {
		assert.True(t, GrantFullAccess.Allow(AccessLibrary))
	})
	t.Run("UnknownAction", func(t *testing.T) {
		assert.False(t, GrantViewAll.Allow("lovecats"))
	})
	t.Run("ViewAllView", func(t *testing.T) {
		assert.True(t, GrantViewAll.Allow(ActionView))
	})
	t.Run("ViewAllAccessAll", func(t *testing.T) {
		assert.True(t, GrantViewAll.Allow(AccessAll))
	})
	t.Run("ViewAllDownload", func(t *testing.T) {
		assert.False(t, GrantViewAll.Allow(ActionDownload))
	})
	t.Run("ViewAllShare", func(t *testing.T) {
		assert.False(t, GrantViewAll.Allow(ActionShare))
	})
	t.Run("ViewOwnShare", func(t *testing.T) {
		assert.False(t, GrantViewOwn.Allow(ActionShare))
	})
	t.Run("ViewOwnView", func(t *testing.T) {
		assert.True(t, GrantViewOwn.Allow(ActionView))
	})
	t.Run("ViewOwnAccessAll", func(t *testing.T) {
		assert.False(t, GrantViewOwn.Allow(AccessAll))
	})
	t.Run("ViewOwnAccessOwn", func(t *testing.T) {
		assert.True(t, GrantViewOwn.Allow(AccessOwn))
	})
}

func TestGrant_DenyAny(t *testing.T) {
	t.Run("GrantFullAccessAll", func(t *testing.T) {
		assert.False(t, GrantFullAccess.DenyAny(Permissions{AccessAll}))
	})
	t.Run("GrantFullAccessDownload", func(t *testing.T) {
		assert.False(t, GrantFullAccess.DenyAny(Permissions{ActionDownload}))
	})
	t.Run("GrantFullAccessDelete", func(t *testing.T) {
		assert.False(t, GrantFullAccess.DenyAny(Permissions{ActionDelete}))
	})
	t.Run("GrantFullAccessLibrary", func(t *testing.T) {
		assert.False(t, GrantFullAccess.DenyAny(Permissions{AccessLibrary}))
	})
	t.Run("UnknownAction", func(t *testing.T) {
		assert.True(t, GrantViewAll.DenyAny(Permissions{"lovecats"}))
	})
	t.Run("ViewAllView", func(t *testing.T) {
		assert.False(t, GrantViewAll.DenyAny(Permissions{ActionView}))
	})
	t.Run("ViewAllAccessAll", func(t *testing.T) {
		assert.False(t, GrantViewAll.DenyAny(Permissions{AccessAll}))
	})
	t.Run("ViewAllDownload", func(t *testing.T) {
		assert.True(t, GrantViewAll.DenyAny(Permissions{ActionDownload}))
	})
	t.Run("ViewAllShare", func(t *testing.T) {
		assert.True(t, GrantViewAll.DenyAny(Permissions{ActionShare}))
	})
	t.Run("ViewOwnShare", func(t *testing.T) {
		assert.True(t, GrantViewOwn.DenyAny(Permissions{ActionShare}))
	})
	t.Run("ViewOwnView", func(t *testing.T) {
		assert.False(t, GrantViewOwn.DenyAny(Permissions{ActionView}))
	})
	t.Run("ViewOwnAccessAll", func(t *testing.T) {
		assert.True(t, GrantViewOwn.DenyAny(Permissions{AccessAll}))
	})
	t.Run("ViewOwnAccessOwn", func(t *testing.T) {
		assert.False(t, GrantViewOwn.DenyAny(Permissions{AccessOwn}))
	})
}
