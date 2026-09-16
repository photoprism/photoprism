package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/time/unix"
)

// clientCredentialToken returns the auth token of a clientCredentialSession.
func clientCredentialToken(t *testing.T, conf *config.Config, role, scope string, user *entity.User) string {
	t.Helper()

	return clientCredentialSession(t, conf, role, scope, user).AuthToken()
}

// clientCredentialSession registers an OAuth client with the given role and scope and returns a
// client-credentials session for it. The account is optional: without one the session carries no
// user at all, which is what a machine client authenticates as.
func clientCredentialSession(t *testing.T, conf *config.Config, role, scope string, user *entity.User) *entity.Session {
	t.Helper()

	client := &entity.Client{
		ClientUID:    rnd.GenerateUID(entity.ClientUID),
		ClientName:   "payload-probe",
		ClientRole:   role,
		ClientType:   authn.ClientConfidential,
		AuthProvider: authn.ProviderClient.String(),
		AuthMethod:   authn.MethodOAuth2.String(),
		AuthScope:    scope,
		AuthExpires:  unix.Hour,
		AuthTokens:   1,
		AuthEnabled:  true,
	}

	if user != nil {
		client.UserUID = user.UserUID
		client.UserName = user.UserName
	}

	require.NoError(t, client.Create())
	t.Cleanup(func() { entity.UnscopedDb().Unscoped().Delete(client) })

	sess := entity.NewSession(conf.SessionMaxAge(), 0)
	sess.SetClient(client)
	sess.SetGrantType(authn.GrantClientCredentials)

	require.NoError(t, sess.Create())
	t.Cleanup(func() { entity.UnscopedDb().Unscoped().Delete(sess) })

	require.True(t, sess.IsClient())
	require.Equal(t, user != nil, sess.IsRegistered(), "the account decides what this session is")

	return sess
}

// markerCount returns the number of markers a picture response carries across all of its files.
func markerCount(res gjson.Result) int {
	n := 0

	for _, f := range res.Get("Files").Array() {
		n += len(f.Get("Markers").Array())
	}

	return n
}

