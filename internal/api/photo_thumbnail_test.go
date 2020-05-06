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
		r := PerformRequest(app, "GET", "/api/v1/thumbnails/1/xxx")

		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("invalid hash", func(t *testing.T) {
		app, router, ctx := NewApiTest()
		GetThumbnail(router, ctx)
		r := PerformRequest(app, "GET", "/api/v1/thumbnails/1/tile_500")

		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("could not find original", func(t *testing.T) {
		app, router, ctx := NewApiTest()
		GetThumbnail(router, ctx)
		r := PerformRequest(app, "GET", "/api/v1/thumbnails/2cad9168fa6acc5c5c2965ddf6ec465ca42fd818/tile_500")
		assert.Equal(t, http.StatusOK, r.Code)
	})
}
