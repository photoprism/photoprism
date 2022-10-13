package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"

	"github.com/photoprism/photoprism/internal/config"
)

func TestGetDownload(t *testing.T) {
	t.Run("NotFound", func(t *testing.T) {
		app, router, conf := NewApiTest()
		GetDownload(router)
		r := PerformRequest(app, "GET", "/api/v1/dl/123xxx?t="+conf.DownloadToken())
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, "record not found", val.String())
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("MissingOriginal", func(t *testing.T) {
		app, router, conf := NewApiTest()
		GetDownload(router)
		r := PerformRequest(app, "GET", "/api/v1/dl/3cad9168fa6acc5c5c2965ddf6ec465ca42fd818?t="+conf.DownloadToken())
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("InvalidDownloadToken", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		GetDownload(router)
		r := PerformRequest(app, "GET", "/api/v1/dl/3cad9168fa6acc5c5c2965ddf6ec465ca42fd818?t=xxx")
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
}
