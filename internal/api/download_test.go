package api

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/config/customize"
)

func TestDownloadName(t *testing.T) {
	t.Run("File", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "api/v1/dl?name=file", nil)
		assert.NoError(t, err)

		c := &gin.Context{
			Request: req,
		}

		assert.Equal(t, customize.DownloadNameFile, DownloadName(c))
	})
	t.Run("Share", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "api/v1/dl?name=share", nil)
		assert.NoError(t, err)

		c := &gin.Context{
			Request: req,
		}

		assert.Equal(t, customize.DownloadNameShare, DownloadName(c))
	})
	t.Run("Original", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "api/v1/dl?name=original", nil)
		assert.NoError(t, err)

		c := &gin.Context{
			Request: req,
		}

		assert.Equal(t, customize.DownloadNameOriginal, DownloadName(c))
	})
}

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
