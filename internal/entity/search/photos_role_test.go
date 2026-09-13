package search

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/authn"
)

// Fixtures shared by the effective-role cases below.
const (
	// roleTestAlbum is a manual album that alice neither created nor shares.
	roleTestAlbum = "as6sg6bxpogaaba8"
	// roleTestSharedAlbum is a manual album the admin user alice has a share for.
	roleTestSharedAlbum = "as6sg6bxpogaaba9"
	// roleTestPrivatePhoto is a private picture outside both albums.
	roleTestPrivatePhoto = "pqkm36fjqvset9uz"
	// roleTestPrivateGeoPhoto is a private picture that carries coordinates.
	roleTestPrivateGeoPhoto = "ps6sg6be2lvl0y12"
	// roleTestPublicPhoto is a public picture outside both albums.
	roleTestPublicPhoto = "ps6sg6be2lvl0yh0"
)

// clientSessionFor builds a client session with the given client role, acting for the named user.
// The user's shares are loaded, since HasShare reads them from the account for a registered user.
func clientSessionFor(role acl.Role, userName string) *entity.Session {
	s := &entity.Session{}
	s.SetClient(&entity.Client{ClientRole: role.String(), AuthProvider: authn.ProviderClient.String()})
	s.SetUser(entity.UserFixtures.Pointer(userName))
	s.GetUser().RefreshShares()
	return s
}

func TestUserPhotos_EffectiveRole(t *testing.T) {
	t.Run("MixedPrincipalUnsharedScopeForbidden", func(t *testing.T) {
		// An instance client acting for an admin is limited by the client role, so it may scope only
		// to an album it owns or has a share for.
		_, _, err := UserPhotos(form.SearchPhotos{Scope: roleTestAlbum}, clientSessionFor(acl.RoleInstance, "alice"))
		assert.Equal(t, ErrForbidden, err)
	})
	t.Run("AdminUnsharedScopeAllowed", func(t *testing.T) {
		_, _, err := UserPhotos(form.SearchPhotos{Scope: roleTestAlbum}, scopeSession("alice"))
		assert.NoError(t, err)
	})
	t.Run("MixedPrincipalSharedScopeAllowed", func(t *testing.T) {
		// A share the owner holds still admits the scope, so the narrower principal is not locked out
		// of what it is entitled to.
		sess := clientSessionFor(acl.RoleInstance, "alice")
		results, _, err := UserPhotos(form.SearchPhotos{Scope: roleTestSharedAlbum, Count: 1000}, sess)
		assert.NoError(t, err)
		assert.NotEmpty(t, results)
	})
	t.Run("MixedPrincipalExcludesPrivate", func(t *testing.T) {
		// The client role carries no access_private, so a request for private pictures is answered as
		// a public one: the results are non-empty and contain no private picture.
		sess := clientSessionFor(acl.RoleInstance, "alice")
		assert.False(t, PhotoSessionSeesPrivate(sess))
		results, _, err := UserPhotos(form.SearchPhotos{Private: true, Count: 1000}, sess)
		assert.NoError(t, err)
		assert.NotEmpty(t, results)
		assert.NotContains(t, photoUIDs(results), roleTestPrivatePhoto)
	})
	t.Run("AdminIncludesPrivate", func(t *testing.T) {
		assert.True(t, PhotoSessionSeesPrivate(scopeSession("alice")))
		results, _, err := UserPhotos(form.SearchPhotos{Private: true, Count: 1000}, scopeSession("alice"))
		assert.NoError(t, err)
		assert.Contains(t, photoUIDs(results), roleTestPrivatePhoto)
	})
	t.Run("MixedPrincipalSharedScopeExcludesPrivateMember", func(t *testing.T) {
		// The share admits the scope, and the client role still withholds the private member inside
		// it, so the scope decides which album is reachable and the role decides what it shows.
		member := entity.NewPhotoAlbum(roleTestPrivatePhoto, roleTestSharedAlbum)
		require.NoError(t, member.Save())
		entity.FlushAlbumCache()

		t.Cleanup(func() {
			_ = entity.UnscopedDb().Delete(member).Error
			entity.FlushAlbumCache()
		})

		sess := clientSessionFor(acl.RoleInstance, "alice")
		results, _, err := UserPhotos(form.SearchPhotos{Scope: roleTestSharedAlbum, Private: true, Count: 1000}, sess)
		require.NoError(t, err)
		require.NotEmpty(t, results, "the scope must still return the album's other members")
		assert.NotContains(t, photoUIDs(results), roleTestPrivatePhoto)

		admin, _, err := UserPhotos(form.SearchPhotos{Scope: roleTestSharedAlbum, Private: true, Count: 1000}, scopeSession("alice"))
		require.NoError(t, err)
		assert.Contains(t, photoUIDs(admin), roleTestPrivatePhoto, "the fixture must place the private picture in the shared album")
	})
	t.Run("BroadClientNarrowUserExcludesPrivate", func(t *testing.T) {
		// The mirror principal, and the shape a credential takes by default: a full-access client role
		// owned by a narrow account. The account is what limits it, so a request for private pictures
		// is answered as a public one. The guest fixture holds no share, so the listing is empty here
		// and the predicate is what carries the case.
		sess := clientSessionFor(acl.RoleClient, "guest")
		assert.True(t, acl.Rules.Allow(acl.ResourcePhotos, acl.RoleClient, acl.AccessPrivate))
		assert.False(t, PhotoSessionSeesPrivate(sess))
		assert.True(t, sess.HasSharedAccessOnly(acl.ResourcePhotos))

		results, _, err := UserPhotos(form.SearchPhotos{Private: true, Count: 1000}, sess)
		require.NoError(t, err)
		assert.NotContains(t, photoUIDs(results), roleTestPrivatePhoto)
	})
	t.Run("MixedPrincipalExcludesArchived", func(t *testing.T) {
		// The client role carries no delete grant, so a request for archived pictures is answered as
		// a request for current ones.
		sess := clientSessionFor(acl.RoleInstance, "alice")
		results, _, err := UserPhotos(form.SearchPhotos{Archived: true, Count: 1000}, sess)
		assert.NoError(t, err)
		assert.NotEmpty(t, results)

		archived, _, err := UserPhotos(form.SearchPhotos{Archived: true, Count: 1000}, scopeSession("alice"))
		assert.NoError(t, err)
		assert.NotSubset(t, photoUIDs(results), photoUIDs(archived))
	})
}

