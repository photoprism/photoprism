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

		assert.Equal(t, http.StatusOK, result.Code)
	})
	t.Run("invalid request", func(t *testing.T) {
		app, router, conf := NewApiTest()
		GetAlbums(router, conf)
		result := PerformRequest(app, "GET", "/api/v1/albums?xxx=10")
		t.Log(result.Body)

		assert.Equal(t, http.StatusBadRequest, result.Code)
	})
	t.Run("invalid request", func(t *testing.T) {
		app, router, conf := NewApiTest()
		t.Log(router)
		t.Log(conf)
		result := PerformRequest(app, "GET", "/api/v1/albums?xxx=10")
		t.Log(result.Body)

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

}

func TestDislikeAlbum(t *testing.T) {
	t.Run("dislike not existing album", func(t *testing.T) {
		app, router, conf := NewApiTest()

		LikeAlbum(router, conf)

		result := PerformRequest(app, "DELETE", "/api/v1/albums/5678/like")
		t.Log(result.Body)
		assert.Equal(t, http.StatusNotFound, result.Code)
	})
}


func TestAlbumThumbnail(t *testing.T) {
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
