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
