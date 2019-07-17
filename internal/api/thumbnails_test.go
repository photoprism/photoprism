package api

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestGetThumbnail(t *testing.T) {
	t.Run("invalid type", func(t *testing.T) {
		app, router, ctx := NewApiTest()
		GetThumbnail(router, ctx)
		result := PerformRequest(app, "GET", "/api/v1/thumbnails/1/xxx")

		assert.Equal(t, http.StatusBadRequest, result.Code)
	})
	t.Run("invalid hash", func(t *testing.T) {
		app, router, ctx := NewApiTest()
		GetThumbnail(router, ctx)
		result := PerformRequest(app, "GET", "/api/v1/thumbnails/1/tile_500")

		assert.Equal(t, http.StatusNotFound, result.Code)
	})
}

func TestLabelThumbnail(t *testing.T) {
	t.Run("invalid type", func(t *testing.T) {
		app, router, ctx := NewApiTest()
		LabelThumbnail(router, ctx)
		result := PerformRequest(app, "GET", "/api/v1/labels/dog/thumbnail/xxx")

		assert.Equal(t, http.StatusBadRequest, result.Code)
	})
	t.Run("invalid label", func(t *testing.T) {
		app, router, ctx := NewApiTest()
		LabelThumbnail(router, ctx)
		result := PerformRequest(app, "GET", "/api/v1/labels/xxx/thumbnail/tile_500")

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