func TestUserAlbums_EffectiveRole(t *testing.T) {
	t.Run("MixedPrincipalLimitedToShared", func(t *testing.T) {
		// The instance client role limits the session to shared, owned and published albums.
		sess := clientSessionFor(acl.RoleInstance, "alice")
		results, err := UserAlbums(form.SearchAlbums{Type: entity.AlbumManual, Count: 1000}, sess)
		assert.NoError(t, err)
		assert.NotContains(t, albumUIDs(results), roleTestAlbum)
	})
	t.Run("AdminListsAll", func(t *testing.T) {
		results, err := UserAlbums(form.SearchAlbums{Type: entity.AlbumManual, Count: 1000}, scopeSession("alice"))
		assert.NoError(t, err)
		assert.Contains(t, albumUIDs(results), roleTestAlbum)
	})
	t.Run("TypelessListingNeedsDefaultResourceAccess", func(t *testing.T) {
		// A request naming neither a type nor a UID is answered against acl.ResourceDefault, which every
		// edition grants to admin alone. A client session therefore has to name what it is listing.
		sess := clientSessionFor(acl.RoleClient, "alice")
		_, err := UserAlbums(form.SearchAlbums{Count: 1000}, sess)
		assert.Equal(t, ErrForbidden, err)

		_, err = UserAlbums(form.SearchAlbums{Type: entity.AlbumManual, Count: 1000}, sess)
		assert.NoError(t, err)

		_, err = UserAlbums(form.SearchAlbums{UID: roleTestAlbum, Count: 1000}, sess)
		assert.NoError(t, err)
	})
}

func TestUserPhotosGeo_EffectiveRole(t *testing.T) {
	t.Run("MixedPrincipalSearchDenied", func(t *testing.T) {
		// An instance client holds no places search grant, so the intersection denies the map view.
		_, err := UserPhotosGeo(form.SearchPhotosGeo{}, clientSessionFor(acl.RoleInstance, "alice"))
		assert.Equal(t, ErrForbidden, err)
	})
	t.Run("AdminSearchAllowed", func(t *testing.T) {
		_, err := UserPhotosGeo(form.SearchPhotosGeo{}, scopeSession("alice"))
		assert.NoError(t, err)
	})
}

