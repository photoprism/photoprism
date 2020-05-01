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
		result := PerformRequest(app, "GET", "/api/v1/files/2cad9168fa6acc5c5c2965ddf6ec465ca42fd818")
		assert.Equal(t, http.StatusOK, result.Code)

		val := gjson.Get(result.Body.String(), "FileName")
		assert.Equal(t, "exampleFileName.jpg", val.String())
	})
	t.Run("search for not existing file", func(t *testing.T) {
		app, router, ctx := NewApiTest()
		GetFile(router, ctx)
		result := PerformRequest(app, "GET", "/api/v1/files/111")
		assert.Equal(t, http.StatusNotFound, result.Code)
	})
}
