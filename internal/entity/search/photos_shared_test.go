package search

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
)

// sharedFilterAlbum creates a smart album selecting exactly the given private photo, plus a share
// link for it, and returns the album. Both are removed when the case ends.
func sharedFilterAlbum(t *testing.T, filter string) *entity.Album {
	t.Helper()

	album := entity.NewAlbum("Shared Filter Boundary", entity.AlbumMoment)
	album.AlbumFilter = filter
	require.NoError(t, album.Create())

	t.Cleanup(func() {
		_ = entity.UnscopedDb().Delete(album).Error
		entity.FlushAlbumCache()
	})

	link := entity.NewUserLink(album.AlbumUID, entity.UserFixtures.Pointer("alice").UserUID)
	require.NoError(t, link.Save())
	t.Cleanup(func() { _ = entity.UnscopedDb().Delete(link).Error })

	require.Len(t, entity.FindRedeemableLinksByToken(link.LinkToken, album.AlbumUID), 1)

	return album
}

// TestSharedPhotos covers the selection a share preview or album cover renders: a smart album's
// stored filter may narrow it, but the public-only constraints hold whatever the filter asks for.
func TestSharedPhotos(t *testing.T) {
	var private entity.Photo
	require.NoError(t, entity.Db().Where("photo_uid = ?", scopePrivatePhotoUID).First(&private).Error)
	require.True(t, private.PhotoPrivate)

	t.Run("FilterCannotSelectPrivateContent", func(t *testing.T) {
		album := sharedFilterAlbum(t, "public:false private:true uid:"+private.PhotoUID)

		// The form a share preview submits, immediately before the search.
		frm := form.SearchPhotos{
			Album: album.AlbumUID, Public: true, Private: false, Hidden: false,
			Archived: false, Review: false, Primary: true, Count: 6, Order: album.AlbumOrder,
		}
		require.NoError(t, frm.ParseQueryString())

		rows, _, err := SharedPhotos(frm)
		require.NoError(t, err)
		assert.Empty(t, rows, "a stored filter must not widen what a share exposes")
	})
	t.Run("FilterCannotSelectHiddenContent", func(t *testing.T) {
		album := sharedFilterAlbum(t, "hidden:true")

		frm := form.SearchPhotos{Album: album.AlbumUID, Public: true, Primary: true, Count: 6}
		require.NoError(t, frm.ParseQueryString())

		rows, _, err := SharedPhotos(frm)
		require.NoError(t, err)
		require.NotEmpty(t, rows, "the filter must select something, or the case proves nothing")

		for _, row := range rows {
			assert.NotEqual(t, -1, row.PhotoQuality, "hidden content must not be rendered")
		}
	})
	t.Run("FilterCannotSelectContentInReview", func(t *testing.T) {
		// Review selects photo_quality < 3, which contains the hidden set, so this decides
		// whether the hidden constraint holds at all.
		album := sharedFilterAlbum(t, "review:true")

		frm := form.SearchPhotos{Album: album.AlbumUID, Public: true, Primary: true, Count: 6}
		require.NoError(t, frm.ParseQueryString())

		rows, _, err := SharedPhotos(frm)
		require.NoError(t, err)
		require.NotEmpty(t, rows, "the filter must select something, or the case proves nothing")

		for _, row := range rows {
			assert.GreaterOrEqual(t, row.PhotoQuality, 3, "content in review must not be rendered")
		}
	})
	t.Run("FilterCannotSelectArchivedContent", func(t *testing.T) {
		album := sharedFilterAlbum(t, "archived:true")

		frm := form.SearchPhotos{Album: album.AlbumUID, Public: true, Primary: true, Count: 6}
		require.NoError(t, frm.ParseQueryString())

		rows, _, err := SharedPhotos(frm)
		require.NoError(t, err)
		require.NotEmpty(t, rows, "the filter must select something, or the case proves nothing")

		for _, row := range rows {
			assert.Nil(t, row.DeletedAt, "archived content must not be rendered")
		}
	})
	t.Run("PublicContentIsStillSelected", func(t *testing.T) {
		var public entity.Photo
		require.NoError(t, entity.Db().Where("photo_private = 0 AND photo_quality >= 3 AND deleted_at IS NULL").
			First(&public).Error)

		album := sharedFilterAlbum(t, "uid:"+public.PhotoUID)

		frm := form.SearchPhotos{Album: album.AlbumUID, Public: true, Primary: true, Count: 6}
		require.NoError(t, frm.ParseQueryString())

		rows, _, err := SharedPhotos(frm)
		require.NoError(t, err)
		require.Len(t, rows, 1, "a filter that selects public content must still work")
		assert.Equal(t, public.PhotoUID, rows[0].PhotoUID)
	})
	t.Run("UnscopedSearchIsUnchanged", func(t *testing.T) {
		// Internal callers that legitimately need everything keep the unscoped entry point.
		album := sharedFilterAlbum(t, "public:false private:true uid:"+private.PhotoUID)

		frm := form.SearchPhotos{Album: album.AlbumUID, Primary: true, Count: 6}
		require.NoError(t, frm.ParseQueryString())

		rows, _, err := Photos(frm)
		require.NoError(t, err)
		require.Len(t, rows, 1)
		assert.Equal(t, private.PhotoUID, rows[0].PhotoUID)
	})
}
