package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func TestZip(t *testing.T) {
	app, router, conf := NewApiTest()
	ZipCreate(router)
	ZipDownload(router)

	t.Run("Download", func(t *testing.T) {
		r := PerformRequestWithBody(app, "POST", "/api/v1/zip", `{"photos": ["ps6sg6be2lvl0y12", "ps6sg6be2lvl0y11"]}`)
		message := gjson.Get(r.Body.String(), "message")
		assert.Contains(t, message.String(), "Zip created")
		assert.Equal(t, http.StatusOK, r.Code)
		filename := gjson.Get(r.Body.String(), "filename")
		dl := PerformRequest(app, "GET", "/api/v1/zip/"+filename.String()+"?t="+conf.DownloadToken())
		assert.Equal(t, http.StatusOK, dl.Code)
	})
	t.Run("ErrNoItemsSelected", func(t *testing.T) {
		r := PerformRequestWithBody(app, "POST", "/api/v1/zip", `{"photos": []}`)
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, "No items selected", val.String())
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("ErrBadRequest", func(t *testing.T) {
		r := PerformRequestWithBody(app, "POST", "/api/v1/zip", `{"photos": [123, "ps6sg6be2lvl0yxx"]}`)
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("ErrNotFound", func(t *testing.T) {
		r := PerformRequest(app, "GET", "/api/v1/zip/xxx?t="+conf.DownloadToken())
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
}
