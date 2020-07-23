package acl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestACL_Allow(t *testing.T) {
	t.Run("photos/admin/update", func(t *testing.T) {
		assert.True(t, Permissions.Allow(ResourcePhotos, RoleAdmin, ActionUpdate))
	})
	t.Run("default/admin", func(t *testing.T) {
		assert.True(t, Permissions.Allow(ResourceDefault, RoleAdmin, ActionDefault))
	})
	t.Run("default/guest", func(t *testing.T) {
		assert.False(t, Permissions.Allow(ResourceDefault, RoleGuest, ActionDefault))
	})
	t.Run("photos/guest/search", func(t *testing.T) {
		assert.True(t, Permissions.Allow(ResourcePhotos, RoleGuest, ActionSearch))
	})
	t.Run("photos/guest/default", func(t *testing.T) {
		assert.False(t, Permissions.Allow(ResourcePhotos, RoleGuest, ActionDefault))
	})
	t.Run("albums/guest/search", func(t *testing.T) {
		assert.True(t, Permissions.Allow(ResourceAlbums, RoleGuest, ActionSearch))
	})
	t.Run("albums/guest/default", func(t *testing.T) {
		assert.False(t, Permissions.Allow(ResourceAlbums, RoleGuest, ActionDefault))
	})
}

func TestACL_Deny(t *testing.T) {
	t.Run("default/admin", func(t *testing.T) {
		assert.False(t, Permissions.Deny(ResourceDefault, RoleAdmin, ActionDefault))
	})
	t.Run("default/guest", func(t *testing.T) {
		assert.True(t, Permissions.Deny(ResourceDefault, RoleGuest, ActionDefault))
	})
	t.Run("photos/guest/search", func(t *testing.T) {
		assert.False(t, Permissions.Deny(ResourcePhotos, RoleGuest, ActionSearch))
	})
	t.Run("photos/guest/default", func(t *testing.T) {
		assert.True(t, Permissions.Deny(ResourcePhotos, RoleGuest, ActionDefault))
	})
	t.Run("albums/guest/search", func(t *testing.T) {
		assert.False(t, Permissions.Deny(ResourceAlbums, RoleGuest, ActionSearch))
	})
	t.Run("albums/guest/default", func(t *testing.T) {
		assert.True(t, Permissions.Deny(ResourceAlbums, RoleGuest, ActionDefault))
	})
}
