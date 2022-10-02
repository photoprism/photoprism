package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSearchGeo(t *testing.T) {
	t.Run("GeoJSON", func(t *testing.T) {
		app, router, _ := NewApiTest()

		SearchGeo(router)

		result := PerformRequest(app, "GET", "/api/v1/geo")
		assert.Equal(t, http.StatusOK, result.Code)
	})
	t.Run("ViewerJSON", func(t *testing.T) {
		app, router, _ := NewApiTest()

		SearchGeo(router)

		r := PerformRequest(app, "GET", "/api/v1/geo/view")

		assert.Equal(t, http.StatusOK, r.Code)
		t.Logf("response: %s", r.Body.String())
	})
}
