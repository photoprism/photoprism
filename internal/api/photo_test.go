package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPhoto(t *testing.T) {
	t.Run("search for existing photo", func(t *testing.T) {
		app, router, ctx := NewApiTest()
		GetPhoto(router, ctx)
		result := PerformRequest(app, "GET", "/api/v1/photos/654")
		assert.Equal(t, http.StatusOK, result.Code)
		assert.Contains(t, result.Body.String(), "\"PhotoLat\":48.519235")
	})
	t.Run("search for not existing photo", func(t *testing.T) {
		app, router, ctx := NewApiTest()
		GetPhoto(router, ctx)
		result := PerformRequest(app, "GET", "/api/v1/photos/xxx")
		assert.Equal(t, http.StatusNotFound, result.Code)
	})
}

func TestGetPhotoDownload(t *testing.T) {
	t.Run("could not find original", func(t *testing.T) {
		app, router, ctx := NewApiTest()
		GetPhotoDownload(router, ctx)
		result := PerformRequest(app, "GET", "/api/v1/photos/654/download")
		assert.Equal(t, http.StatusNotFound, result.Code)
	})
	t.Run("not existing photo", func(t *testing.T) {
		app, router, ctx := NewApiTest()
		GetPhotoDownload(router, ctx)
		result := PerformRequest(app, "GET", "/api/v1/photos/xxx/download")
		assert.Equal(t, http.StatusNotFound, result.Code)
	})
}

func TestLikePhoto(t *testing.T) {
	t.Run("existing photo", func(t *testing.T) {
		app, router, ctx := NewApiTest()
		LikePhoto(router, ctx)
		result := PerformRequest(app, "POST", "/api/v1/photos/654/like")
		assert.Equal(t, http.StatusOK, result.Code)
	})
	t.Run("not existing photo", func(t *testing.T) {
		app, router, ctx := NewApiTest()
		LikePhoto(router, ctx)
		result := PerformRequest(app, "POST", "/api/v1/photos/xxx/like")
		assert.Equal(t, http.StatusNotFound, result.Code)
	})
}

func TestDislikePhoto(t *testing.T) {
	t.Run("existing photo", func(t *testing.T) {
		app, router, ctx := NewApiTest()
		DislikePhoto(router, ctx)
		result := PerformRequest(app, "DELETE", "/api/v1/photos/655/like")
		assert.Equal(t, http.StatusOK, result.Code)
	})
	t.Run("not existing photo", func(t *testing.T) {
		app, router, ctx := NewApiTest()
		DislikePhoto(router, ctx)
		result := PerformRequest(app, "DELETE", "/api/v1/photos/xxx/like")
		assert.Equal(t, http.StatusNotFound, result.Code)
	})
}
