package api

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config/customize"
)

func TestAlbumDownloadName(t *testing.T) {
	t.Run("File", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/api/v1/albums?name=file", nil)
		assert.NoError(t, err)

		c := &gin.Context{
			Request: req,
		}

		assert.Equal(t, customize.DownloadNameFile, AlbumDownloadName(c))
	})
	t.Run("Share", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/api/v1/albums?name=share", nil)
		assert.NoError(t, err)

		c := &gin.Context{
			Request: req,
		}

		assert.Equal(t, customize.DownloadNameShare, AlbumDownloadName(c))
	})
	t.Run("Original", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/api/v1/albums?name=original", nil)
		assert.NoError(t, err)

		c := &gin.Context{
			Request: req,
		}

		assert.Equal(t, customize.DownloadNameOriginal, AlbumDownloadName(c))
	})
}

func TestDownloadAlbum(t *testing.T) {
	t.Run("DownloadNotExistingAlbum", func(t *testing.T) {
		app, router, conf := NewApiTest()

		DownloadAlbum(router)

		r := PerformRequest(app, "GET", "/api/v1/albums/5678/dl?t="+conf.DownloadToken())
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("DownloadExistingAlbum", func(t *testing.T) {
		app, router, conf := NewApiTest()

		DownloadAlbum(router)

		r := PerformRequest(app, "GET", "/api/v1/albums/as6sg6bxpogaaba8/dl?t="+conf.DownloadToken())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("DownloadDisabled", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.Settings().Features.Download = false

		DownloadAlbum(router)

		r := PerformRequest(app, "GET", "/api/v1/albums/as6sg6bxpogaaba8/dl?t="+conf.DownloadToken())
		assert.Equal(t, http.StatusForbidden, r.Code)

		conf.Settings().Features.Download = true
	})
}
