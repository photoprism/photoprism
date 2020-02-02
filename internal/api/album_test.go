package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAlbums(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		app, router, conf := NewApiTest()
		GetAlbums(router, conf)
		result := PerformRequest(app, "GET", "/api/v1/albums?count=10")
		assert.Contains(t, result.Body.String(), "Christmas2030")
		assert.Equal(t, http.StatusOK, result.Code)
	})
	t.Run("invalid request", func(t *testing.T) {
		app, router, conf := NewApiTest()
		GetAlbums(router, conf)
		result := PerformRequest(app, "GET", "/api/v1/albums?xxx=10")

		assert.Equal(t, http.StatusBadRequest, result.Code)
	})
}

func TestGetAlbum(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		app, router, conf := NewApiTest()
		GetAlbum(router, conf)
		result := PerformRequest(app, "GET", "/api/v1/albums/4")
		assert.Contains(t, result.Body.String(), "holiday-2030")
		assert.Equal(t, http.StatusOK, result.Code)
	})
	t.Run("invalid request", func(t *testing.T) {
		app, router, conf := NewApiTest()
		GetAlbum(router, conf)
		result := PerformRequest(app, "GET", "/api/v1/albums/999000")
		assert.Equal(t, http.StatusNotFound, result.Code)
	})
}

func TestDeleteAlbum(t *testing.T) {
	t.Run("delete existing album", func(t *testing.T) {
		app, router, conf := NewApiTest()
		DeleteAlbum(router, conf)
		result := PerformRequest(app, "DELETE", "/api/v1/albums/5")
		assert.Equal(t, http.StatusOK, result.Code)
		assert.Contains(t, result.Body.String(), "Berlin2019")
		GetAlbums(router, conf)
		result2 := PerformRequest(app, "GET", "/api/v1/albums?count=10")
		assert.NotContains(t, result2.Body.String(), "Berlin2019")
	})
	t.Run("delete not existing album", func(t *testing.T) {
		app, router, conf := NewApiTest()
		DeleteAlbum(router, conf)
		result := PerformRequest(app, "DELETE", "/api/v1/albums/999000")
		assert.Equal(t, http.StatusNotFound, result.Code)
	})
}

func TestLikeAlbum(t *testing.T) {
	t.Run("like not existing album", func(t *testing.T) {
		app, router, ctx := NewApiTest()

		LikeAlbum(router, ctx)

		result := PerformRequest(app, "POST", "/api/v1/albums/98789876/like")
		assert.Equal(t, http.StatusNotFound, result.Code)
	})
	t.Run("like existing album", func(t *testing.T) {
		app, router, ctx := NewApiTest()

		LikeAlbum(router, ctx)

		result := PerformRequest(app, "POST", "/api/v1/albums/3/like")
		assert.Equal(t, http.StatusOK, result.Code)
	})
}

func TestDislikeAlbum(t *testing.T) {
	t.Run("dislike not existing album", func(t *testing.T) {
		app, router, conf := NewApiTest()

		DislikeAlbum(router, conf)

		result := PerformRequest(app, "DELETE", "/api/v1/albums/5678/like")
		assert.Equal(t, http.StatusNotFound, result.Code)
	})
	t.Run("dislike existing album", func(t *testing.T) {
		app, router, conf := NewApiTest()

		DislikeAlbum(router, conf)

		result := PerformRequest(app, "DELETE", "/api/v1/albums/4/like")
		assert.Equal(t, http.StatusOK, result.Code)
	})
}

func TestDownloadAlbum(t *testing.T) {
	t.Run("download not existing album", func(t *testing.T) {
		app, router, conf := NewApiTest()

		DownloadAlbum(router, conf)

		result := PerformRequest(app, "GET", "/api/v1/albums/5678/download")
		assert.Equal(t, http.StatusNotFound, result.Code)
	})
	t.Run("download existing album", func(t *testing.T) {
		app, router, conf := NewApiTest()

		DownloadAlbum(router, conf)

		result := PerformRequest(app, "GET", "/api/v1/albums/4/download")
		assert.Equal(t, http.StatusOK, result.Code)
	})
}

func TestAlbumThumbnail(t *testing.T) {
	t.Run("could not find original", func(t *testing.T) {
		app, router, ctx := NewApiTest()
		AlbumThumbnail(router, ctx)
		result := PerformRequest(app, "GET", "/api/v1/albums/4/thumbnail/tile_500")
		assert.Equal(t, http.StatusNotFound, result.Code)
	})
	t.Run("invalid type", func(t *testing.T) {
		app, router, ctx := NewApiTest()
		AlbumThumbnail(router, ctx)
		result := PerformRequest(app, "GET", "/api/v1/albums/1/thumbnail/xxx")

		assert.Equal(t, http.StatusBadRequest, result.Code)
	})
	t.Run("album has no photo (because is not existing)", func(t *testing.T) {
		app, router, ctx := NewApiTest()
		AlbumThumbnail(router, ctx)
		result := PerformRequest(app, "GET", "/api/v1/albums/987-986435/thumbnail/tile_500")

		assert.Equal(t, http.StatusOK, result.Code)
	})
}
