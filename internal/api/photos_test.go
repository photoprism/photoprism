package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPhotos(t *testing.T) {
	app, router, conf := NewApiTest()

	GetPhotos(router, conf)

	result := PerformRequest(app, "GET", "/api/v1/photos?count=10")

	assert.Equal(t, http.StatusOK, result.Code)
}

func TestLikePhoto(t *testing.T) {
	app, router, conf := NewApiTest()

	LikePhoto(router, conf)

	result := PerformRequest(app, "POST", "/api/v1/photos/1/like")

	assert.Equal(t, http.StatusOK, result.Code)
}

func TestDislikePhoto(t *testing.T) {
	app, router, conf := NewApiTest()

	DislikePhoto(router, conf)

	result := PerformRequest(app, "DELETE", "/api/v1/photos/1/like")

	assert.Equal(t, http.StatusOK, result.Code)
}
