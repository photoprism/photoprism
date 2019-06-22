package api

import (
	"net/http"
	"testing"
	"encoding/json"
	"io/ioutil"
	"fmt"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/models"
)

func TestGetPhotos(t *testing.T) {
	app, router, conf := NewApiTest()
	conf.InitializeTestData(t)

	GetPhotos(router, conf)

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
	app, router, conf := NewApiTest()
	// conf.InitializeTestData(t)

	LikePhoto(router, conf)
	
	t.Run("Like Existing record", func(t *testing.T) {
		var photoFirst models.Photo
		photoFirst.TakenAt = time.Date(2019, time.June, 6, 21, 0, 0, 0, time.UTC)
		conf.Db().FirstOrCreate(&photoFirst)

		result := PerformRequest(app, "POST", fmt.Sprintf("/api/v1/photos/%d/like", photoFirst.ID))
		assert.Equal(t, http.StatusOK, result.Code)
		conf.Db().Delete(&photoFirst)
	})

	t.Run("Like missing record", func(t *testing.T) {
		var photoLast models.Photo
		conf.Db().Last(&photoLast)
		result := PerformRequest(app, "POST", fmt.Sprintf("/api/v1/photos/%d/like", photoLast.ID + 1))
		assert.Equal(t, http.StatusNotFound, result.Code)
	})

}

func TestDislikePhoto(t *testing.T) {
	app, router, conf := NewApiTest()
	conf.InitializeTestData(t)

	DislikePhoto(router, conf)

	t.Run("Dislike Existing record", func(t *testing.T) {
		var photoFirst models.Photo
		conf.Db().FirstOrCreate(&photoFirst)

		result := PerformRequest(app, "DELETE", fmt.Sprintf("/api/v1/photos/%d/like", photoFirst.ID))
		assert.Equal(t, http.StatusOK, result.Code)
	})

	t.Run("Dislike missing record", func(t *testing.T) {
		var photoLast models.Photo
		conf.Db().Last(&photoLast)
		result := PerformRequest(app, "DELETE", fmt.Sprintf("/api/v1/photos/%d/like", photoLast.ID + 1))
		assert.Equal(t, http.StatusNotFound, result.Code)
	})
}
