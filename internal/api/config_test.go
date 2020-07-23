package api

import (
	"github.com/tidwall/gjson"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetConfig(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		app, router, _ := NewApiTest()
		GetConfig(router)
		r := PerformRequest(app, "GET", "/api/v1/config")
		val := gjson.Get(r.Body.String(), "flags")
		assert.Equal(t, "public debug settings", val.String())
		assert.Equal(t, http.StatusOK, r.Code)
	})
}
