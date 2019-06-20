package api

import (
	"net/http"
	"testing"
	"encoding/json"

	"github.com/stretchr/testify/assert"
	"github.com/photoprism/photoprism/internal/photoprism"
)

func TestGetPhotos(t *testing.T) {
	app, router, ctx := NewApiTest()

	GetPhotos(router, ctx)

	result := PerformRequest(app, "GET", "/api/v1/photos?count=10")

	assert.Equal(t, http.StatusOK, result.Code)

	var photoSearchRes []photoprism.PhotoSearchResult

	jsonResult, err := ioutil.ReadAll(result.Body)
		if err != nil {
			t.Fail()
		}

	if err = json.Unmarshal(jsonResult, &photoSearchRes); err != nil {
		t.Fail()
	}
}

func TestLikePhoto(t *testing.T) {
	app, router, ctx := NewApiTest()

	LikePhoto(router, ctx)

	result := PerformRequest(app, "POST", "/api/v1/photos/1/like")

	// TODO: Test database can be empty
	if result.Code != http.StatusNotFound {
		assert.Equal(t, http.StatusOK, result.Code)
	}
}

func TestDislikePhoto(t *testing.T) {
	app, router, ctx := NewApiTest()

	DislikePhoto(router, ctx)

	result := PerformRequest(app, "DELETE", "/api/v1/photos/1/like")

	// TODO: Test database can be empty
	if result.Code != http.StatusNotFound {
		assert.Equal(t, http.StatusOK, result.Code)
	}
}
