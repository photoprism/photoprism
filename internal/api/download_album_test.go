package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