func TestUserPhotos_NearVisibility(t *testing.T) {
	t.Run("AbsentReferenceNotFound", func(t *testing.T) {
		_, _, err := UserPhotos(form.SearchPhotos{Near: "ps6sg6be2lvl0000"}, scopeSession("guest"))
		assert.Equal(t, ErrNotFound, err)
	})
	t.Run("InvisibleReferenceTakesTheSamePath", func(t *testing.T) {
		_, _, err := UserPhotos(form.SearchPhotos{Near: roleTestPrivatePhoto}, scopeSession("guest"))
		assert.Equal(t, ErrNotFound, err)
	})
	t.Run("VisibleReferenceAccepted", func(t *testing.T) {
		_, _, err := UserPhotos(form.SearchPhotos{Near: roleTestPublicPhoto}, scopeSession("alice"))
		assert.NoError(t, err)
	})
	t.Run("MixedPrincipalAbsentReferenceNotFound", func(t *testing.T) {
		// A UID with no row behind it cannot be found whatever the visibility check answers, so this
		// states the contract rather than detecting a change in it.
		_, _, err := UserPhotos(form.SearchPhotos{Near: "ps6sg6be2lvl0000"}, clientSessionFor(acl.RoleInstance, "alice"))
		assert.Equal(t, ErrNotFound, err)
	})
	t.Run("MixedPrincipalInvisibleReferenceTakesTheSamePath", func(t *testing.T) {
		// The owner is an admin and sees this reference, so the refusal comes from the client role.
		_, _, err := UserPhotos(form.SearchPhotos{Near: roleTestPrivatePhoto}, clientSessionFor(acl.RoleInstance, "alice"))
		assert.Equal(t, ErrNotFound, err)

		_, _, err = UserPhotos(form.SearchPhotos{Near: roleTestPrivatePhoto}, scopeSession("alice"))
		assert.NoError(t, err)
	})
	t.Run("MixedPrincipalVisibleReferenceAccepted", func(t *testing.T) {
		_, _, err := UserPhotos(form.SearchPhotos{Near: roleTestPublicPhoto}, clientSessionFor(acl.RoleInstance, "alice"))
		assert.NoError(t, err)
	})
	t.Run("NoSessionUnrestricted", func(t *testing.T) {
		_, _, err := UserPhotos(form.SearchPhotos{Near: roleTestPrivatePhoto}, nil)
		assert.NoError(t, err)
	})
}

func TestUserPhotosGeo_NearVisibility(t *testing.T) {
	t.Run("AbsentReferenceNotFound", func(t *testing.T) {
		_, err := UserPhotosGeo(form.SearchPhotosGeo{Near: "ps6sg6be2lvl0000"}, scopeSession("guest"))
		assert.Equal(t, ErrNotFound, err)
	})
	t.Run("InvisibleReferenceTakesTheSamePath", func(t *testing.T) {
		_, err := UserPhotosGeo(form.SearchPhotosGeo{Near: roleTestPrivatePhoto}, scopeSession("guest"))
		assert.Equal(t, ErrNotFound, err)
	})
	t.Run("VisibleReferenceAccepted", func(t *testing.T) {
		_, err := UserPhotosGeo(form.SearchPhotosGeo{Near: roleTestPublicPhoto}, scopeSession("alice"))
		assert.NoError(t, err)
	})
	t.Run("MixedPrincipalAbsentReference", func(t *testing.T) {
		// The map view refuses an instance client outright, so these three cases also pin that the
		// reference is resolved before the search gate runs.
		_, err := UserPhotosGeo(form.SearchPhotosGeo{Near: "ps6sg6be2lvl0000"}, clientSessionFor(acl.RoleInstance, "alice"))
		assert.Equal(t, ErrNotFound, err)
	})
	t.Run("MixedPrincipalInvisibleReference", func(t *testing.T) {
		_, err := UserPhotosGeo(form.SearchPhotosGeo{Near: roleTestPrivateGeoPhoto}, clientSessionFor(acl.RoleInstance, "alice"))
		assert.Equal(t, ErrNotFound, err)
	})
	t.Run("MixedPrincipalVisibleReferenceReachesTheSearchGate", func(t *testing.T) {
		_, err := UserPhotosGeo(form.SearchPhotosGeo{Near: roleTestPublicPhoto}, clientSessionFor(acl.RoleInstance, "alice"))
		assert.Equal(t, ErrForbidden, err)
	})
	t.Run("NoSessionUnrestricted", func(t *testing.T) {
		_, err := UserPhotosGeo(form.SearchPhotosGeo{Near: roleTestPrivatePhoto}, nil)
		assert.NoError(t, err)
	})
}

