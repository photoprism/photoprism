package search

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/authn"
)

func TestSessionGrantsPhotos(t *testing.T) {
	t.Run("NilUnrestricted", func(t *testing.T) {
		assert.True(t, sessionGrantsPhotos(nil, acl.AccessPrivate))
		assert.True(t, sessionGrantsPhotos(nil, acl.ActionDelete))
	})
	t.Run("AdminGranted", func(t *testing.T) {
		s := scopeSession("alice")
		assert.True(t, sessionGrantsPhotos(s, acl.AccessPrivate))
		assert.True(t, sessionGrantsPhotos(s, acl.ActionDelete))
		assert.True(t, sessionGrantsPhotos(s, acl.AccessLibrary))
	})
	t.Run("GuestDeniedPrivate", func(t *testing.T) {
		s := scopeSession("guest")
		assert.False(t, sessionGrantsPhotos(s, acl.AccessPrivate))
		assert.False(t, sessionGrantsPhotos(s, acl.ActionDelete))
	})
	t.Run("ClientRoleLimitsPrivilegedUser", func(t *testing.T) {
		// A restricted client role limits access even when a privileged user is attached, so a
		// client cannot inherit a library user's whole-library reach (the client and user roles
		// are intersected).
		client := &entity.Client{ClientRole: acl.RoleInstance.String(), AuthProvider: authn.ProviderClient.String()}
		s := &entity.Session{}
		s.SetClient(client)
		s.SetUser(entity.UserFixtures.Pointer("alice")) // admin user
		assert.True(t, s.IsClient())
		assert.Equal(t, acl.RoleInstance, s.GetClientRole())
		// RoleInstance (GrantSearchShared) lacks AccessPrivate, so the intersection denies it even
		// though the admin user alone would grant it.
		assert.False(t, sessionGrantsPhotos(s, acl.AccessPrivate))
		assert.False(t, sessionGrantsAnyPhotos(s, acl.Permissions{acl.AccessAll, acl.AccessLibrary}))
	})
}

func TestSessionGrantsAnyPhotos(t *testing.T) {
	t.Run("NilTrue", func(t *testing.T) {
		assert.True(t, sessionGrantsAnyPhotos(nil, acl.Permissions{acl.AccessAll, acl.AccessLibrary}))
	})
	t.Run("AdminLibrary", func(t *testing.T) {
		assert.True(t, sessionGrantsAnyPhotos(scopeSession("alice"), acl.Permissions{acl.AccessAll, acl.AccessLibrary}))
	})
	t.Run("GuestNoLibrary", func(t *testing.T) {
		assert.False(t, sessionGrantsAnyPhotos(scopeSession("guest"), acl.Permissions{acl.AccessAll, acl.AccessLibrary}))
	})
}

// Fixture identifiers used by the scope tests. In the public repo only the admin,
// guest, and visitor roles are available, so these tests cover the admin (full access)
// and guest (shared-only) paths; viewer/user behavior is validated in the
// edition-specific repositories.
const (
	scopeNormalPhotoUID  = "ps6sg6be2lvl0yh7"                         // not private, not archived
	scopePrivatePhotoUID = "ps6sg6be2lvl0y13"                         // "Photo06", private
	scopeNormalFileHash  = "2cad9168fa6acc5c5c2965ddf6ec465ca42fd818" // file of a non-private photo
	scopePrivateFileHash = "pcad9a68fa6acc5c5ba965adf6ec465ca42fd917" // "Photo06.png", private photo
)

// scopeSession builds an in-memory session for the named user fixture.
func scopeSession(name string) *entity.Session {
	s := &entity.Session{}
	s.SetUser(entity.UserFixtures.Pointer(name))
	return s
}

func TestPhotoSessionSeesEverything(t *testing.T) {
	t.Run("NilSession", func(t *testing.T) {
		assert.True(t, PhotoSessionSeesEverything(nil))
	})
	t.Run("Admin", func(t *testing.T) {
		assert.True(t, PhotoSessionSeesEverything(scopeSession("alice")))
	})
	t.Run("Guest", func(t *testing.T) {
		assert.False(t, PhotoSessionSeesEverything(scopeSession("guest")))
	})
}

func TestScopePhotosForSession(t *testing.T) {
	t.Run("NilUnchanged", func(t *testing.T) {
		base := UnscopedDb().Table("photos")
		assert.Same(t, base, ScopePhotosForSession(base, nil))
	})
	t.Run("AdminUnchanged", func(t *testing.T) {
		base := UnscopedDb().Table("photos")
		assert.Same(t, base, ScopePhotosForSession(base, scopeSession("alice")))
	})
	t.Run("GuestScoped", func(t *testing.T) {
		base := UnscopedDb().Table("photos")
		scoped := ScopePhotosForSession(base, scopeSession("guest"))
		assert.NotSame(t, base, scoped)
		var count int
		assert.NoError(t, scoped.Count(&count).Error)
	})
}

