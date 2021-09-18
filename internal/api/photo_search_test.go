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

	t.Run("InvalidRequest", func(t *testing.T) {
		app, router, _ := NewApiTest()
		SearchPhotos(router)
		result := PerformRequest(app, "GET", "/api/v1/photos?xxx=10")
		assert.Equal(t, http.StatusBadRequest, result.Code)
	})
}
