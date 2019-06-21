package api

import (
	"net/http"
	"testing"
	"encoding/json"
	"io/ioutil"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/models"
)

func TestGetPhotos(t *testing.T) {
	app, router, ctx := NewApiTest()

	GetPhotos(router, ctx)

	result := PerformRequest(app, "GET", "/api/v1/photos?count=10")

	assert.Equal(t, http.StatusOK, result.Code)

	var photoSearchRes []photoprism.PhotoSearchResult

	// Test if the response body is a json matching the PhotoSearchResult struct
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
	
	t.Run("Like Existing record", func(t *testing.T) {
		// Ensure that at least one record exist in the test database
		
		photoFirst := ctx.Db().FirstOrCreate(models.Photo)
		idFirst := photoFirst.ID
		t.Print(idfirst)
		result := PerformRequest(app, "POST", "/api/v1/photos/1/like")
		assert.Equal(t, http.StatusOK, result.Code)
	})

	t.Run("Like Non-existing record", func(t *testing.T) {
		ctx.Db().Delete(
			models.Photo{
				Model: models.Model{ID: 1},
			},
		)
		result := PerformRequest(app, "POST", "/api/v1/photos/1/like")
		assert.Equal(t, http.StatusNotFound, result.Code)
	})
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
