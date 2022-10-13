package api

import (
	"net/http"
	"testing"

	"github.com/tidwall/gjson"

	"github.com/stretchr/testify/assert"
)

func TestSearchFaces(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		app, router, _ := NewApiTest()
		SearchFaces(router)
		r := PerformRequest(app, "GET", "/api/v1/faces?count=15")
		count := gjson.Get(r.Body.String(), "#")
		assert.LessOrEqual(t, int64(4), count.Int())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		app, router, _ := NewApiTest()
		SearchFaces(router)
		r := PerformRequest(app, "GET", "/api/v1/faces?xxx=15")
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
}
