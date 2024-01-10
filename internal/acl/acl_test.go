package acl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestACL_Allow(t *testing.T) {
	t.Run("ResourcePhotosRoleAdminActionModify", func(t *testing.T) {
		assert.True(t, Resources.Allow(ResourcePhotos, RoleAdmin, ActionUpdate))
	})
	t.Run("ResourceDefaultRoleAdminActionDefault", func(t *testing.T) {
		assert.True(t, Resources.Allow(ResourceDefault, RoleAdmin, FullAccess))
	})
	t.Run("ResourceDefaultRoleVisitorActionDefault", func(t *testing.T) {
		assert.False(t, Resources.Allow(ResourceDefault, RoleVisitor, FullAccess))
	})
	t.Run("ResourcePhotosRoleVisitorActionDefault", func(t *testing.T) {
		assert.False(t, Resources.Allow(ResourcePhotos, RoleVisitor, FullAccess))
	})
	t.Run("ResourceAlbumsRoleVisitorAccessShared", func(t *testing.T) {
		assert.True(t, Resources.Allow(ResourceAlbums, RoleVisitor, AccessShared))
	})
	t.Run("ResourceAlbumsRoleVisitorActionDefault", func(t *testing.T) {
		assert.False(t, Resources.Allow(ResourceAlbums, RoleVisitor, FullAccess))
	})
	t.Run("WrongResourceRoleAdminActionDefault", func(t *testing.T) {
		assert.True(t, Resources.Allow("wrong", RoleAdmin, FullAccess))
	})
	t.Run("WrongResourceRoleVisitorActionDefault", func(t *testing.T) {
		assert.False(t, Resources.Allow("wrong", RoleVisitor, FullAccess))
	})
}

func TestACL_AllowAny(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.False(t, Resources.AllowAny(ResourceAlbums, RoleVisitor, Permissions{}))
	})
	t.Run("VisitorAccess", func(t *testing.T) {
		assert.True(t, Resources.AllowAny(ResourceAlbums, RoleVisitor, Permissions{AccessAll, AccessShared}))
		assert.True(t, Resources.AllowAny(ResourceAlbums, RoleVisitor, Permissions{AccessShared}))
		assert.False(t, Resources.AllowAny(ResourceAlbums, RoleVisitor, Permissions{AccessAll}))
	})
	t.Run("ResourcePhotosRoleAdminActionModify", func(t *testing.T) {
		assert.True(t, Resources.AllowAny(ResourcePhotos, RoleAdmin, Permissions{ActionUpdate}))
	})
	t.Run("ResourceDefaultRoleAdminActionDefault", func(t *testing.T) {
		assert.True(t, Resources.AllowAny(ResourceDefault, RoleAdmin, Permissions{FullAccess}))
	})
	t.Run("ResourceDefaultRoleVisitorActionDefault", func(t *testing.T) {
		assert.False(t, Resources.AllowAny(ResourceDefault, RoleVisitor, Permissions{FullAccess}))
	})
	t.Run("ResourcePhotosRoleVisitorActionDefault", func(t *testing.T) {
		assert.False(t, Resources.AllowAny(ResourcePhotos, RoleVisitor, Permissions{FullAccess}))
	})
	t.Run("ResourceAlbumsRoleVisitorAccessShared", func(t *testing.T) {
		assert.True(t, Resources.AllowAny(ResourceAlbums, RoleVisitor, Permissions{AccessShared}))
	})
	t.Run("ResourceAlbumsRoleVisitorActionDefault", func(t *testing.T) {
		assert.False(t, Resources.AllowAny(ResourceAlbums, RoleVisitor, Permissions{FullAccess}))
	})
}

func TestACL_AllowAll(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.False(t, Resources.AllowAll(ResourceAlbums, RoleVisitor, Permissions{}))
	})
	t.Run("VisitorAccess", func(t *testing.T) {
		assert.False(t, Resources.AllowAll(ResourceAlbums, RoleVisitor, Permissions{AccessAll, AccessShared}))
		assert.True(t, Resources.AllowAll(ResourceAlbums, RoleVisitor, Permissions{AccessShared}))
		assert.False(t, Resources.AllowAll(ResourceAlbums, RoleVisitor, Permissions{AccessAll}))
	})
	t.Run("ResourcePhotosRoleAdminActionModify", func(t *testing.T) {
		assert.True(t, Resources.AllowAll(ResourcePhotos, RoleAdmin, Permissions{ActionUpdate}))
	})
	t.Run("ResourceDefaultRoleAdminActionDefault", func(t *testing.T) {
		assert.True(t, Resources.AllowAll(ResourceDefault, RoleAdmin, Permissions{FullAccess}))
	})
	t.Run("ResourceDefaultRoleVisitorActionDefault", func(t *testing.T) {
		assert.False(t, Resources.AllowAll(ResourceDefault, RoleVisitor, Permissions{FullAccess}))
	})
	t.Run("ResourcePhotosRoleVisitorActionDefault", func(t *testing.T) {
		assert.False(t, Resources.AllowAll(ResourcePhotos, RoleVisitor, Permissions{FullAccess}))
	})
	t.Run("ResourceAlbumsRoleVisitorAccessShared", func(t *testing.T) {
		assert.True(t, Resources.AllowAll(ResourceAlbums, RoleVisitor, Permissions{AccessShared}))
	})
	t.Run("ResourceAlbumsRoleVisitorActionDefault", func(t *testing.T) {
		assert.False(t, Resources.AllowAll(ResourceAlbums, RoleVisitor, Permissions{FullAccess}))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.False(t, Resources.AllowAll(ResourceAlbums, RoleVisitor, Permissions{}))
	})
}

func TestACL_Deny(t *testing.T) {
	t.Run("ResourceDefaultRoleAdminActionDefault", func(t *testing.T) {
		assert.False(t, Resources.Deny(ResourceDefault, RoleAdmin, FullAccess))
	})
	t.Run("ResourceDefaultRoleVisitorActionDefault", func(t *testing.T) {
		assert.True(t, Resources.Deny(ResourceDefault, RoleVisitor, FullAccess))
	})
	t.Run("ResourceAlbumsRoleVisitorActionAccessShared", func(t *testing.T) {
		assert.False(t, Resources.Deny(ResourceAlbums, RoleVisitor, AccessShared))
	})
	t.Run("ResourcePhotosRoleVisitorActionDefault", func(t *testing.T) {
		assert.True(t, Resources.Deny(ResourcePhotos, RoleVisitor, FullAccess))
	})
	t.Run("ResourceAlbumsRoleVisitorActionDefault", func(t *testing.T) {
		assert.True(t, Resources.Deny(ResourceAlbums, RoleVisitor, FullAccess))
	})
}

func TestACL_DenyAll(t *testing.T) {
	t.Run("ResourceFilesRoleVisitorActionDefault", func(t *testing.T) {
		assert.True(t, Resources.DenyAll(ResourceFiles, RoleVisitor, Permissions{FullAccess, AccessShared, ActionView}))
	})
	t.Run("ResourceFilesRoleAdminActionDefault", func(t *testing.T) {
		assert.False(t, Resources.DenyAll(ResourceFiles, RoleAdmin, Permissions{FullAccess, AccessShared, ActionView}))
	})
}

func TestACL_Resources(t *testing.T) {
	t.Run("Resources", func(t *testing.T) {
		result := Resources.Resources()
		assert.Len(t, result, 21)
	})
}
