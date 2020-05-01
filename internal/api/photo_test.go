package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func TestGetPhoto(t *testing.T) {
	t.Run("search for existing photo", func(t *testing.T) {
		app, router, ctx := NewApiTest()
		GetPhoto(router, ctx)
		result := PerformRequest(app, "GET", "/api/v1/photos/pt9jtdre2lvl0yh7")
		assert.Equal(t, http.StatusOK, result.Code)
		val := gjson.Get(result.Body.String(), "PhotoLat")
		assert.Equal(t, "48.519234", val.String())
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
		result := PerformRequest(app, "GET", "/api/v1/photos/pt9jtdre2lvl0yh7/download")
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
		result := PerformRequest(app, "POST", "/api/v1/photos/pt9jtdre2lvl0yh7/like")
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
		result := PerformRequest(app, "DELETE", "/api/v1/photos/pt9jtdre2lvl0yh8/like")
		assert.Equal(t, http.StatusOK, result.Code)
	})
	t.Run("not existing photo", func(t *testing.T) {
		app, router, ctx := NewApiTest()
		DislikePhoto(router, ctx)
		result := PerformRequest(app, "DELETE", "/api/v1/photos/xxx/like")
		assert.Equal(t, http.StatusNotFound, result.Code)
	})
}
