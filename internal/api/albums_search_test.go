package api

import (
	"net/http"
	"testing"

	"github.com/tidwall/gjson"

	"github.com/stretchr/testify/assert"
)

func TestSearchAlbums(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		app, router, _ := NewApiTest()
		SearchAlbums(router)
		r := PerformRequest(app, "GET", "/api/v1/albums?count=10")
		count := gjson.Get(r.Body.String(), "#")
		assert.LessOrEqual(t, int64(3), count.Int())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("BadRequest", func(t *testing.T) {
		app, router, _ := NewApiTest()
		SearchAlbums(router)
		r := PerformRequest(app, "GET", "/api/v1/albums?xxx=10")
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
}
