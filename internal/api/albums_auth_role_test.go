package api

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/authn"
)

// mixedPrincipalSession builds a restricted client session acting for a privileged account: an
// instance client (GrantSearchShared) owned by the admin user alice.
func mixedPrincipalSession() *entity.Session {
	s := &entity.Session{}
	s.SetClient(&entity.Client{ClientRole: acl.RoleInstance.String(), AuthProvider: authn.ProviderClient.String()})
	s.SetUser(entity.UserFixtures.Pointer("alice"))
	s.GetUser().RefreshShares()
	return s
}

// sharedAlbumUID is the album the admin user alice holds a share for.
const sharedAlbumUID = "as6sg6bxpogaaba9"

func TestAlbumViewableBySessionEffectiveRole(t *testing.T) {
	// An album the owning account neither created nor shares.
	album := entity.AlbumFixtures.Get("holiday-2030")

	t.Run("AdminMayView", func(t *testing.T) {
		s := &entity.Session{}
		s.SetUser(entity.UserFixtures.Pointer("alice"))
		assert.True(t, albumViewableBySession(s, album))
	})
	t.Run("MixedPrincipalMayNot", func(t *testing.T) {
		s := mixedPrincipalSession()
		assert.False(t, s.GetUser().HasSharedAccessOnly(acl.ResourceAlbums))
		assert.True(t, s.HasSharedAccessOnly(acl.ResourceAlbums))
		assert.False(t, albumViewableBySession(s, album))
	})
	t.Run("MixedPrincipalMayViewASharedAlbum", func(t *testing.T) {
		// The share the owning account holds still admits the album, so the narrower principal keeps
		// what it is entitled to and the gate is not simply refusing everything.
		assert.True(t, albumViewableBySession(mixedPrincipalSession(), entity.Album{AlbumUID: sharedAlbumUID}))
	})
	t.Run("MixedPrincipalMayViewItsOwnAlbum", func(t *testing.T) {
		s := mixedPrincipalSession()
		own := entity.Album{AlbumUID: album.AlbumUID, CreatedBy: s.UserUID}
		assert.True(t, albumViewableBySession(s, own))
	})
}

func TestAlbumShareRequiredEffectiveRole(t *testing.T) {
	t.Run("AdminNeedsNoShare", func(t *testing.T) {
		s := &entity.Session{}
		s.SetUser(entity.UserFixtures.Pointer("alice"))
		assert.False(t, albumShareRequired(s, "as6sg6bxpogaaba8"))
	})
	t.Run("MixedPrincipalNeedsShare", func(t *testing.T) {
		assert.True(t, albumShareRequired(mixedPrincipalSession(), "as6sg6bxpogaaba8"))
	})
	t.Run("MixedPrincipalHoldsTheShare", func(t *testing.T) {
		// Unlike the view gate, this one has no self-owned exception, so the share is the only thing
		// that admits it.
		assert.False(t, albumShareRequired(mixedPrincipalSession(), sharedAlbumUID))
	})
}

// TestGetAlbum_EffectiveRole covers the album detail route, so the wiring between the handler's own
// authorization and the view gate is pinned as well as the gate itself.
func TestGetAlbum_EffectiveRole(t *testing.T) {
	app, router, conf := NewApiTest()
	conf.SetAuthMode(config.AuthModePasswd)
	defer conf.SetAuthMode(config.AuthModePublic)
	GetAlbum(router)

	unshared := entity.AlbumFixtures.Get("holiday-2030").AlbumUID

	t.Run("MixedPrincipalRefused", func(t *testing.T) {
		// The client role admits the request and the view gate refuses the album, so the route
		// answers not-found for an album the owning administrator could open.
		token := mixedPrincipalToken(t, acl.RoleInstance, "alice")
		r := AuthenticatedRequest(app, http.MethodGet, "/api/v1/albums/"+unshared, token)

		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("MixedPrincipalSharedAlbum", func(t *testing.T) {
		token := mixedPrincipalToken(t, acl.RoleInstance, "alice")
		r := AuthenticatedRequest(app, http.MethodGet, "/api/v1/albums/"+sharedAlbumUID, token)

		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("AdminAllowed", func(t *testing.T) {
		token := mixedPrincipalToken(t, acl.RoleClient, "alice")
		r := AuthenticatedRequest(app, http.MethodGet, "/api/v1/albums/"+unshared, token)

		assert.Equal(t, http.StatusOK, r.Code)
	})
}

// TestSearchQuality_EffectiveRole covers the review-queue gate both search handlers apply. A picture
// below the quality threshold is surfaced only to a principal that may manage pictures, and a client
// session is evaluated on the intersection of both its roles.
func TestSearchQuality_EffectiveRole(t *testing.T) {
	app, router, conf := NewApiTest()
	conf.SetAuthMode(config.AuthModePasswd)
	defer conf.SetAuthMode(config.AuthModePublic)

	settings := conf.Settings()
	review := settings.Features.Review
	settings.Features.Review = true
	defer func() { settings.Features.Review = review }()

	SearchPhotos(router)
	SearchGeo(router)

	found := func(t *testing.T, route string, role acl.Role) []string {
		t.Helper()

		token := mixedPrincipalToken(t, role, "alice")
		r := AuthenticatedRequest(app, http.MethodGet, route+"?count=1000", token)
		require.Equal(t, http.StatusOK, r.Code)

		var results []struct {
			UID string `json:"UID"`
		}

		require.NoError(t, json.Unmarshal(r.Body.Bytes(), &results))

		out := make([]string, 0, len(results))

		for _, res := range results {
			out = append(out, res.UID)
		}

		return out
	}

	// The target is taken from what the narrow principal already sees, so lowering its quality is the
	// only thing that can remove it from the second listing.
	visible := found(t, "/api/v1/photos", acl.RoleInstance)
	require.NotEmpty(t, visible)
	target := visible[0]

	require.NoError(t, entity.UnscopedDb().Model(entity.Photo{}).
		Where("photo_uid = ?", target).Update("photo_quality", 1).Error)

	t.Cleanup(func() {
		_ = entity.UnscopedDb().Model(entity.Photo{}).
			Where("photo_uid = ?", target).Update("photo_quality", 3).Error
	})

	t.Run("PhotoSearchWithholdsFromAMixedPrincipal", func(t *testing.T) {
		// The owning account may manage pictures and the client role may not, so the queue stays
		// closed while the rest of what the role may see is still returned.
		after := found(t, "/api/v1/photos", acl.RoleInstance)

		assert.NotContains(t, after, target)
		assert.NotEmpty(t, after, "the search must still return what the role may see")
	})
	t.Run("PhotoSearchShowsAManagingPrincipal", func(t *testing.T) {
		assert.Contains(t, found(t, "/api/v1/photos", acl.RoleClient), target)
	})
	t.Run("MapSearchAdmitsOnTheSameIntersection", func(t *testing.T) {
		// The map view carries its own copy of the gate, and this case pins the admission split only.
		// Pro has the principal that reaches the gate itself, so the content case lives there.
		refused := AuthenticatedRequest(app, http.MethodGet, "/api/v1/geo?count=1000",
			mixedPrincipalToken(t, acl.RoleInstance, "alice"))
		assert.Equal(t, http.StatusForbidden, refused.Code)

		admitted := AuthenticatedRequest(app, http.MethodGet, "/api/v1/geo?count=1000",
			mixedPrincipalToken(t, acl.RoleClient, "alice"))
		assert.Equal(t, http.StatusOK, admitted.Code)
	})
}
