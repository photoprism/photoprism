package api

import (
	"github.com/tidwall/gjson"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFile(t *testing.T) {
	t.Run("search for existing file", func(t *testing.T) {
		app, router, ctx := NewApiTest()
		GetFile(router, ctx)
		r := PerformRequest(app, "GET", "/api/v1/files/2cad9168fa6acc5c5c2965ddf6ec465ca42fd818")
		assert.Equal(t, http.StatusOK, r.Code)

		val := gjson.Get(r.Body.String(), "FileName")
		assert.Equal(t, "exampleFileName.jpg", val.String())
	})
	t.Run("search for not existing file", func(t *testing.T) {
		app, router, ctx := NewApiTest()
		GetFile(router, ctx)
		r := PerformRequest(app, "GET", "/api/v1/files/111")
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
}

func TestLinkFile(t *testing.T) {
	t.Run("album not found", func(t *testing.T) {
		app, router, ctx := NewApiTest()
		LinkFile(router, ctx)
		r := PerformRequest(app, "POST", "/api/v1/files/3cad9168fa6acc5c5c2965ddf6ec465ca42fd818/link")
		t.Log(r.Body.String())
		assert.Equal(t, http.StatusNotFound, r.Code)
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, "Album not found", val.String())
	})
	t.Run("invalid request", func(t *testing.T) {
		app, router, ctx := NewApiTest()
		LinkFile(router, ctx)
		r := PerformRequest(app, "POST", "/api/v1/files/ft9es39w45bnlqdw/link")
		assert.Equal(t, http.StatusBadRequest, r.Code)
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, "Invalid request", val.String())
	})

}
