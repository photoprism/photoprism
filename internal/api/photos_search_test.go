package api

import (
	"net/http"
	"testing"

	"github.com/tidwall/gjson"

	"github.com/stretchr/testify/assert"
)

func TestSearchPhotos(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, router, _ := NewApiTest()
		SearchPhotos(router)
		r := PerformRequest(app, "GET", "/api/v1/photos?count=10")
		count := gjson.Get(r.Body.String(), "#")
		assert.LessOrEqual(t, int64(2), count.Int())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("ViewerJSON", func(t *testing.T) {
		app, router, _ := NewApiTest()
		SearchPhotos(router)
		r := PerformRequest(app, "GET", "/api/v1/photos?count=10&format=view")
		body := r.Body.String()

		t.Logf("response body: %s", body)

		count := gjson.Get(body, "#")
		assert.LessOrEqual(t, int64(2), count.Int())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		app, router, _ := NewApiTest()
		SearchPhotos(router)
		result := PerformRequest(app, "GET", "/api/v1/photos?xxx=10")
		assert.Equal(t, http.StatusBadRequest, result.Code)
	})
	t.Run("PrivateQualityQueryTo4", func(t *testing.T) {
		app, router, _ := NewApiTest()
		SearchPhotos(router)
		r := PerformRequest(app, "GET", "/api/v1/photos?count=156&offset=0&merged=true&country=&camera=0&lens=0&label=&latlng=&year=0&month=0&color=&order=added&reverse=false&q=quality:4&public=&quality=3&private=true")
		body := r.Body.String()

		count := gjson.Get(body, "#")
		assert.LessOrEqual(t, int64(1), count.Int())
		assert.Equal(t, http.StatusOK, r.Code)
		qualities := gjson.Get(body, "#.Quality")
		assert.GreaterOrEqual(t, len(qualities.Array()), 1)
		for _, quality := range qualities.Array() {
			assert.GreaterOrEqual(t, quality.Int(), int64(4))
		}
	})
}

func TestSearchPhotosView(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, router, _ := NewApiTest()
		SearchPhotosView(router)
		r := PerformRequest(app, "GET", "/api/v1/photos/view?count=10")
		body := r.Body.String()

		t.Logf("response body: %s", body)

		assert.Equal(t, http.StatusOK, r.Code)
		count := gjson.Get(body, "#")
		assert.LessOrEqual(t, int64(2), count.Int())

		// Viewer-formatted results expose resolved download/preview URLs
		// that the default search endpoint does not populate.
		first := gjson.Get(body, "0")
		assert.True(t, first.Exists(), "expected at least one result")
		assert.NotEmpty(t, first.Get("DownloadUrl").String(), "DownloadUrl should be set for viewer results")
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		app, router, _ := NewApiTest()
		SearchPhotosView(router)
		result := PerformRequest(app, "GET", "/api/v1/photos/view?xxx=10")
		assert.Equal(t, http.StatusBadRequest, result.Code)
	})
}
