package api

import (
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"net/http"
	"testing"
)

func TestCreateZip(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		app, router, conf := NewApiTest()
		CreateZip(router, conf)
		r := PerformRequestWithBody(app, "POST", "/api/v1/zip", `{"photos": ["pt9jtdre2lvl0y12", "pt9jtdre2lvl0y11"]}`)
		val := gjson.Get(r.Body.String(), "message")
		assert.Contains(t, val.String(), "zip created")
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("no photos selected", func(t *testing.T) {
		app, router, conf := NewApiTest()
		CreateZip(router, conf)
		r := PerformRequestWithBody(app, "POST", "/api/v1/zip", `{"photos": []}`)
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, "No photos selected", val.String())
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("invalid request", func(t *testing.T) {
		app, router, conf := NewApiTest()
		CreateZip(router, conf)
		r := PerformRequestWithBody(app, "POST", "/api/v1/zip", `{"photos": [123, "pt9jtdre2lvl0yxx"]}`)
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
}

func TestDownloadZip(t *testing.T) {
	app, router, conf := NewApiTest()
	CreateZip(router, conf)
	r := PerformRequestWithBody(app, "POST", "/api/v1/zip", `{"photos": ["pt9jtdre2lvl0y12", "pt9jtdre2lvl0y11"]}`)
	filename := gjson.Get(r.Body.String(), "filename")
	assert.Equal(t, http.StatusOK, r.Code)

	t.Run("successful request", func(t *testing.T) {
		app, router, conf := NewApiTest()
		DownloadZip(router, conf)
		r := PerformRequest(app, "GET", "/api/v1/zip/"+filename.String())
		assert.Equal(t, http.StatusOK, r.Code)
	})

	t.Run("zip not existing", func(t *testing.T) {
		app, router, conf := NewApiTest()
		DownloadZip(router, conf)
		r := PerformRequest(app, "GET", "/api/v1/zip/xxx")
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
}
