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

	var photoTest models.Photo
	photoTest.TakenAt = time.Date(2019, time.June, 6, 21, 0, 0, 0, time.UTC) // TakenAt required as SQL complains for default value 0000-00-00
	conf.Db().Create(&photoTest)

	GetPhotos(router, conf)

	result := PerformRequest(app, "GET", "/api/v1/photos?count=10")
	assert.Equal(t, http.StatusOK, result.Code)
	
	// Test if the response body is a json matching the PhotoSearchResult struct
	var photoSearchRes []photoprism.PhotoSearchResult
	jsonResult, err := ioutil.ReadAll(result.Body)
	if err != nil {
		t.Error("error reading Body response")
	}
	if err = json.Unmarshal(jsonResult, &photoSearchRes); err != nil {
		t.Error("Json does correspond to required struct")
	}
	fmt.Printf("%+v \n", photoSearchRes)
	conf.Db().Delete(&photoTest)
}

func TestLikePhoto(t *testing.T) {
	app, router, conf := NewApiTest()

	LikePhoto(router, conf)
	
	var photoTest models.Photo
	photoTest.TakenAt = time.Date(2019, time.June, 6, 21, 0, 0, 0, time.UTC) // TakenAt required as SQL complains for default value 0000-00-00
	conf.Db().Create(&photoTest)

	t.Run("Like Existing record", func(t *testing.T) {

		result := PerformRequest(app, "POST", fmt.Sprintf("/api/v1/photos/%d/like", photoTest.ID))
		assert.Equal(t, http.StatusOK, result.Code)

		conf.Db().Take(&photoTest)
		assert.True(t, photoTest.PhotoFavorite)
		
	})
	
	t.Run("Like missing record", func(t *testing.T) {
		var photoLast models.Photo
		conf.Db().Last(&photoLast)

		result := PerformRequest(app, "POST", fmt.Sprintf("/api/v1/photos/%d/like", photoLast.ID + 1))
		assert.Equal(t, http.StatusNotFound, result.Code)
	})
	
	conf.Db().Delete(&photoTest)
}

func TestDislikePhoto(t *testing.T) {
	app, router, conf := NewApiTest()
	
	DislikePhoto(router, conf)
	
	var photoTest models.Photo
	photoTest.TakenAt = time.Date(2019, time.June, 6, 21, 0, 0, 0, time.UTC) // TakenAt required as SQL complains for default value 0000-00-00
	photoTest.PhotoFavorite = true
	conf.Db().Create(&photoTest)

	t.Run("Dislike Existing record", func(t *testing.T) {

		result := PerformRequest(app, "DELETE", fmt.Sprintf("/api/v1/photos/%d/like", photoTest.ID))
		assert.Equal(t, http.StatusOK, result.Code)

		conf.Db().Take(&photoTest)
		assert.False(t, photoTest.PhotoFavorite)
	})
	
	t.Run("Dislike missing record", func(t *testing.T) {
		var photoLast models.Photo
		conf.Db().Last(&photoLast)
		
		result := PerformRequest(app, "DELETE", fmt.Sprintf("/api/v1/photos/%d/like", photoLast.ID + 1))
		
		assert.Equal(t, http.StatusNotFound, result.Code)
	})
	
	conf.Db().Delete(&photoTest)
}
