package search

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
)

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
		assert.True(t, photoSessionSeesEverything(nil))
	})
	t.Run("Admin", func(t *testing.T) {
		assert.True(t, photoSessionSeesEverything(scopeSession("alice")))
	})
	t.Run("Guest", func(t *testing.T) {
		assert.False(t, photoSessionSeesEverything(scopeSession("guest")))
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
