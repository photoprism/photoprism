package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPhotos(t *testing.T) {
	// TODO assert for json response
	t.Run("successfulrequest", func(t *testing.T) {
		app, router, ctx := NewApiTest()

		GetPhotos(router, ctx)

		result := PerformRequest(app, "GET", "/api/v1/photos?count=10")
		assert.Equal(t, http.StatusOK, result.Code)
	})

	t.Run("invalid request", func(t *testing.T) {
		app, router, ctx := NewApiTest()
		GetPhotos(router, ctx)
		result := PerformRequest(app, "GET", "/api/v1/photos?xxx=10")
		t.Log(result.Body)

		assert.Equal(t, http.StatusBadRequest, result.Code)
	})
	t.Run("invalid request", func(t *testing.T) {
		app, router, ctx := NewApiTest()
		t.Log(router)
		t.Log(ctx)
		result := PerformRequest(app, "GET", "/api/v1/photos?count=10")
		t.Log(result.Body)

		assert.Equal(t, http.StatusNotFound, result.Code)
	})

}

func TestLikePhoto(t *testing.T) {
	t.Run("like if resultCode is not 404", func(t *testing.T) {
		app, router, ctx := NewApiTest()

		LikePhoto(router, ctx)

		result := PerformRequest(app, "POST", "/api/v1/photos/1/like")
		t.Log(result.Body)

		// TODO: Test database can be empty
		if result.Code != http.StatusNotFound {
			assert.Equal(t, http.StatusOK, result.Code)
		}
	})
	t.Run("like not existing photo", func(t *testing.T) {
		app, router, ctx := NewApiTest()

		LikePhoto(router, ctx)

		result := PerformRequest(app, "POST", "/api/v1/photos/98789876/like")
		t.Log(result.Body)
		assert.Equal(t, http.StatusNotFound, result.Code)
	})

}

func TestDislikePhoto(t *testing.T) {
	t.Run("dislike if resultCode is not 404", func(t *testing.T) {
		app, router, ctx := NewApiTest()

		DislikePhoto(router, ctx)

		result := PerformRequest(app, "DELETE", "/api/v1/photos/1/like")
		t.Log(result.Body)

		// TODO: Test database can be empty
		if result.Code != http.StatusNotFound {
			assert.Equal(t, http.StatusOK, result.Code)
		}
	})
	t.Run("dislike not existing photo", func(t *testing.T) {
		app, router, ctx := NewApiTest()

		LikePhoto(router, ctx)

		result := PerformRequest(app, "DELETE", "/api/v1/photos/98789876/like")
		t.Log(result.Body)
		assert.Equal(t, http.StatusNotFound, result.Code)
	})
}
