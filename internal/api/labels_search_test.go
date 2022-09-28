package api

import (
	"net/http"
	"testing"

	"github.com/tidwall/gjson"

	"github.com/stretchr/testify/assert"
)

func TestSearchLabels(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		app, router, _ := NewApiTest()
		SearchLabels(router)
		r := PerformRequest(app, "GET", "/api/v1/labels?count=15")
		count := gjson.Get(r.Body.String(), "#")
		assert.LessOrEqual(t, int64(4), count.Int())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("invalid request", func(t *testing.T) {
		app, router, _ := NewApiTest()
		SearchLabels(router)
		r := PerformRequest(app, "GET", "/api/v1/labels?xxx=15")
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
}
