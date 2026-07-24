package api

import (
	"net/http"
	"testing"

	"github.com/tidwall/gjson"

	"github.com/stretchr/testify/assert"
)

func TestSearchAlbums(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, router, _ := NewApiTest()
		SearchAlbums(router)
		r := PerformRequest(app, "GET", "/api/v1/albums?count=10&uid=as6sg6bxpogaaba7")
		count := gjson.Get(r.Body.String(), "#")
		assert.LessOrEqual(t, int64(1), count.Int())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("BadRequest", func(t *testing.T) {
		app, router, _ := NewApiTest()
		SearchAlbums(router)
		r := PerformRequest(app, "GET", "/api/v1/albums?xxx=10")
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
}
