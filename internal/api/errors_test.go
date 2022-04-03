package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetErrors(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		app, router, _ := NewApiTest()
		GetErrors(router)
		r := PerformRequest(app, "GET", "/api/v1/errors")
		// Ok if no error is thrown.
		assert.Equal(t, http.StatusOK, r.Code)
	})
}

func TestDeleteErrors(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		app, router, _ := NewApiTest()
		DeleteErrors(router)
		r := PerformRequest(app, "DELETE", "/api/v1/errors")
		// Disabled in public mode, so error 403 is expected.
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
}
