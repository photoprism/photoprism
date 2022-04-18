package api

import (
	"net/http"
	"testing"

	"github.com/tidwall/gjson"

	"github.com/stretchr/testify/assert"
)

func TestGetDownload(t *testing.T) {
	t.Run("download not existing file", func(t *testing.T) {
		app, router, conf := NewApiTest()

		GetDownload(router)

		r := PerformRequest(app, "GET", "/api/v1/dl/123xxx?t="+conf.DownloadToken())
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, "record not found", val.String())
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("could not find original", func(t *testing.T) {
		app, router, conf := NewApiTest()
		GetDownload(router)
		r := PerformRequest(app, "GET", "/api/v1/dl/3cad9168fa6acc5c5c2965ddf6ec465ca42fd818?t="+conf.DownloadToken())
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("invalid download token", func(t *testing.T) {
		app, router, _ := NewApiTest()
		GetDownload(router)
		r := PerformRequest(app, "GET", "/api/v1/dl/3cad9168fa6acc5c5c2965ddf6ec465ca42fd818?t=xxx")
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
}
