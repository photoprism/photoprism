package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
)

// TestGetAlbumHeader checks cover resolution and scoped access.
func TestGetAlbumHeader(t *testing.T) {
	t.Run("PublishedCover", func(t *testing.T) {
		app, router, _ := NewApiTest()
		GetAlbumHeader(router)
		a := entity.AlbumFixtures.Get("holiday-2030")
		file := entity.FileFixtures.Get("exampleFileName.jpg")
		var original entity.Album
		require.NoError(t, entity.Db().First(&original, "album_uid = ?", a.AlbumUID).Error)
		require.NoError(t, entity.UpdateAlbum(a.AlbumUID, entity.Values{"thumb": file.FileHash}))
		t.Cleanup(func() { _ = entity.UpdateAlbum(a.AlbumUID, entity.Values{"thumb": original.Thumb}) })
		r := PerformRequest(app, "GET", "/api/v1/albums/"+a.AlbumUID+"/header")
		require.Equal(t, http.StatusOK, r.Code)
		assert.Equal(t, file.FileHash, gjson.Get(r.Body.String(), "Hash").String())
		assert.Equal(t, int64(file.FileWidth), gjson.Get(r.Body.String(), "Width").Int())
		assert.Equal(t, "no-store", r.Header().Get("Cache-Control"))
		assert.NotContains(t, r.Body.String(), "SubjUID")
		assert.NotContains(t, r.Body.String(), "FileUID")
	})
	t.Run("Missing", func(t *testing.T) {
		app, router, _ := NewApiTest()
		GetAlbumHeader(router)
		r := PerformRequest(app, "GET", "/api/v1/albums/invalid/header")
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("GuestScope", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		GetAlbumHeader(router)
		sid := AuthenticateUser(app, router, "gandalf", "Gandalf123!")
		r := AuthenticatedRequest(app, "GET", "/api/v1/albums/as6sg6bxpogaaba8/header", sid)
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
}

// TestAlbumHeaderFaces checks that invalid rectangles are omitted and valid regions are enclosed.
func TestAlbumHeaderFaces(t *testing.T) {
	assert.Nil(t, albumHeaderFaces(nil))
	assert.Nil(t, albumHeaderFaces(entity.Markers{{W: 0, H: 1}}))
	r := albumHeaderFaces(entity.Markers{{X: 0.1, Y: 0.2, W: 0.3, H: 0.2}, {X: 0.6, Y: 0.1, W: 0.5, H: 0.5}})
	require.NotNil(t, r)
	assert.InDelta(t, 0.1, r.X, 0.00001)
	assert.InDelta(t, 0.1, r.Y, 0.00001)
	assert.InDelta(t, 0.9, r.W, 0.00001)
	assert.InDelta(t, 0.5, r.H, 0.00001)
}