// TestClientCredential_PictureResponses pins that a credential the role and scope admit on the
// library receives the picture an account receives, measured against an admin control so an empty
// fixture cannot pass for a full payload.
func TestClientCredential_PictureResponses(t *testing.T) {
	const (
		photoUID = "ps6sg6be2lvl0yh7" // names two people and carries markers
		labelUID = "ps6sg6be2lvl0y14" // Photo07, mutated by the round trip below
		label    = "Client Payload Probe"
	)

	app, router, conf := NewApiTest()
	GetPhoto(router)
	UpdatePhoto(router)
	AddPhotoLabel(router)

	prevAuthMode := conf.AuthMode()
	conf.SetAuthMode(config.AuthModePasswd)
	t.Cleanup(func() { conf.SetAuthMode(prevAuthMode) })

	token := func(t *testing.T, scope string) string {
		return clientCredentialToken(t, conf, acl.RoleClient.String(), scope, nil)
	}

	picture := func(t *testing.T, uid, token string) gjson.Result {
		r := AuthenticatedRequest(app, http.MethodGet, "/api/v1/photos/"+uid, token)
		require.Equal(t, http.StatusOK, r.Code)

		return gjson.Parse(r.Body.String())
	}

	admin := picture(t, photoUID, AuthenticateAdmin(app, router))

	require.NotZero(t, markerCount(admin), "the fixture has to carry markers for the case to mean anything")
	require.NotEmpty(t, admin.Get("Details").Raw, "and details, which the reduced payload drops")
	require.NotEmpty(t, admin.Get("Albums").Array(), "and album membership, which is answered separately")

	t.Run("UnrestrictedScope", func(t *testing.T) {
		client := picture(t, photoUID, token(t, "*"))
		assert.Equal(t, markerCount(admin), markerCount(client))
		assert.Equal(t, len(admin.Get("Labels").Array()), len(client.Get("Labels").Array()))
		assert.Equal(t, admin.Get("Details.Keywords").String(), client.Get("Details.Keywords").String())
		assert.Equal(t, admin.Get("CameraSerial").String(), client.Get("CameraSerial").String())
		assert.Equal(t, admin.Get("UUID").String(), client.Get("UUID").String())
		assert.Equal(t, admin.Get("CreatedBy").String(), client.Get("CreatedBy").String())
	})
	t.Run("ScopeNamingPictures", func(t *testing.T) {
		client := picture(t, photoUID, token(t, "photos"))
		assert.Equal(t, markerCount(admin), markerCount(client))
		assert.Equal(t, admin.Get("Details.Keywords").String(), client.Get("Details.Keywords").String())
		assert.Empty(t, client.Get("Albums").Array(),
			"an album is a record of its own, and this credential is not admitted on albums")
	})
	t.Run("ReadScope", func(t *testing.T) {
		client := picture(t, photoUID, token(t, "read"))
		assert.Equal(t, admin.Get("Details.Keywords").String(), client.Get("Details.Keywords").String())
		assert.Equal(t, markerCount(admin), markerCount(client))

		r := AuthenticatedRequestWithBody(app, http.MethodPut, "/api/v1/photos/"+photoUID,
			`{"Favorite": true}`, token(t, "read"))
		assert.Equal(t, http.StatusForbidden, r.Code, "and it may not write")
	})
	t.Run("WriteScopeIsNotPromoted", func(t *testing.T) {
		// An admitted mutation does not hand back detail the credential may not read.
		wr := token(t, "write photos")

		r := AuthenticatedRequestWithBody(app, http.MethodPut, "/api/v1/photos/"+photoUID,
			`{"Favorite": true}`, wr)
		require.Equal(t, http.StatusOK, r.Code)

		res := gjson.Parse(r.Body.String())
		assert.Zero(t, markerCount(res))
		assert.Empty(t, res.Get("Labels").Array())
		assert.Equal(t, "null", res.Get("Details").Raw)
		assert.Empty(t, res.Get("CameraSerial").String())

		g := AuthenticatedRequest(app, http.MethodGet, "/api/v1/photos/"+photoUID, wr)
		assert.Equal(t, http.StatusForbidden, g.Code, "and reading it back is refused outright")
	})
	t.Run("ScopeWithoutPictures", func(t *testing.T) {
		// The scope answers before the payload does, so this credential never reaches the picture.
		r := AuthenticatedRequest(app, http.MethodGet, "/api/v1/photos/"+photoUID, token(t, "metrics"))
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
	// The rows, which is where a ceiling breach would show first. The field-level half is pinned by
	// TestSession_SeesFullDetail/ClientActingForAccount, where the predicate is reachable directly.
	t.Run("AttachedAccountIsACeiling", func(t *testing.T) {
		// A broad client role and an unrestricted scope do not lift the account it acts for.
		guest := clientCredentialToken(t, conf, acl.RoleClient.String(), "*", entity.UserFixtures.Pointer("guest"))
		r := AuthenticatedRequest(app, http.MethodGet, "/api/v1/photos/"+photoUID, guest)
		assert.Equal(t, http.StatusNotFound, r.Code, "the picture is outside that account's scope")

		// And a narrow client role is not lifted by a privileged account either.
		instance := clientCredentialToken(t, conf, acl.RoleInstance.String(), "*", entity.UserFixtures.Pointer("alice"))
		r = AuthenticatedRequest(app, http.MethodGet, "/api/v1/photos/"+photoUID, instance)
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("AddsALabelAndReadsItBack", func(t *testing.T) {
		tk := token(t, "*")

		photo := entity.Photo{}
		require.NoError(t, entity.UnscopedDb().First(&photo, "photo_uid = ?", labelUID).Error)

		t.Cleanup(func() {
			entity.UnscopedDb().Unscoped().
				Exec("DELETE FROM photos_labels WHERE photo_id = ? AND label_id IN (SELECT id FROM labels WHERE label_name = ?)",
					photo.ID, label)
			entity.UnscopedDb().Unscoped().Delete(&entity.Label{}, "label_name = ?", label)
		})

		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/photos/"+labelUID+"/label",
			`{"Name": "`+label+`", "Uncertainty": 25}`, tk)
		require.Equal(t, http.StatusOK, r.Code)

		stored := entity.Label{}
		require.NoError(t, entity.UnscopedDb().First(&stored, "label_name = ?", label).Error)
		require.NotEmpty(t, stored.LabelUID)

		assert.Contains(t, gjson.Get(r.Body.String(), "Labels").Raw, stored.LabelUID,
			"the response names the association it just created")
		assert.Contains(t, picture(t, labelUID, tk).Get("Labels").Raw, stored.LabelUID)
	})
	// The markers this credential now receives are still filtered on people, which its scope does
	// not name - so the two predicates answer separately about the same response.
	t.Run("WithheldPeopleStillOmitted", func(t *testing.T) {
		withheld := entity.SubjectFixtures.Pointer("actress-1")
		markPrivate(t, withheld, false)

		client := picture(t, photoUID, token(t, "photos"))

		assert.NotZero(t, markerCount(client), "everyone else is kept, so an empty list is a failure")
		assert.Less(t, markerCount(client), markerCount(admin))
		assert.NotContains(t, client.Raw, withheld.SubjUID)
	})
}

// TestClientCredential_AlbumAndFileResponses pins the two records a picture points at, which are
// addressed on their own resources and so answer for their own authorization.
func TestClientCredential_AlbumAndFileResponses(t *testing.T) {
	const (
		albumUID = "as6sg6bxpogaaba8"                         // Holiday 2030, shared with the visitor fixture
		fileHash = "3cad9168fa6acc5c5c2965ddf6ec465ca42fd818" // Photo01.dng, which carries an InstanceID
	)

	app, router, conf := NewApiTest()
	GetAlbum(router)
	GetFile(router)
	SearchAlbums(router)

	prevAuthMode := conf.AuthMode()
	conf.SetAuthMode(config.AuthModePasswd)
	t.Cleanup(func() { conf.SetAuthMode(prevAuthMode) })

	// An account-attached credential reaching an album it owns, which is the one path the row
	// policy admits without whole-library reach.
	t.Run("OwnedAlbumNeedsAlbumScope", func(t *testing.T) {
		const photoUID = "ps6sg6be2lvl0yh7"

		GetPhoto(router)
		UpdatePhoto(router)

		user := entity.UserFixtures.Pointer("alice")
		owned := entity.Album{AlbumUID: rnd.GenerateUID(entity.AlbumUID), AlbumSlug: "client-owned-probe",
			AlbumTitle: "Client Owned Probe", AlbumType: entity.AlbumManual, CreatedBy: user.UserUID}

		require.NoError(t, entity.UnscopedDb().Create(&owned).Error)
		require.NoError(t, entity.NewPhotoAlbum(photoUID, owned.AlbumUID).Create())

		t.Cleanup(func() {
			entity.UnscopedDb().Unscoped().Delete(&entity.PhotoAlbum{}, "album_uid = ?", owned.AlbumUID)
			entity.UnscopedDb().Unscoped().Delete(&entity.Album{}, "album_uid = ?", owned.AlbumUID)
			entity.FlushAlbumCache()
		})

		albums := func(t *testing.T, scope string) string {
			r := AuthenticatedRequest(app, http.MethodGet, "/api/v1/photos/"+photoUID,
				clientCredentialToken(t, conf, acl.RoleClient.String(), scope, user))
			require.Equal(t, http.StatusOK, r.Code)

			return gjson.Get(r.Body.String(), "Albums").Raw
		}

		require.Contains(t, albums(t, "photos albums"), owned.AlbumUID,
			"the control: with albums in scope the record arrives")
		assert.NotContains(t, albums(t, "photos"), owned.AlbumUID,
			"and a credential issued for pictures alone receives no album record")

		r := AuthenticatedRequest(app, http.MethodGet, "/api/v1/albums/"+owned.AlbumUID,
			clientCredentialToken(t, conf, acl.RoleClient.String(), "photos", user))
		assert.Equal(t, http.StatusForbidden, r.Code, "as a read of the album itself answers")

		// And the mutation response answers the same, so a write cannot hand back the record.
		photo := entity.Photo{}
		require.NoError(t, entity.UnscopedDb().First(&photo, "photo_uid = ?", photoUID).Error)

		wasFavorite := photo.PhotoFavorite

		t.Cleanup(func() {
			entity.UnscopedDb().Model(&entity.Photo{}).Where("photo_uid = ?", photoUID).
				UpdateColumn("photo_favorite", wasFavorite)
		})

		w := AuthenticatedRequestWithBody(app, http.MethodPut, "/api/v1/photos/"+photoUID,
			`{"Favorite": true}`, clientCredentialToken(t, conf, acl.RoleClient.String(), "photos", user))
		require.Equal(t, http.StatusOK, w.Code)
		assert.NotContains(t, gjson.Get(w.Body.String(), "Albums").Raw, owned.AlbumUID)
	})
	t.Run("AlbumByUID", func(t *testing.T) {
		r := AuthenticatedRequest(app, http.MethodGet, "/api/v1/albums/"+albumUID,
			clientCredentialToken(t, conf, acl.RoleClient.String(), "*", nil))
		require.Equal(t, http.StatusOK, r.Code)
		assert.Equal(t, albumUID, gjson.Get(r.Body.String(), "UID").String())
	})
	t.Run("AlbumListAgreesWithTheRow", func(t *testing.T) {
		// What the listing returns and what a read by uid admits answer the same question.
		r := AuthenticatedRequest(app, http.MethodGet, "/api/v1/albums?count=100&type=album",
			clientCredentialToken(t, conf, acl.RoleClient.String(), "*", nil))
		require.Equal(t, http.StatusOK, r.Code)

		uids := make([]string, 0)
		for _, a := range gjson.Parse(r.Body.String()).Array() {
			uids = append(uids, a.Get("UID").String())
		}

		assert.Contains(t, uids, albumUID)
	})
	t.Run("AlbumOutOfScope", func(t *testing.T) {
		r := AuthenticatedRequest(app, http.MethodGet, "/api/v1/albums/"+albumUID,
			clientCredentialToken(t, conf, acl.RoleClient.String(), "photos", nil))
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
	t.Run("AlbumForANarrowClientRole", func(t *testing.T) {
		r := AuthenticatedRequest(app, http.MethodGet, "/api/v1/albums/"+albumUID,
			clientCredentialToken(t, conf, acl.RoleInstance.String(), "*", nil))
		assert.Equal(t, http.StatusNotFound, r.Code, "it reaches shared albums only, and holds no share")
	})
	t.Run("FileByHash", func(t *testing.T) {
		r := AuthenticatedRequest(app, http.MethodGet, "/api/v1/files/"+fileHash,
			clientCredentialToken(t, conf, acl.RoleClient.String(), "*", nil))
		require.Equal(t, http.StatusOK, r.Code)
		assert.NotEmpty(t, gjson.Get(r.Body.String(), "InstanceID").String(),
			"the fixture has to carry one for the case to mean anything")
	})
	t.Run("FileOnTheFilesResource", func(t *testing.T) {
		// Admitted on files by its scope, so the file answers in full; the picture that carries it
		// would be reduced for the same credential.
		r := AuthenticatedRequest(app, http.MethodGet, "/api/v1/files/"+fileHash,
			clientCredentialToken(t, conf, acl.RoleClient.String(), "files", nil))
		require.Equal(t, http.StatusOK, r.Code)
		assert.NotEmpty(t, gjson.Get(r.Body.String(), "InstanceID").String())
	})
	t.Run("FileMarkersNeedPictureReach", func(t *testing.T) {
		// The markers and the XMP identifier are picture data, so a role admitted on files without
		// whole-library reach on pictures receives the reduced file.
		f := &entity.File{FileUID: "fs6sg6bw45bnlqdw", InstanceID: "xmp-instance-id"}
		s := &entity.Session{}
		s.SetClient(&entity.Client{ClientRole: acl.RoleClient.String(), AuthScope: "*",
			AuthProvider: authn.ProviderClient.String()})

		require.True(t, s.SeesFullDetail(acl.ResourceFiles))
		require.True(t, s.GrantsAny(acl.ResourcePhotos, acl.Permissions{acl.AccessAll, acl.AccessLibrary}),
			"the CE client role reaches both, so this control must not be redacted")
		assert.False(t, f.RedactForSession(s, acl.ResourceFiles).OmitMarkers)
	})
	t.Run("FileOutOfScope", func(t *testing.T) {
		r := AuthenticatedRequest(app, http.MethodGet, "/api/v1/files/"+fileHash,
			clientCredentialToken(t, conf, acl.RoleClient.String(), "photos", nil))
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
}
