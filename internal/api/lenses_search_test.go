package api

import (
	"net/http"
	"testing"

	"github.com/tidwall/gjson"

	"github.com/stretchr/testify/assert"
)

func TestSearchLenses(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, router, _ := NewApiTest()
		SearchLenses(router)
		r := PerformRequest(app, "GET", "/api/v1/lenses?count=15")
		count := gjson.Get(r.Body.String(), "#")
		assert.LessOrEqual(t, int64(3), count.Int())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		app, router, _ := NewApiTest()
		SearchLenses(router)
		r := PerformRequest(app, "GET", "/api/v1/lenses?xxx=15")
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
}
