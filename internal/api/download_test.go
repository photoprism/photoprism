package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDownload(t *testing.T) {
	t.Run("could not find original", func(t *testing.T) {
		app, router, conf := NewApiTest()

		GetDownload(router, conf)

		result := PerformRequest(app, "GET", "/api/v1/download/123xxx")
		assert.Equal(t, http.StatusNotFound, result.Code)
	})
	t.Run("download existing not existing file", func(t *testing.T) {
		app, router, conf := NewApiTest()

		DownloadAlbum(router, conf)

		result := PerformRequest(app, "GET", "/api/v1/download/555")
		assert.Equal(t, http.StatusNotFound, result.Code)
	})
}
