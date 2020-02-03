package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPhotos(t *testing.T) {
	// TODO assert for json response
	t.Run("successful request", func(t *testing.T) {
		app, router, ctx := NewApiTest()

		GetPhotos(router, ctx)
		result := PerformRequest(app, "GET", "/api/v1/photos?count=10")
		assert.Equal(t, http.StatusOK, result.Code)
	})

	t.Run("invalid request", func(t *testing.T) {
		app, router, ctx := NewApiTest()
		GetPhotos(router, ctx)
		result := PerformRequest(app, "GET", "/api/v1/photos?xxx=10")
		assert.Equal(t, http.StatusBadRequest, result.Code)
	})
}
