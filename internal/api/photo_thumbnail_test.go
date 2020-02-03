package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
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
	t.Run("could not find original", func(t *testing.T) {
		app, router, ctx := NewApiTest()
		GetThumbnail(router, ctx)
		result := PerformRequest(app, "GET", "/api/v1/thumbnails/123xxx/tile_500")

		assert.Equal(t, http.StatusNotFound, result.Code)
	})
}