func TestScopeVisiblePhotos(t *testing.T) {
	t.Run("AdminSeesPrivate", func(t *testing.T) {
		var count int
		err := ScopeVisiblePhotos(UnscopedDb().Table("photos").Where("photos.photo_uid = ?", scopePrivatePhotoUID), scopeSession("alice")).Count(&count).Error
		assert.NoError(t, err)
		assert.Equal(t, 1, count)
	})
	t.Run("GuestDeniedPrivate", func(t *testing.T) {
		var count int
		err := ScopeVisiblePhotos(UnscopedDb().Table("photos").Where("photos.photo_uid = ?", scopePrivatePhotoUID), scopeSession("guest")).Count(&count).Error
		assert.NoError(t, err)
		assert.Equal(t, 0, count)
	})
}

func TestPhotoVisibleToSession(t *testing.T) {
	t.Run("EmptyUID", func(t *testing.T) {
		ok, err := PhotoVisibleToSession("", scopeSession("guest"))
		assert.NoError(t, err)
		assert.False(t, ok)
	})
	t.Run("NilSession", func(t *testing.T) {
		ok, err := PhotoVisibleToSession(scopePrivatePhotoUID, nil)
		assert.NoError(t, err)
		assert.True(t, ok)
	})
	t.Run("AdminPrivate", func(t *testing.T) {
		ok, err := PhotoVisibleToSession(scopePrivatePhotoUID, scopeSession("alice"))
		assert.NoError(t, err)
		assert.True(t, ok)
	})
	t.Run("GuestDeniedPrivate", func(t *testing.T) {
		ok, err := PhotoVisibleToSession(scopePrivatePhotoUID, scopeSession("guest"))
		assert.NoError(t, err)
		assert.False(t, ok)
	})
	t.Run("GuestDeniedUnshared", func(t *testing.T) {
		ok, err := PhotoVisibleToSession(scopeNormalPhotoUID, scopeSession("guest"))
		assert.NoError(t, err)
		assert.False(t, ok)
	})
}

// A non-empty Scope skips the personal ScopePhotosForSession filter, so it must be authorized by
// the album ownership/share gate in searchPhotos: a restricted session may scope only to an album
// it owns or has shared, never to an arbitrary album.
func TestUserPhotos_ScopeAuthorization(t *testing.T) {
	const album = "as6sg6bxpogaaba8" // manual album owned by admin, not shared with guests

	t.Run("GuestNonSharedScopeForbidden", func(t *testing.T) {
		_, _, err := UserPhotos(form.SearchPhotos{Scope: album}, scopeSession("guest"))
		assert.Equal(t, ErrForbidden, err)
	})
	t.Run("VisitorSharedScopeAllowed", func(t *testing.T) {
		// Unregistered visitor whose session carries a share for the scoped album.
		visitor := &entity.Session{}
		visitor.SetData(&entity.SessionData{Shares: entity.UIDs{album}})
		_, _, err := UserPhotos(form.SearchPhotos{Scope: album}, visitor)
		assert.NoError(t, err)
	})
	t.Run("AdminScopeAllowed", func(t *testing.T) {
		_, _, err := UserPhotos(form.SearchPhotos{Scope: album}, scopeSession("alice"))
		assert.NoError(t, err)
	})
}

func TestFileVisibleToSession(t *testing.T) {
	t.Run("EmptyHash", func(t *testing.T) {
		ok, err := FileVisibleToSession("", scopeSession("guest"))
		assert.NoError(t, err)
		assert.False(t, ok)
	})
	t.Run("NilSession", func(t *testing.T) {
		ok, err := FileVisibleToSession(scopeNormalFileHash, nil)
		assert.NoError(t, err)
		assert.True(t, ok)
	})
	t.Run("AdminPrivateFile", func(t *testing.T) {
		ok, err := FileVisibleToSession(scopePrivateFileHash, scopeSession("alice"))
		assert.NoError(t, err)
		assert.True(t, ok)
	})
	t.Run("GuestDeniedPrivateFile", func(t *testing.T) {
		ok, err := FileVisibleToSession(scopePrivateFileHash, scopeSession("guest"))
		assert.NoError(t, err)
		assert.False(t, ok)
	})
}
