package api

import (
	"net/http"
	"testing"

	"github.com/tidwall/gjson"

	"github.com/stretchr/testify/assert"
)

func TestGetFile(t *testing.T) {
	t.Run("search for existing file", func(t *testing.T) {
		app, router, _ := NewApiTest()
		GetFile(router)
		r := PerformRequest(app, "GET", "/api/v1/files/2cad9168fa6acc5c5c2965ddf6ec465ca42fd818")
		assert.Equal(t, http.StatusOK, r.Code)

		val := gjson.Get(r.Body.String(), "Name")
		assert.Equal(t, "2790/07/27900704_070228_D6D51B6C.jpg", val.String())
	})
	t.Run("search for not existing file", func(t *testing.T) {
		app, router, _ := NewApiTest()
		GetFile(router)
		r := PerformRequest(app, "GET", "/api/v1/files/111")
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
}
