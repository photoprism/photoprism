package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConnect(t *testing.T) {
	t.Run("NoNameOrToken", func(t *testing.T) {
		app, router, _ := NewApiTest()
		Connect(router)
		r := PerformRequest(app, "PUT", "/api/v1/connect")
		assert.Equal(t, 404, r.Code)
	})
	t.Run("NoName", func(t *testing.T) {
		app, router, _ := NewApiTest()
		Connect(router)
		r := PerformRequest(app, "PUT", "/api/v1/connect/")
		assert.Equal(t, 404, r.Code)
	})
	t.Run("NoToken", func(t *testing.T) {
		app, router, _ := NewApiTest()
		Connect(router)
		r := PerformRequest(app, "PUT", "/api/v1/connect/hub/")
		assert.Equal(t, 307, r.Code)
	})
	t.Run("InvalidUrl", func(t *testing.T) {
		app, router, _ := NewApiTest()
		Connect(router)
		r := PerformRequest(app, "PUT", "/api/v1/connect/hub/a")
		assert.Equal(t, 404, r.Code)
	})
	t.Run("Redirect", func(t *testing.T) {
		app, router, _ := NewApiTest()
		Connect(router)
		r := PerformRequest(app, "PUT", "/api/v1/connect/hub/foobar123")
		assert.NotEqual(t, 301, r.Code)
	})
	t.Run("HasNameAndToken", func(t *testing.T) {
		app, router, _ := NewApiTest()
		Connect(router)
		r := PerformRequestWithBody(app, "PUT", "/api/v1/connect/hub/", `{"Token": "foobar123"}`)
		assert.NotEqual(t, 200, r.Code)
	})
}
