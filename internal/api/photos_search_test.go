package api

import (
	"net/http"
	"testing"

	"github.com/tidwall/gjson"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestSearchPhotos_NulValues(t *testing.T) {
	app, router, _ := NewApiTest()
	SearchPhotos(router)

	// Values with a NUL byte or invalid UTF-8 are searched without them, and a value of only those matches nothing.
	for _, param := range []string{"q", "path", "folder", "name", "filename", "original", "title", "caption",
		"description", "hash", "keywords", "label", "category", "country", "state", "city", "face", "person",
		"subject", "people", "subjects", "album", "albums", "camera", "lens", "codec", "color", "id",
		"type", "s2", "olc", "mm", "iso", "f", "year"} {
		t.Run(param, func(t *testing.T) {
			r := PerformRequest(app, "GET", "/api/v1/photos?count=10&"+param+"=%00")
			assert.Equal(t, http.StatusOK, r.Code, r.Body.String())
			assert.Equal(t, "[]", r.Body.String())
			r = PerformRequest(app, "GET", "/api/v1/photos?count=10&"+param+"=%FF%FE")
			assert.Equal(t, http.StatusOK, r.Code, r.Body.String())
			assert.Equal(t, "[]", r.Body.String())
			r = PerformRequest(app, "GET", "/api/v1/photos?count=10&"+param+"=a%00b%FF")
			assert.Equal(t, http.StatusOK, r.Code, r.Body.String())
		})
	}

	t.Run("OnlySyntaxLeft", func(t *testing.T) {
		for _, query := range []string{"q=%00%20", "q=%FF%7C", "path=%FF%20", "path=%FF*", "label=%FF%20",
			"q=before:%FF", "q=after:%FF", "filter=%FF", "path=%FF/", "path=%FF/*", "label=%FF!"} {
			r := PerformRequest(app, "GET", "/api/v1/photos?count=10&"+query)
			assert.Equal(t, http.StatusOK, r.Code, query)
			assert.Equal(t, "[]", r.Body.String(), query)
		}
	})

	t.Run("Unfiltered", func(t *testing.T) {
		for _, query := range []string{"", "&q=", "&path=", "&q=path:"} {
			r := PerformRequest(app, "GET", "/api/v1/photos?count=10"+query)
			assert.Equal(t, http.StatusOK, r.Code, r.Body.String())
			assert.Equal(t, "10", r.Header().Get("X-Count"), query)
		}
	})
	t.Run("ValidCharactersLeft", func(t *testing.T) {
		expected := PerformRequest(app, "GET", "/api/v1/photos?count=10&path=2790%2F07")
		require.Equal(t, http.StatusOK, expected.Code)
		require.NotEqual(t, "[]", expected.Body.String())
		r := PerformRequest(app, "GET", "/api/v1/photos?count=10&path=2790%2F07%FF")
		assert.Equal(t, http.StatusOK, r.Code)
		assert.Equal(t, expected.Header().Get("X-Count"), r.Header().Get("X-Count"))
	})
}
