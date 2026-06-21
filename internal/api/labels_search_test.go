package api

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/tidwall/gjson"

	"github.com/stretchr/testify/assert"
)

func TestSearchLabels(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, router, _ := NewApiTest()
		SearchLabels(router)
		r := PerformRequest(app, "GET", "/api/v1/labels?count=15")
		count := gjson.Get(r.Body.String(), "#")
		assert.LessOrEqual(t, int64(4), count.Int())
		assert.Equal(t, http.StatusOK, r.Code)
		result := r.Result()
		xCount, err := strconv.Atoi(result.Header.Get("X-Count"))
		assert.NoError(t, err, "strconv for X-Count failed")
		assert.LessOrEqual(t, 4, xCount)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		app, router, _ := NewApiTest()
		SearchLabels(router)
		r := PerformRequest(app, "GET", "/api/v1/labels?xxx=15")
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
}
