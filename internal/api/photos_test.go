package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPhotos(t *testing.T) {
	app, router, conf := NewTest()

	GetPhotos(router, conf)

	result := TestRequest(app, "GET", "/api/v1/photos?count=10")

	assert.Equal(t, http.StatusOK, result.Code)
}

func TestLikePhoto(t *testing.T) {
	app, router, conf := NewTest()

	LikePhoto(router, conf)

	result := TestRequest(app, "POST", "/api/v1/photos/1/like")

	assert.Equal(t, http.StatusAccepted, result.Code)
}

func TestDislikePhoto(t *testing.T) {
	app, router, conf := NewTest()

	DislikePhoto(router, conf)

	result := TestRequest(app, "DELETE", "/api/v1/photos/1/like")

	assert.Equal(t, http.StatusAccepted, result.Code)
}
