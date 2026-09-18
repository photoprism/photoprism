package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
)

func TestAlbumViewableBySession(t *testing.T) {
	entity.ValidateFixtures(t)
	// albumViewableBySession reads only AlbumUID and CreatedBy, so a minimal album is sufficient.
	shared := entity.Album{AlbumUID: "as6sg6bxpogaaba8"}   // redeemed by the visitor fixture
	unshared := entity.Album{AlbumUID: "as6sg6bxpogaaba9"} // not shared with the visitor

	t.Run("NilSession", func(t *testing.T) {
		assert.False(t, albumViewableBySession(nil, shared))
	})
	t.Run("AdminSeesEverything", func(t *testing.T) {
		sess, err := entity.FindSession(entity.SessionFixtures.Get("alice").ID)
		assert.NoError(t, err)
		assert.True(t, albumViewableBySession(sess, unshared))
		t.Cleanup(func() { require.NoError(t, entity.UnscopedDb().Save(entity.SessionFixtures.Pointer("alice")).Error) })
	})
	t.Run("VisitorSharedAlbum", func(t *testing.T) {
		sess, err := entity.FindSession(entity.SessionFixtures.Get("visitor").ID)
		assert.NoError(t, err)
		assert.True(t, albumViewableBySession(sess, shared))
		t.Cleanup(func() { require.NoError(t, entity.UnscopedDb().Save(entity.SessionFixtures.Pointer("visitor")).Error) })
	})
	t.Run("VisitorUnsharedAlbum", func(t *testing.T) {
		sess, err := entity.FindSession(entity.SessionFixtures.Get("visitor").ID)
		assert.NoError(t, err)
		assert.False(t, albumViewableBySession(sess, unshared))
		t.Cleanup(func() { require.NoError(t, entity.UnscopedDb().Save(entity.SessionFixtures.Pointer("visitor")).Error) })
	})
	t.Run("SharedAccessOnlyOwnAlbum", func(t *testing.T) {
		// A registered, shared-access-only user (guest role) may view an album it created even without
		// a share — the self-owned exception the mutation gate deliberately omits.
		guest := entity.UserFixtures.Pointer("guest")
		sess := (&entity.Session{}).SetUser(guest)
		own := entity.Album{AlbumUID: unshared.AlbumUID, CreatedBy: guest.UserUID}
		assert.False(t, sess.NotRegistered())
		assert.True(t, albumViewableBySession(sess, own))
		t.Cleanup(func() { require.NoError(t, entity.UnscopedDb().Save(entity.UserFixtures.Pointer("guest")).Error) })
	})
	t.Run("SharedAccessOnlyOtherAlbum", func(t *testing.T) {
		guest := entity.UserFixtures.Pointer("guest")
		sess := (&entity.Session{}).SetUser(guest)
		other := entity.Album{AlbumUID: unshared.AlbumUID, CreatedBy: "us000000000000zz"}
		assert.False(t, albumViewableBySession(sess, other))
		t.Cleanup(func() { require.NoError(t, entity.UnscopedDb().Save(entity.UserFixtures.Pointer("guest")).Error) })
	})
}

func TestAlbumShareRequired(t *testing.T) {
	entity.ValidateFixtures(t)
	shared := "as6sg6bxpogaaba8"   // redeemed by the visitor fixture
	unshared := "as6sg6bxpogaaba9" // not shared with the visitor

	t.Run("AdminNeverRequiresShare", func(t *testing.T) {
		sess, err := entity.FindSession(entity.SessionFixtures.Get("alice").ID)
		assert.NoError(t, err)
		assert.False(t, albumShareRequired(sess, unshared))
	})
	t.Run("VisitorSharedAlbum", func(t *testing.T) {
		sess, err := entity.FindSession(entity.SessionFixtures.Get("visitor").ID)
		assert.NoError(t, err)
		assert.False(t, albumShareRequired(sess, shared))
	})
	t.Run("VisitorUnsharedAlbum", func(t *testing.T) {
		sess, err := entity.FindSession(entity.SessionFixtures.Get("visitor").ID)
		assert.NoError(t, err)
		assert.True(t, albumShareRequired(sess, unshared))
	})
	t.Run("SharedAccessOnlyOwnAlbumStillRequiresShare", func(t *testing.T) {
		// Unlike the view gate, creating an album grants a shared-access-only user no right to modify it
		// without a share, so the mutation gate stays required even for its own unshared album.
		guest := entity.UserFixtures.Pointer("guest")
		sess := (&entity.Session{}).SetUser(guest)
		assert.True(t, albumShareRequired(sess, unshared))
	})
}
