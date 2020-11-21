package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func TestGetStatus(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		app, router, _ := NewApiTest()
		GetStatus(router)
		r := PerformRequest(app, "GET", "/api/v1/status")
		val := gjson.Get(r.Body.String(), "status")
		assert.Equal(t, "operational", val.String())
		assert.Equal(t, http.StatusOK, r.Code)
	})
}
