package acl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestACL_Allow(t *testing.T) {
	t.Run("ResourceSessions", func(t *testing.T) {
		assert.True(t, Rules.Allow(ResourceSessions, RoleAdmin, AccessOwn))
		assert.True(t, Rules.Allow(ResourceSessions, RoleAdmin, ActionManageOwn))
		assert.True(t, Rules.Allow(ResourceSessions, RoleAdmin, AccessOwn))
		assert.False(t, Rules.Allow(ResourceSessions, RoleVisitor, AccessAll))
		assert.True(t, Rules.Allow(ResourceSessions, RoleVisitor, AccessOwn))
		assert.False(t, Rules.Allow(ResourceSessions, RoleClient, AccessAll))
		assert.True(t, Rules.Allow(ResourceSessions, RoleClient, AccessOwn))
	})
	t.Run("ResourcePhotosRoleAdminActionModify", func(t *testing.T) {
		assert.True(t, Rules.Allow(ResourcePhotos, RoleAdmin, ActionUpdate))
	})
	t.Run("ResourceDefaultRoleAdminActionDefault", func(t *testing.T) {
		assert.True(t, Rules.Allow(ResourceDefault, RoleAdmin, FullAccess))
	})
	t.Run("ResourceDefaultRoleVisitorActionDefault", func(t *testing.T) {
		assert.False(t, Rules.Allow(ResourceDefault, RoleVisitor, FullAccess))
	})
	t.Run("ResourcePhotosRoleVisitorActionDefault", func(t *testing.T) {
		assert.False(t, Rules.Allow(ResourcePhotos, RoleVisitor, FullAccess))
	})
	t.Run("ResourceAlbumsRoleVisitorAccessShared", func(t *testing.T) {
		assert.True(t, Rules.Allow(ResourceAlbums, RoleVisitor, AccessShared))
	})
	t.Run("ResourceAlbumsRoleVisitorActionDefault", func(t *testing.T) {
		assert.False(t, Rules.Allow(ResourceAlbums, RoleVisitor, FullAccess))
	})
	t.Run("WrongResourceRoleAdminActionDefault", func(t *testing.T) {
		assert.True(t, Rules.Allow("wrong", RoleAdmin, FullAccess))
	})
	t.Run("WrongResourceRoleVisitorActionDefault", func(t *testing.T) {
		assert.False(t, Rules.Allow("wrong", RoleVisitor, FullAccess))
	})
}

func TestACL_AllowAny(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.False(t, Rules.AllowAny(ResourceAlbums, RoleVisitor, Permissions{}))
	})
	t.Run("VisitorAccess", func(t *testing.T) {
		assert.True(t, Rules.AllowAny(ResourceAlbums, RoleVisitor, Permissions{AccessAll, AccessShared}))
		assert.True(t, Rules.AllowAny(ResourceAlbums, RoleVisitor, Permissions{AccessShared}))
		assert.False(t, Rules.AllowAny(ResourceAlbums, RoleVisitor, Permissions{AccessAll}))
	})
	t.Run("ResourcePhotosRoleAdminActionModify", func(t *testing.T) {
		assert.True(t, Rules.AllowAny(ResourcePhotos, RoleAdmin, Permissions{ActionUpdate}))
	})
	t.Run("ResourceDefaultRoleAdminActionDefault", func(t *testing.T) {
		assert.True(t, Rules.AllowAny(ResourceDefault, RoleAdmin, Permissions{FullAccess}))
	})
	t.Run("ResourceDefaultRoleVisitorActionDefault", func(t *testing.T) {
		assert.False(t, Rules.AllowAny(ResourceDefault, RoleVisitor, Permissions{FullAccess}))
	})
	t.Run("ResourcePhotosRoleVisitorActionDefault", func(t *testing.T) {
		assert.False(t, Rules.AllowAny(ResourcePhotos, RoleVisitor, Permissions{FullAccess}))
	})
	t.Run("ResourceAlbumsRoleVisitorAccessShared", func(t *testing.T) {
		assert.True(t, Rules.AllowAny(ResourceAlbums, RoleVisitor, Permissions{AccessShared}))
	})
	t.Run("ResourceAlbumsRoleVisitorActionDefault", func(t *testing.T) {
		assert.False(t, Rules.AllowAny(ResourceAlbums, RoleVisitor, Permissions{FullAccess}))
	})
}

func TestACL_AllowAll(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.False(t, Rules.AllowAll(ResourceAlbums, RoleVisitor, Permissions{}))
	})
	t.Run("VisitorAccess", func(t *testing.T) {
		assert.False(t, Rules.AllowAll(ResourceAlbums, RoleVisitor, Permissions{AccessAll, AccessShared}))
		assert.True(t, Rules.AllowAll(ResourceAlbums, RoleVisitor, Permissions{AccessShared}))
		assert.False(t, Rules.AllowAll(ResourceAlbums, RoleVisitor, Permissions{AccessAll}))
	})
	t.Run("ResourcePhotosRoleAdminActionModify", func(t *testing.T) {
		assert.True(t, Rules.AllowAll(ResourcePhotos, RoleAdmin, Permissions{ActionUpdate}))
	})
	t.Run("ResourceDefaultRoleAdminActionDefault", func(t *testing.T) {
		assert.True(t, Rules.AllowAll(ResourceDefault, RoleAdmin, Permissions{FullAccess}))
	})
	t.Run("ResourceDefaultRoleVisitorActionDefault", func(t *testing.T) {
		assert.False(t, Rules.AllowAll(ResourceDefault, RoleVisitor, Permissions{FullAccess}))
	})
	t.Run("ResourcePhotosRoleVisitorActionDefault", func(t *testing.T) {
		assert.False(t, Rules.AllowAll(ResourcePhotos, RoleVisitor, Permissions{FullAccess}))
	})
	t.Run("ResourceAlbumsRoleVisitorAccessShared", func(t *testing.T) {
		assert.True(t, Rules.AllowAll(ResourceAlbums, RoleVisitor, Permissions{AccessShared}))
	})
	t.Run("ResourceAlbumsRoleVisitorActionDefault", func(t *testing.T) {
		assert.False(t, Rules.AllowAll(ResourceAlbums, RoleVisitor, Permissions{FullAccess}))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.False(t, Rules.AllowAll(ResourceAlbums, RoleVisitor, Permissions{}))
	})
}

