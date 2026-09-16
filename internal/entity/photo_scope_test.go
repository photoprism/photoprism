package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/pkg/authn"
)

func TestPhoto_RedactForSession(t *testing.T) {
	// newPhoto returns a photo populated with the fields RedactForSession may trim.
	newPhoto := func() *Photo {
		return &Photo{
			CreatedBy:    "uqxetse3cy5eo9z2",
			PhotoPath:    "2020/01",
			OriginalName: "orig.jpg",
			UUID:         "a1b2c3d4-document-id",
			CameraSerial: "SN-123456",
			Albums:       []Album{{AlbumUID: "as6sg6bxpogaaba9"}, {AlbumUID: "as6sg6bxpogaaba8"}},
			Labels:       []PhotoLabel{{}},
			Details:      &Details{},
			Files:        []File{{FileUID: "fs6sg6bw45bnlqdw", FileName: "2020/01/orig.jpg", InstanceID: "xmp-instance-id"}},
		}
	}

	session := func(name string) *Session {
		s := &Session{}
		s.SetUser(UserFixtures.Pointer(name))
		return s
	}

	t.Run("AdminUnchanged", func(t *testing.T) {
		p := newPhoto()
		p.RedactForSession(session("alice"))
		assert.Len(t, p.Albums, 2)
		assert.Len(t, p.Labels, 1)
		assert.Equal(t, "uqxetse3cy5eo9z2", p.CreatedBy)
		assert.Equal(t, "a1b2c3d4-document-id", p.UUID)
		assert.Equal(t, "SN-123456", p.CameraSerial)
		assert.Equal(t, "xmp-instance-id", p.Files[0].InstanceID)
		assert.NotNil(t, p.Details)
		assert.False(t, p.Files[0].OmitMarkers)
	})
	t.Run("NilSession", func(t *testing.T) {
		p := newPhoto()
		p.RedactForSession(nil)
		assert.Len(t, p.Albums, 2)
		assert.Equal(t, "uqxetse3cy5eo9z2", p.CreatedBy)
	})
	t.Run("FullAccessClientUnchanged", func(t *testing.T) {
		p := newPhoto()
		p.RedactForSession(fullAccessClientSession())
		assert.Len(t, p.Albums, 2)
		assert.Len(t, p.Labels, 1)
		assert.Equal(t, "uqxetse3cy5eo9z2", p.CreatedBy)
		assert.Equal(t, "a1b2c3d4-document-id", p.UUID)
		assert.Equal(t, "SN-123456", p.CameraSerial)
		assert.Equal(t, "xmp-instance-id", p.Files[0].InstanceID)
		assert.NotNil(t, p.Details)
		assert.False(t, p.Files[0].OmitMarkers, "a credential admitted on the library keeps the markers")
	})
	t.Run("NarrowlyScopedClientRedacted", func(t *testing.T) {
		p := newPhoto()
		p.RedactForSession(SessionFixtures.Pointer("client_metrics"))
		assert.Empty(t, p.Albums)
		assert.Empty(t, p.Labels)
		assert.Nil(t, p.Details)
		assert.True(t, p.Files[0].OmitMarkers)
	})
	t.Run("OwnAlbumKept", func(t *testing.T) {
		// The nested list answers the question a read of the album itself answers, so an album this
		// account created stays even though the rest of the picture is reduced.
		s := session("guest")

		p := newPhoto()
		p.Albums[0].CreatedBy = s.UserUID
		p.RedactForSession(s)

		require.Len(t, p.Albums, 1)
		assert.Equal(t, "as6sg6bxpogaaba9", p.Albums[0].AlbumUID)
		assert.Nil(t, p.Details, "and the picture is reduced around it")
	})
	t.Run("OwnAlbumNeedsAlbumScope", func(t *testing.T) {
		// Ownership reaches the record, but a credential issued for pictures alone was never
		// admitted on albums, so it receives none - as a read of the album itself would answer.
		user := UserFixtures.Pointer("alice")

		s := &Session{}
		s.SetClient(&Client{ClientRole: acl.RoleClient.String(), AuthScope: "photos",
			AuthProvider: authn.ProviderClient.String()})
		s.SetUser(user)

		p := newPhoto()
		p.Albums[0].CreatedBy = user.UserUID
		p.RedactForSession(s)

		assert.Empty(t, p.Albums)

		// The positive control: the same session and album with albums in scope.
		s.SetScope("photos albums")

		p = newPhoto()
		p.Albums[0].CreatedBy = user.UserUID
		p.RedactForSession(s)

		assert.Len(t, p.Albums, 2, "an account that reaches every album keeps them all")
	})
	t.Run("SharedAlbumSurvivesForAVisitor", func(t *testing.T) {
		// The other direction: a share-link visitor is not a credential and keeps what it shares.
		p := newPhoto()
		p.RedactForSession(SessionFixtures.Pointer("visitor"))

		require.Len(t, p.Albums, 1)
		assert.Equal(t, "as6sg6bxpogaaba8", p.Albums[0].AlbumUID)
	})
	t.Run("GuestRedacted", func(t *testing.T) {
		p := newPhoto()
		p.RedactForSession(session("guest"))
		// This guest holds no share and created neither album, which is what the album records
		// answer on.
		assert.Empty(t, p.Albums)
		assert.Empty(t, p.Labels)
		assert.Equal(t, "", p.CreatedBy)
		assert.Equal(t, "", p.UUID)
		assert.Equal(t, "", p.CameraSerial)
		assert.Equal(t, "", p.Files[0].InstanceID)
		assert.Nil(t, p.Details)
		assert.True(t, p.Files[0].OmitMarkers)
		// Storage path and file names stay — search returns them to every in-scope session.
		assert.Equal(t, "2020/01", p.PhotoPath)
		assert.Equal(t, "orig.jpg", p.OriginalName)
		assert.Equal(t, "2020/01/orig.jpg", p.Files[0].FileName)
	})
}
