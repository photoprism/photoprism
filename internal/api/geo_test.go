package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetGeo(t *testing.T) {
	t.Run("get geo", func(t *testing.T) {
		app, router, _ := NewApiTest()

		GetGeo(router)

		result := PerformRequest(app, "GET", "/api/v1/geo")
		assert.Equal(t, http.StatusOK, result.Code)
	})
}
