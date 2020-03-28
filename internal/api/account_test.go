package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAccounts(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		app, router, conf := NewApiTest()
		GetAccounts(router, conf)
		result := PerformRequest(app, "GET", "/api/v1/accounts?count=10")
		assert.Contains(t, result.Body.String(), "Test Account")
		assert.Equal(t, http.StatusOK, result.Code)
	})
	t.Run("invalid request", func(t *testing.T) {
		app, router, conf := NewApiTest()
		GetAccounts(router, conf)
		result := PerformRequest(app, "GET", "/api/v1/accounts?xxx=10")

		assert.Equal(t, http.StatusBadRequest, result.Code)
	})
}

func TestGetAccount(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		app, router, conf := NewApiTest()
		GetAccount(router, conf)
		result := PerformRequest(app, "GET", "/api/v1/accounts/1")
		assert.Contains(t, result.Body.String(), "Test Account")
		assert.Equal(t, http.StatusOK, result.Code)
	})
	t.Run("invalid request", func(t *testing.T) {
		app, router, conf := NewApiTest()
		GetAccount(router, conf)
		result := PerformRequest(app, "GET", "/api/v1/accounts/999000")
		assert.Equal(t, http.StatusNotFound, result.Code)
	})
}
