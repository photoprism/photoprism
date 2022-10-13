package api

import (
	"net/http"
	"testing"

	"github.com/tidwall/gjson"

	"github.com/stretchr/testify/assert"
)

func TestSearchSubjects(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		app, router, _ := NewApiTest()
		SearchSubjects(router)
		r := PerformRequest(app, "GET", "/api/v1/subjects?count=10")
		count := gjson.Get(r.Body.String(), "#")
		assert.LessOrEqual(t, int64(3), count.Int())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		app, router, _ := NewApiTest()
		SearchSubjects(router)
		r := PerformRequest(app, "GET", "/api/v1/subjects?xxx=10")
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
}
