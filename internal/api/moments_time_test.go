package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetMomentsTime(t *testing.T) {
	t.Run("get geo", func(t *testing.T) {
		app, router, conf := NewApiTest()

		GetMomentsTime(router, conf)

		result := PerformRequest(app, "GET", "/api/v1/moments/time")
		assert.Equal(t, http.StatusOK, result.Code)
	})
}
