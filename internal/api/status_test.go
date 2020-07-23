package api

import (
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"net/http"
	"testing"
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
