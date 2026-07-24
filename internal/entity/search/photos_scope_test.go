package search

import (
	"testing"

	"github.com/jinzhu/gorm"
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

	// scopeFolderPhotoUID and scopeFolderFileHash belong to the "april-1990" folder (smart) album,
	// whose members derive from a path filter rather than photos_albums rows.
	scopeFolderPhotoUID = "ps6sg6be2lvl0yh0"                         // "Photo03", public, path 1990/04
	scopeFolderFileHash = "pcad9168fa6acc5c5c2965adf6ec465ca42fd818" // primary file of "Photo03"

	// scopeRegularPhotoUID is a non-private member of the regular album shared by scopeRegularShareToken;
	// it is resolved by the personal-scope predicate without the smart-album fallback.
	scopeRegularPhotoUID = "ps6sg6be2lvl0y21" // photos_albums member of "christmas-2030"

	// Share-link tokens resolve to shared albums; a visitor session derives its shares from them.
	//nolint:gosec // G101: deterministic fixture share-link tokens for tests only.
	scopeFolderShareToken = "8jxf3jfn2k" // link to "april-1990" folder album (contains 1990/04)
	//nolint:gosec // G101: deterministic fixture share-link tokens for tests only.
	scopeStateShareToken = "9jxf3jfn2k" // link to "california-usa" state album (excludes 1990/04)
	//nolint:gosec // G101: deterministic fixture share-link tokens for tests only.
	scopeRegularShareToken = "4jxf3jfn2k" // link to "christmas-2030" regular album (no filter)
)

// scopeSession builds an in-memory session for the named user fixture.
func scopeSession(name string) *entity.Session {
	s := &entity.Session{}
	s.SetUser(entity.UserFixtures.Pointer(name))
	return s
}