// photoUIDs returns the photo UIDs of the given results.
func photoUIDs(results PhotoResults) []string {
	uids := make([]string, 0, len(results))

	for i := range results {
		uids = append(uids, results[i].PhotoUID)
	}

	return uids
}

// albumUIDs returns the album UIDs of the given results.
func albumUIDs(results AlbumResults) []string {
	uids := make([]string, 0, len(results))

	for i := range results {
		uids = append(uids, results[i].AlbumUID)
	}

	return uids
}

func TestUserPhotos_ScopeRequiresLibraryReach(t *testing.T) {
	// A registered account whose role grants neither shared nor library access may not scope to an
	// album it does not own; the gate asks for whole-library reach, not for shared-only status.
	t.Run("RoleNoneForbidden", func(t *testing.T) {
		sess := scopeSession("unauthorized")
		assert.True(t, sess.DeniesAll(acl.ResourcePhotos, acl.Permissions{acl.AccessAll, acl.AccessLibrary}))
		_, _, err := UserPhotos(form.SearchPhotos{Scope: roleTestAlbum}, sess)
		assert.Equal(t, ErrForbidden, err)
	})
}

func TestUserPhotosGeo_ExcludesRestrictedPictures(t *testing.T) {
	// The map view is the only surface that returns coordinates, so its private and archived
	// exclusions are pinned separately from the photo search.
	t.Run("GuestExcludesPrivate", func(t *testing.T) {
		sess := scopeSession("guest")
		results, err := UserPhotosGeo(form.SearchPhotosGeo{Private: true, Count: 1000}, sess)
		assert.NoError(t, err)
		assert.NotContains(t, geoUIDs(results), roleTestPrivateGeoPhoto)
	})
	t.Run("AdminIncludesPrivate", func(t *testing.T) {
		results, err := UserPhotosGeo(form.SearchPhotosGeo{Private: true, Count: 1000}, scopeSession("alice"))
		assert.NoError(t, err)
		assert.Contains(t, geoUIDs(results), roleTestPrivateGeoPhoto)
	})
	t.Run("ArchivedIsFilteredByTheBaseQuery", func(t *testing.T) {
		// UserPhotosGeo selects "photos.deleted_at IS NULL" outright, so the map never returns an
		// archived picture and the per-session archived flag cannot widen it.
		results, err := UserPhotosGeo(form.SearchPhotosGeo{Archived: true, Count: 1000}, scopeSession("alice"))
		assert.NoError(t, err)
		assert.Empty(t, results)
	})
	t.Run("GuestLimitedToSharedRows", func(t *testing.T) {
		guest, err := UserPhotosGeo(form.SearchPhotosGeo{Count: 1000}, scopeSession("guest"))
		assert.NoError(t, err)
		admin, err := UserPhotosGeo(form.SearchPhotosGeo{Count: 1000}, scopeSession("alice"))
		assert.NoError(t, err)
		assert.Less(t, len(guest), len(admin))
	})
}

func TestSessionAuditRole(t *testing.T) {
	t.Run("ClientWithUser", func(t *testing.T) {
		assert.Equal(t, "instance/admin", sessionAuditRole(clientSessionFor(acl.RoleInstance, "alice")))
	})
	t.Run("ClientWithoutUser", func(t *testing.T) {
		s := &entity.Session{}
		s.SetClient(&entity.Client{ClientRole: acl.RoleInstance.String(), AuthProvider: authn.ProviderClient.String()})
		assert.Equal(t, "instance", sessionAuditRole(s))
	})
	t.Run("User", func(t *testing.T) {
		assert.Equal(t, "admin", sessionAuditRole(scopeSession("alice")))
	})
}

// geoUIDs returns the photo UIDs of the given map results.
func geoUIDs(results GeoResults) []string {
	uids := make([]string, 0, len(results))

	for i := range results {
		uids = append(uids, results[i].PhotoUID)
	}

	return uids
}
