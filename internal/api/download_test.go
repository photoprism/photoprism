package api

import (
	"github.com/tidwall/gjson"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDownload(t *testing.T) {
	t.Run("download not existing file", func(t *testing.T) {
		app, router, conf := NewApiTest()

		GetDownload(router, conf)

		r := PerformRequest(app, "GET", "/api/v1/download/123xxx")
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, "record not found", val.String())
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("could not find original", func(t *testing.T) {
		app, router, conf := NewApiTest()
		GetDownload(router, conf)
		r := PerformRequest(app, "GET", "/api/v1/download/3cad9168fa6acc5c5c2965ddf6ec465ca42fd818")
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
}
