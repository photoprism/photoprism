package api

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/entity/search"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/fs"
)

// TestSharePreview_StoredFilter covers what a share link renders for a smart album: the album's
// stored filter selects the content, but it cannot reach content the share is meant to exclude.
func TestSharePreview_StoredFilter(t *testing.T) {
	app, router, conf := NewApiTest()
	conf.SetAuthMode(config.AuthModePasswd)
	defer conf.SetAuthMode(config.AuthModePublic)
	SharePreview(router)

	// The same private fixture the query-level cases use, so both operate on known content.
	var private entity.Photo
	require.NoError(t, entity.Db().Where("photo_uid = ?", "ps6sg6be2lvl0y13").First(&private).Error)
	require.True(t, private.PhotoPrivate)

	album := entity.NewAlbum("Share Preview Stored Filter", entity.AlbumMoment)
	album.AlbumFilter = "public:false private:true uid:" + private.PhotoUID
	require.NoError(t, album.Create())
	t.Cleanup(func() {
		_ = entity.UnscopedDb().Delete(album).Error
		entity.FlushAlbumCache()
	})

	link := entity.NewUserLink(album.AlbumUID, entity.Admin.UserUID)
	require.NoError(t, link.Save())
	t.Cleanup(func() { _ = entity.UnscopedDb().Delete(link).Error })

	require.Len(t, entity.FindRedeemableLinksByToken(link.LinkToken, album.AlbumUID), 1)

	// Fixture sanity: the album's filter matches content, so an empty preview below reflects the
	// selection contract rather than an album that matches nothing.
	sanity := form.SearchPhotos{Album: album.AlbumUID, Primary: true, Count: 6}
	require.NoError(t, sanity.ParseQueryString())

	selected, _, err := search.Photos(sanity)
	require.NoError(t, err)
	require.NotEmpty(t, selected, "the album filter must match the fixture")

	// A cached preview from an earlier run would answer before the query runs at all.
	previewFile := filepath.Join(path.Join(conf.ThumbCachePath(), "share"), album.AlbumUID+fs.ExtJpeg)
	_ = os.Remove(previewFile)
	t.Cleanup(func() { _ = os.Remove(previewFile) })

	r := PerformRequest(app, "GET", "/api/v1/"+link.LinkToken+"/"+album.AlbumUID+"/preview")

	// No selectable content leaves the handler with nothing to render, so it redirects to the
	// site preview rather than composing an image from the private photo. This is a smoke check:
	// the fixture has no rendered thumbnail either, so the discriminating assertion is the cover
	// query below, and the selection itself is covered in internal/entity/search.
	assert.Equal(t, http.StatusTemporaryRedirect, r.Code)
	assert.NoFileExists(t, previewFile)

	// The query behind /albums/{uid}/t/{token}/{size} must reach the same conclusion.
	_, err = query.AlbumCoverByUID(album.AlbumUID, true)
	assert.Error(t, err, "a cover must not resolve to private content")
}