func TestACL_Deny(t *testing.T) {
	t.Run("ResourceDefaultRoleAdminActionDefault", func(t *testing.T) {
		assert.False(t, Rules.Deny(ResourceDefault, RoleAdmin, FullAccess))
	})
	t.Run("ResourceDefaultRoleVisitorActionDefault", func(t *testing.T) {
		assert.True(t, Rules.Deny(ResourceDefault, RoleVisitor, FullAccess))
	})
	t.Run("ResourceAlbumsRoleVisitorActionAccessShared", func(t *testing.T) {
		assert.False(t, Rules.Deny(ResourceAlbums, RoleVisitor, AccessShared))
	})
	t.Run("ResourcePhotosRoleVisitorActionDefault", func(t *testing.T) {
		assert.True(t, Rules.Deny(ResourcePhotos, RoleVisitor, FullAccess))
	})
	t.Run("ResourceAlbumsRoleVisitorActionDefault", func(t *testing.T) {
		assert.True(t, Rules.Deny(ResourceAlbums, RoleVisitor, FullAccess))
	})
}

func TestACL_DenyAll(t *testing.T) {
	t.Run("ResourceFilesRoleVisitorActionDefault", func(t *testing.T) {
		assert.True(t, Rules.DenyAll(ResourceFiles, RoleVisitor, Permissions{FullAccess, AccessShared, ActionView}))
	})
	t.Run("ResourceFilesRoleAdminActionDefault", func(t *testing.T) {
		assert.False(t, Rules.DenyAll(ResourceFiles, RoleAdmin, Permissions{FullAccess, AccessShared, ActionView}))
	})
}

func TestACL_ResourceMCP(t *testing.T) {
	t.Run("AdminActionView", func(t *testing.T) {
		assert.True(t, Rules.Allow(ResourceMCP, RoleAdmin, ActionView))
	})
	t.Run("AdminActionManage", func(t *testing.T) {
		assert.True(t, Rules.Allow(ResourceMCP, RoleAdmin, ActionManage))
	})
	t.Run("AdminFullAccess", func(t *testing.T) {
		assert.True(t, Rules.Allow(ResourceMCP, RoleAdmin, FullAccess))
	})
	t.Run("InstanceSearchAndView", func(t *testing.T) {
		assert.True(t, Rules.Allow(ResourceMCP, RoleInstance, ActionView))
		assert.True(t, Rules.Allow(ResourceMCP, RoleInstance, ActionSearch))
	})
	t.Run("InstanceManageDenied", func(t *testing.T) {
		assert.True(t, Rules.DenyAll(ResourceMCP, RoleInstance, Permissions{ActionManage, ActionUpdate, ActionDelete, FullAccess}))
	})
	t.Run("ServiceSearchAndView", func(t *testing.T) {
		assert.True(t, Rules.Allow(ResourceMCP, RoleService, ActionView))
		assert.True(t, Rules.Allow(ResourceMCP, RoleService, ActionSearch))
	})
	t.Run("ServiceManageDenied", func(t *testing.T) {
		assert.True(t, Rules.DenyAll(ResourceMCP, RoleService, Permissions{ActionManage, ActionUpdate, ActionDelete, FullAccess}))
	})
	t.Run("PortalSearchAndView", func(t *testing.T) {
		assert.True(t, Rules.Allow(ResourceMCP, RolePortal, ActionView))
		assert.True(t, Rules.Allow(ResourceMCP, RolePortal, ActionSearch))
	})
	t.Run("PortalManageDenied", func(t *testing.T) {
		assert.True(t, Rules.DenyAll(ResourceMCP, RolePortal, Permissions{ActionManage, ActionUpdate, ActionDelete, FullAccess}))
	})
	t.Run("ClientSearchAndView", func(t *testing.T) {
		assert.True(t, Rules.Allow(ResourceMCP, RoleClient, ActionView))
		assert.True(t, Rules.Allow(ResourceMCP, RoleClient, ActionSearch))
	})
	t.Run("ClientManageDenied", func(t *testing.T) {
		assert.True(t, Rules.DenyAll(ResourceMCP, RoleClient, Permissions{ActionManage, ActionUpdate, ActionDelete, FullAccess}))
	})
	t.Run("GuestDenied", func(t *testing.T) {
		assert.True(t, Rules.DenyAll(ResourceMCP, RoleGuest, Permissions{ActionView, ActionSearch, ActionManage}))
	})
	t.Run("VisitorDenied", func(t *testing.T) {
		assert.True(t, Rules.DenyAll(ResourceMCP, RoleVisitor, Permissions{ActionView, ActionSearch, ActionManage}))
	})
	t.Run("DefaultDenied", func(t *testing.T) {
		assert.True(t, Rules.DenyAll(ResourceMCP, RoleDefault, Permissions{ActionView, ActionSearch, ActionManage}))
	})
}

func TestACL_Resources(t *testing.T) {
	t.Run("Rules", func(t *testing.T) {
		result := Rules.Resources()
		assert.Len(t, result, len(ResourceNames)-1)
	})
}
