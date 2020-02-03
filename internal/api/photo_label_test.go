package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemovePhotoLabel(t *testing.T) {
	t.Run("photo with label", func(t *testing.T) {
		app, router, ctx := NewApiTest()
		RemovePhotoLabel(router, ctx)
		result := PerformRequest(app, "DELETE", "/api/v1/photos/654/label/1")
		assert.Equal(t, http.StatusOK, result.Code)
		assert.NotContains(t, result.Body.String(), "\"LabelID\":1")
	})
	t.Run("try to remove wrong label", func(t *testing.T) {
		app, router, ctx := NewApiTest()
		RemovePhotoLabel(router, ctx)
		result := PerformRequest(app, "DELETE", "/api/v1/photos/654/label/3")
		assert.Equal(t, http.StatusOK, result.Code)
	})
	t.Run("not existing photo", func(t *testing.T) {
		app, router, ctx := NewApiTest()
		RemovePhotoLabel(router, ctx)
		result := PerformRequest(app, "DELETE", "/api/v1/photos/xx/label/")
		assert.Equal(t, http.StatusNotFound, result.Code)
	})
}
