package api

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestWebsocket(t *testing.T) {
	t.Run("bad request", func(t *testing.T) {
		app, router, _ := NewApiTest()
		Websocket(router)
		r := PerformRequest(app, "GET", "/api/v1/ws")
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})

	t.Run("router nil", func(t *testing.T) {
		app, _, _ := NewApiTest()
		Websocket(nil)
		r := PerformRequest(app, "GET", "/api/v1/ws")
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
}