// scopeVisitorWithShares builds an unregistered visitor session that has redeemed the given share
// link tokens, so its SharedUIDs resolve from the matching links exactly as in production.
func scopeVisitorWithShares(tokens ...string) *entity.Session {
	s := &entity.Session{}
	s.SetData(&entity.SessionData{Tokens: tokens})
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
	t.Run("VisitorSharedFolderAlbum", func(t *testing.T) {
		// A picture shared only through a folder (smart) album has no photos_albums row, so it is
		// visible only via the shared-album fallback.
		ok, err := PhotoVisibleToSession(scopeFolderPhotoUID, scopeVisitorWithShares(scopeFolderShareToken))
		assert.NoError(t, err)
		assert.True(t, ok)
	})
	t.Run("VisitorWrongSmartAlbum", func(t *testing.T) {
		// Sharing a different smart album must not expose the folder picture.
		ok, err := PhotoVisibleToSession(scopeFolderPhotoUID, scopeVisitorWithShares(scopeStateShareToken))
		assert.NoError(t, err)
		assert.False(t, ok)
	})
	t.Run("VisitorNoShares", func(t *testing.T) {
		ok, err := PhotoVisibleToSession(scopeFolderPhotoUID, scopeVisitorWithShares())
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
	t.Run("VisitorSharedFolderAlbumFile", func(t *testing.T) {
		// The file hash resolves to a picture shared only through a folder (smart) album.
		ok, err := FileVisibleToSession(scopeFolderFileHash, scopeVisitorWithShares(scopeFolderShareToken))
		assert.NoError(t, err)
		assert.True(t, ok)
	})
	t.Run("VisitorWrongSmartAlbumFile", func(t *testing.T) {
		ok, err := FileVisibleToSession(scopeFolderFileHash, scopeVisitorWithShares(scopeStateShareToken))
		assert.NoError(t, err)
		assert.False(t, ok)
	})
}

func TestScopePhotosForSessionAllowUIDs(t *testing.T) {
	t.Run("VisitorFolderPhotoDroppedWithoutAllow", func(t *testing.T) {
		// Without the allow-list a folder (smart) album picture has no photos_albums row, so the
		// shared-scope predicate drops it even though the folder link is shared.
		base := UnscopedDb().Table("photos").Where("photos.photo_uid = ?", scopeFolderPhotoUID)
		var count int
		assert.NoError(t, scopePhotosForSession(base, scopeVisitorWithShares(scopeFolderShareToken), nil).Count(&count).Error)
		assert.Equal(t, 0, count)
	})
	t.Run("VisitorFolderPhotoAllowed", func(t *testing.T) {
		base := UnscopedDb().Table("photos").Where("photos.photo_uid = ?", scopeFolderPhotoUID)
		var count int
		allow := []string{scopeFolderPhotoUID}
		assert.NoError(t, scopePhotosForSession(base, scopeVisitorWithShares(scopeFolderShareToken), allow).Count(&count).Error)
		assert.Equal(t, 1, count)
	})
	t.Run("AdminUnchanged", func(t *testing.T) {
		base := UnscopedDb().Table("photos")
		assert.Same(t, base, scopePhotosForSession(base, scopeSession("alice"), []string{scopeFolderPhotoUID}))
	})
}

func TestExcludeRestrictedPhotos(t *testing.T) {
	t.Run("AdminKeepsPrivate", func(t *testing.T) {
		var count int
		err := excludeRestrictedPhotos(UnscopedDb().Table("photos").Where("photos.photo_uid = ?", scopePrivatePhotoUID), scopeSession("alice")).Count(&count).Error
		assert.NoError(t, err)
		assert.Equal(t, 1, count)
	})
	t.Run("GuestExcludesPrivate", func(t *testing.T) {
		var count int
		err := excludeRestrictedPhotos(UnscopedDb().Table("photos").Where("photos.photo_uid = ?", scopePrivatePhotoUID), scopeSession("guest")).Count(&count).Error
		assert.NoError(t, err)
		assert.Equal(t, 0, count)
	})
}

func TestScopeVisibleSelection(t *testing.T) {
	folderFileStmt := func() *gorm.DB {
		return UnscopedDb().Table("files").
			Joins("JOIN photos ON photos.id = files.photo_id").
			Where("files.file_hash = ? AND files.deleted_at IS NULL", scopeFolderFileHash)
	}
	t.Run("AdminUnchanged", func(t *testing.T) {
		base := UnscopedDb().Table("photos")
		assert.Same(t, base, ScopeVisibleSelection(base, scopeSession("alice"), []string{scopeFolderPhotoUID}))
	})
	t.Run("VisitorFolderSelectionDownloadable", func(t *testing.T) {
		// A file whose picture is shared only through a folder (smart) album must be downloadable when
		// its UID is part of the selection, matching FileVisibleToSession.
		var count int
		err := ScopeVisibleSelection(folderFileStmt(), scopeVisitorWithShares(scopeFolderShareToken), []string{scopeFolderPhotoUID}).Count(&count).Error
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, count, 1)
	})
	t.Run("VisitorWrongSmartAlbumNotDownloadable", func(t *testing.T) {
		var count int
		err := ScopeVisibleSelection(folderFileStmt(), scopeVisitorWithShares(scopeStateShareToken), []string{scopeFolderPhotoUID}).Count(&count).Error
		assert.NoError(t, err)
		assert.Equal(t, 0, count)
	})
	t.Run("VisitorNoSharesNotDownloadable", func(t *testing.T) {
		var count int
		err := ScopeVisibleSelection(folderFileStmt(), scopeVisitorWithShares(), []string{scopeFolderPhotoUID}).Count(&count).Error
		assert.NoError(t, err)
		assert.Equal(t, 0, count)
	})
}

func TestSharedSmartAlbumPhotoUIDs(t *testing.T) {
	t.Run("EmptySelection", func(t *testing.T) {
		assert.Nil(t, sharedSmartAlbumPhotoUIDs(nil, scopeVisitorWithShares(scopeFolderShareToken)))
	})
	t.Run("NilSession", func(t *testing.T) {
		assert.Nil(t, sharedSmartAlbumPhotoUIDs([]string{scopeFolderPhotoUID}, nil))
	})
	t.Run("NoShares", func(t *testing.T) {
		assert.Empty(t, sharedSmartAlbumPhotoUIDs([]string{scopeFolderPhotoUID}, scopeVisitorWithShares()))
	})
	t.Run("FolderShareMatches", func(t *testing.T) {
		uids := sharedSmartAlbumPhotoUIDs([]string{scopeFolderPhotoUID}, scopeVisitorWithShares(scopeFolderShareToken))
		assert.Contains(t, uids, scopeFolderPhotoUID)
	})
	t.Run("WrongSmartAlbum", func(t *testing.T) {
		assert.Empty(t, sharedSmartAlbumPhotoUIDs([]string{scopeFolderPhotoUID}, scopeVisitorWithShares(scopeStateShareToken)))
	})
	t.Run("RegularAlbumSkipped", func(t *testing.T) {
		// A shared regular album (empty filter) is skipped because ScopePhotosForSession already covers
		// its photos_albums membership.
		assert.Empty(t, sharedSmartAlbumPhotoUIDs([]string{scopeFolderPhotoUID}, scopeVisitorWithShares(scopeRegularShareToken)))
	})
	t.Run("FiltersUnselectedUIDs", func(t *testing.T) {
		// Only the selected UID that belongs to the shared folder is returned; a non-member UID is not.
		uids := sharedSmartAlbumPhotoUIDs([]string{scopeFolderPhotoUID, scopeNormalPhotoUID}, scopeVisitorWithShares(scopeFolderShareToken))
		assert.Contains(t, uids, scopeFolderPhotoUID)
		assert.NotContains(t, uids, scopeNormalPhotoUID)
	})
	t.Run("MultiFileMemberReturned", func(t *testing.T) {
		// Every selected member of the shared folder must be returned, including scopeFolderPhotoUID
		// ("Photo03"), which has several files (JPEG + extra image + video). The membership search must
		// count one row per photo, not per file, so a multi-file picture cannot crowd others out.
		member17 := entity.PhotoFixtures.Pointer("Photo17").PhotoUID
		member20 := entity.PhotoFixtures.Pointer("Photo20").PhotoUID
		selection := []string{scopeFolderPhotoUID, member17, member20}
		uids := sharedSmartAlbumPhotoUIDs(selection, scopeVisitorWithShares(scopeFolderShareToken))
		assert.Contains(t, uids, scopeFolderPhotoUID)
		assert.Contains(t, uids, member17)
		assert.Contains(t, uids, member20)
		assert.Len(t, uids, len(selection))
	})
	t.Run("PrimaryYieldsOneRowPerMultiFilePhoto", func(t *testing.T) {
		// Locks the reason sharedSmartAlbumPhotoUIDs sets Primary: the album-scoped membership search
		// returns one row per file, so scopeFolderPhotoUID ("Photo03", which has three files) yields
		// several rows. A Count sized to the number of selected photos would then let one multi-file
		// picture consume the limit and drop the others; Primary collapses each photo to a single row so
		// Count bounds photos rather than files.
		sess := scopeVisitorWithShares(scopeFolderShareToken)
		const folderAlbumUID = "as6sg6bipogaaba1" // "april-1990" folder album shared by scopeFolderShareToken
		perFile, _, err := UserPhotos(form.SearchPhotos{Scope: folderAlbumUID, UID: scopeFolderPhotoUID, Count: 100}, sess)
		assert.NoError(t, err)
		perPhoto, _, err := UserPhotos(form.SearchPhotos{Scope: folderAlbumUID, UID: scopeFolderPhotoUID, Count: 100, Primary: true}, sess)
		assert.NoError(t, err)
		assert.Greater(t, len(perFile), 1) // several files → several rows without Primary
		assert.Len(t, perPhoto, 1)         // exactly one row per photo with Primary
	})
}

func TestSharedSmartAlbumContains(t *testing.T) {
	t.Run("EmptyID", func(t *testing.T) {
		ok, err := sharedSmartAlbumContains("", scopeVisitorWithShares(scopeFolderShareToken))
		assert.NoError(t, err)
		assert.False(t, ok)
	})
	t.Run("NilSession", func(t *testing.T) {
		ok, err := sharedSmartAlbumContains(scopeFolderPhotoUID, nil)
		assert.NoError(t, err)
		assert.False(t, ok)
	})
	t.Run("NoShares", func(t *testing.T) {
		ok, err := sharedSmartAlbumContains(scopeFolderPhotoUID, scopeVisitorWithShares())
		assert.NoError(t, err)
		assert.False(t, ok)
	})
	t.Run("PhotoUIDInSharedFolder", func(t *testing.T) {
		ok, err := sharedSmartAlbumContains(scopeFolderPhotoUID, scopeVisitorWithShares(scopeFolderShareToken))
		assert.NoError(t, err)
		assert.True(t, ok)
	})
	t.Run("WrongSmartAlbum", func(t *testing.T) {
		ok, err := sharedSmartAlbumContains(scopeFolderPhotoUID, scopeVisitorWithShares(scopeStateShareToken))
		assert.NoError(t, err)
		assert.False(t, ok)
	})
	t.Run("RegularAlbumSkipped", func(t *testing.T) {
		// A shared regular album (empty filter) is skipped here because ScopePhotosForSession already
		// covers its photos_albums membership, so the fallback reports no match.
		ok, err := sharedSmartAlbumContains(scopeFolderPhotoUID, scopeVisitorWithShares(scopeRegularShareToken))
		assert.NoError(t, err)
		assert.False(t, ok)
	})
}

// BenchmarkPhotoVisibleToSession measures the per-call cost of the single-item visibility check on
// the paths a restricted session hits when the lightbox prefetches photo metadata: the privileged
// short-circuit, the personal-scope hit (regular share), and the smart-album fallback (folder share)
// including the worst case where the fallback searches a shared album and finds no match.
func BenchmarkPhotoVisibleToSession(b *testing.B) {
	admin := scopeSession("alice")
	folderVisitor := scopeVisitorWithShares(scopeFolderShareToken)
	regularVisitor := scopeVisitorWithShares(scopeRegularShareToken)
	stateVisitor := scopeVisitorWithShares(scopeStateShareToken)

	b.Run("AdminShortCircuit", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = PhotoVisibleToSession(scopeFolderPhotoUID, admin)
		}
	})
	b.Run("RegularSharePersonalScopeHit", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = PhotoVisibleToSession(scopeRegularPhotoUID, regularVisitor)
		}
	})
	b.Run("FolderShareSmartAlbumFallback", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = PhotoVisibleToSession(scopeFolderPhotoUID, folderVisitor)
		}
	})
	b.Run("OutOfScopeFallbackMiss", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = PhotoVisibleToSession(scopeFolderPhotoUID, stateVisitor)
		}
	})
}

func TestFileVisibleToPublic(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		v, err := FileVisibleToPublic("")
		assert.NoError(t, err)
		assert.False(t, v)
	})
}

func TestPhotoVisibleToPublic(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		v, err := PhotoVisibleToPublic("")
		assert.NoError(t, err)
		assert.False(t, v)
	})
}

func TestPhotoSessionSeesPrivate(t *testing.T) {
	t.Run("NilDenied", func(t *testing.T) {
		assert.False(t, PhotoSessionSeesPrivate(nil))
	})
	t.Run("VisitorDenied", func(t *testing.T) {
		assert.False(t, PhotoSessionSeesPrivate(entity.SessionFixtures.Pointer("visitor")))
	})
	t.Run("AdminAllowed", func(t *testing.T) {
		assert.True(t, PhotoSessionSeesPrivate(entity.SessionFixtures.Pointer("alice")))
	})
}
