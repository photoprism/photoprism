package api

import (
	"net/http"
	"testing"

	"github.com/tidwall/gjson"

	"github.com/stretchr/testify/assert"
)

func TestGetConfig(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		app, router, _ := NewApiTest()
		GetClientConfig(router)
		r := PerformRequest(app, "GET", "/api/v1/config")
		val := gjson.Get(r.Body.String(), "flags")
		assert.Equal(t, "public debug test sponsor experimental settings", val.String())
		assert.Equal(t, http.StatusOK, r.Code)
	})
}

func TestGetConfigOptions(t *testing.T) {
	t.Run("unauthorised", func(t *testing.T) {
		app, router, _ := NewApiTest()
		GetConfigOptions(router)
		r := PerformRequest(app, "GET", "/api/v1/config/options")
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
}

func TestSaveConfigOptions(t *testing.T) {
	t.Run("unauthorised", func(t *testing.T) {
		app, router, _ := NewApiTest()
		SaveConfigOptions(router)
		r := PerformRequest(app, "POST", "/api/v1/config/options")
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
}
