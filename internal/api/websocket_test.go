package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWebsocket(t *testing.T) {
	t.Run("BadRequest", func(t *testing.T) {
		app, router, _ := NewApiTest()
		WebSocket(router)
		r := PerformRequest(app, "GET", "/api/v1/ws")
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})

	t.Run("NoRouter", func(t *testing.T) {
		app, _, _ := NewApiTest()
		WebSocket(nil)
		r := PerformRequest(app, "GET", "/api/v1/ws")
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
}
