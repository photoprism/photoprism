package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/auth/acl"
)

// sharedAlbumUID is the album the visitor fixture holds a share for.
const sharedAlbumUID = "as6sg6bxpogaaba8"

func TestAlbum_SharedWithSession(t *testing.T) {
	shared := Album{AlbumUID: sharedAlbumUID}
	other := Album{AlbumUID: "as6sg6bxpogaaba9"}

	t.Run("Share", func(t *testing.T) {
		visitor := SessionFixtures.Pointer("visitor")
		assert.True(t, shared.SharedWithSession(visitor))
		assert.False(t, other.SharedWithSession(visitor))
	})
	t.Run("OwnAlbum", func(t *testing.T) {
		s := &Session{}
		s.SetUser(UserFixtures.Pointer("guest"))

		owned := Album{AlbumUID: "as6sg6bxpogaabcd", CreatedBy: s.UserUID}
		assert.True(t, owned.SharedWithSession(s))
		assert.False(t, other.SharedWithSession(s))
	})
	t.Run("UnownedAlbumIsNotOwnedByANamelessSession", func(t *testing.T) {
		unowned := Album{AlbumUID: "as6sg6bxpogaabce", CreatedBy: ""}
		assert.False(t, unowned.SharedWithSession(fullAccessClientSession()))
		assert.False(t, unowned.SharedWithSession(&Session{}))
	})
	t.Run("AdminWithoutShare", func(t *testing.T) {
		s := &Session{}
		s.SetUser(UserFixtures.Pointer("alice"))
		assert.False(t, other.SharedWithSession(s), "this half answers for shares alone")
	})
	t.Run("NilSession", func(t *testing.T) {
		assert.False(t, shared.SharedWithSession(nil))
	})
}

func TestAlbum_VisibleToSession(t *testing.T) {
	album := Album{AlbumUID: "as6sg6bxpogaaba9"}

	t.Run("Admin", func(t *testing.T) {
		s := &Session{}
		s.SetUser(UserFixtures.Pointer("alice"))
		assert.True(t, album.VisibleToSession(s))
	})
	t.Run("Guest", func(t *testing.T) {
		s := &Session{}
		s.SetUser(UserFixtures.Pointer("guest"))
		assert.False(t, album.VisibleToSession(s))
	})
	t.Run("VisitorWithShare", func(t *testing.T) {
		visitor := SessionFixtures.Pointer("visitor")
		assert.True(t, (&Album{AlbumUID: sharedAlbumUID}).VisibleToSession(visitor))
		assert.False(t, album.VisibleToSession(visitor))
	})
	t.Run("Credential", func(t *testing.T) {
		assert.True(t, album.VisibleToSession(fullAccessClientSession()),
			"a credential the role admits on the library is not a visitor for having no account")
		// The row policy reads the role alone, so an album download keeps the resource it was
		// authorized against. The scope answers on the handler, and for a nested album record.
		assert.True(t, album.VisibleToSession(SessionFixtures.Pointer("client_metrics")))
	})
	t.Run("ClientRoleWithoutAlbums", func(t *testing.T) {
		s := &Session{}
		s.SetClient(&Client{ClientRole: acl.RoleInstance.String(), AuthScope: "*"})
		assert.False(t, album.VisibleToSession(s))
	})
	t.Run("NilSession", func(t *testing.T) {
		assert.False(t, album.VisibleToSession(nil))
	})
}

func TestSharedAlbums(t *testing.T) {
	albums := []Album{{AlbumUID: sharedAlbumUID}, {AlbumUID: "as6sg6bxpogaaba9"}}

	t.Run("Share", func(t *testing.T) {
		kept := SharedAlbums(append([]Album{}, albums...), SessionFixtures.Pointer("visitor"))
		assert.Len(t, kept, 1)
		assert.Equal(t, sharedAlbumUID, kept[0].AlbumUID)
	})
	t.Run("None", func(t *testing.T) {
		// Nil rather than an empty slice, so a picture that keeps no album serializes as it always
		// has.
		assert.Nil(t, SharedAlbums(append([]Album{}, albums...), &Session{}))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Nil(t, SharedAlbums(nil, SessionFixtures.Pointer("visitor")))
	})
	t.Run("InputIsNotRewritten", func(t *testing.T) {
		// The kept album sits after the dropped one, so compacting in place would overwrite the
		// first element of the caller's slice.
		in := []Album{{AlbumUID: "as6sg6bxpogaaba9"}, {AlbumUID: sharedAlbumUID}}

		require.Len(t, SharedAlbums(in, SessionFixtures.Pointer("visitor")), 1)
		assert.Equal(t, "as6sg6bxpogaaba9", in[0].AlbumUID, "the caller's slice is left alone")
	})
}
