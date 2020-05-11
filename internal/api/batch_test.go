package api

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"net/http"
	"testing"
)

func TestBatchPhotosArchive(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		app, router, conf := NewApiTest()
		GetPhoto(router, conf)
		r := PerformRequest(app, "GET", "/api/v1/photos/pt9jtdre2lvl0yh7")
		assert.Equal(t, http.StatusOK, r.Code)
		val := gjson.Get(r.Body.String(), "DeletedAt")
		assert.Empty(t, val.String())

		BatchPhotosArchive(router, conf)
		r2 := PerformRequestWithBody(app, "POST", "/api/v1/batch/photos/archive", `{"photos": ["pt9jtdre2lvl0yh7", "pt9jtdre2lvl0ycc"]}`)
		val2 := gjson.Get(r2.Body.String(), "message")
		assert.Contains(t, val2.String(), "photos archived")
		assert.Equal(t, http.StatusOK, r2.Code)

		r3 := PerformRequest(app, "GET", "/api/v1/photos/pt9jtdre2lvl0yh7")
		assert.Equal(t, http.StatusOK, r3.Code)
		val3 := gjson.Get(r3.Body.String(), "DeletedAt")
		assert.NotEmpty(t, val3.String())
	})
	t.Run("no photos selected", func(t *testing.T) {
		app, router, conf := NewApiTest()
		BatchPhotosArchive(router, conf)
		r := PerformRequestWithBody(app, "POST", "/api/v1/batch/photos/archive", `{"photos": []}`)
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, "No photos selected", val.String())
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("invalid request", func(t *testing.T) {
		app, router, conf := NewApiTest()
		BatchPhotosArchive(router, conf)
		r := PerformRequestWithBody(app, "POST", "/api/v1/batch/photos/archive", `{"photos": 123}`)
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
}

func TestBatchPhotosRestore(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		app, router, conf := NewApiTest()

		BatchPhotosArchive(router, conf)
		r2 := PerformRequestWithBody(app, "POST", "/api/v1/batch/photos/archive", `{"photos": ["pt9jtdre2lvl0yh8", "pt9jtdre2lvl0ycc"]}`)
		val2 := gjson.Get(r2.Body.String(), "message")
		assert.Contains(t, val2.String(), "photos archived")
		assert.Equal(t, http.StatusOK, r2.Code)

		GetPhoto(router, conf)
		r3 := PerformRequest(app, "GET", "/api/v1/photos/pt9jtdre2lvl0yh8")
		assert.Equal(t, http.StatusOK, r3.Code)
		val3 := gjson.Get(r3.Body.String(), "DeletedAt")
		assert.NotEmpty(t, val3.String())

		BatchPhotosRestore(router, conf)
		r := PerformRequestWithBody(app, "POST", "/api/v1/batch/photos/restore", `{"photos": ["pt9jtdre2lvl0yh8", "pt9jtdre2lvl0ycc"]}`)
		val := gjson.Get(r.Body.String(), "message")
		assert.Contains(t, val.String(), "photos restored")
		assert.Equal(t, http.StatusOK, r.Code)

		r4 := PerformRequest(app, "GET", "/api/v1/photos/pt9jtdre2lvl0yh8")
		assert.Equal(t, http.StatusOK, r4.Code)
		val4 := gjson.Get(r4.Body.String(), "DeletedAt")
		assert.Empty(t, val4.String())
	})
	t.Run("no photos selected", func(t *testing.T) {
		app, router, conf := NewApiTest()
		BatchPhotosRestore(router, conf)
		r := PerformRequestWithBody(app, "POST", "/api/v1/batch/photos/restore", `{"photos": []}`)
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, "No photos selected", val.String())
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("invalid request", func(t *testing.T) {
		app, router, conf := NewApiTest()
		BatchPhotosRestore(router, conf)
		r := PerformRequestWithBody(app, "POST", "/api/v1/batch/photos/restore", `{"photos": 123}`)
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
}

func TestBatchAlbumsDelete(t *testing.T) {
	app, router, conf := NewApiTest()
	CreateAlbum(router, conf)
	r := PerformRequestWithBody(app, "POST", "/api/v1/albums", `{"AlbumName": "BatchDelete", "AlbumDescription": "To be deleted", "AlbumNotes": "", "AlbumFavorite": true}`)
	assert.Equal(t, http.StatusOK, r.Code)
	uuid := gjson.Get(r.Body.String(), "AlbumUUID").String()

	t.Run("successful request", func(t *testing.T) {
		app, router, conf := NewApiTest()

		GetAlbum(router, conf)
		r := PerformRequest(app, "GET", "/api/v1/albums/"+uuid)
		val := gjson.Get(r.Body.String(), "AlbumSlug")
		assert.Equal(t, "batchdelete", val.String())

		BatchAlbumsDelete(router, conf)
		r2 := PerformRequestWithBody(app, "POST", "/api/v1/batch/albums/delete", fmt.Sprintf(`{"albums": ["%s", "pt9jtdre2lvl0ycc"]}`, uuid))
		val2 := gjson.Get(r2.Body.String(), "message")
		assert.Contains(t, val2.String(), "albums deleted")
		assert.Equal(t, http.StatusOK, r2.Code)

		r3 := PerformRequest(app, "GET", "/api/v1/albums/"+uuid)
		val3 := gjson.Get(r3.Body.String(), "error")
		assert.Equal(t, "Album not found", val3.String())
		assert.Equal(t, http.StatusNotFound, r3.Code)
	})
	t.Run("no albums selected", func(t *testing.T) {
		app, router, conf := NewApiTest()
		BatchAlbumsDelete(router, conf)
		r := PerformRequestWithBody(app, "POST", "/api/v1/batch/albums/delete", `{"albums": []}`)
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, "No albums selected", val.String())
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("invalid request", func(t *testing.T) {
		app, router, conf := NewApiTest()
		BatchAlbumsDelete(router, conf)
		r := PerformRequestWithBody(app, "POST", "/api/v1/batch/albums/delete", `{"albums": 123}`)
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
}

func TestBatchPhotosPrivate(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		app, router, conf := NewApiTest()
		GetPhoto(router, conf)
		r := PerformRequest(app, "GET", "/api/v1/photos/pt9jtdre2lvl0yh8")
		assert.Equal(t, http.StatusOK, r.Code)
		val := gjson.Get(r.Body.String(), "PhotoPrivate")
		assert.Equal(t, "false", val.String())

		BatchPhotosPrivate(router, conf)
		r2 := PerformRequestWithBody(app, "POST", "/api/v1/batch/photos/private", `{"photos": ["pt9jtdre2lvl0yh8", "pt9jtdre2lvl0ycc"]}`)
		val2 := gjson.Get(r2.Body.String(), "message")
		assert.Contains(t, val2.String(), "photos marked as private")
		assert.Equal(t, http.StatusOK, r2.Code)

		r3 := PerformRequest(app, "GET", "/api/v1/photos/pt9jtdre2lvl0yh8")
		assert.Equal(t, http.StatusOK, r3.Code)
		val3 := gjson.Get(r3.Body.String(), "PhotoPrivate")
		assert.Equal(t, "true", val3.String())
	})
	t.Run("no photos selected", func(t *testing.T) {
		app, router, conf := NewApiTest()
		BatchPhotosPrivate(router, conf)
		r := PerformRequestWithBody(app, "POST", "/api/v1/batch/photos/private", `{"photos": []}`)
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, "No photos selected", val.String())
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("invalid request", func(t *testing.T) {
		app, router, conf := NewApiTest()
		BatchPhotosPrivate(router, conf)
		r := PerformRequestWithBody(app, "POST", "/api/v1/batch/photos/private", `{"photos": 123}`)
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
}

func TestBatchLabelsDelete(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		app, router, conf := NewApiTest()
		GetLabels(router, conf)
		r := PerformRequest(app, "GET", "/api/v1/labels?count=15")
		val := gjson.Get(r.Body.String(), `#(LabelName=="BatchDelete").LabelSlug`)
		assert.Equal(t, val.String(), "batchdelete")

		BatchLabelsDelete(router, conf)
		r2 := PerformRequestWithBody(app, "POST", "/api/v1/batch/labels/delete", fmt.Sprintf(`{"labels": ["lt9k3pw1wowuy3c6", "pt9jtdre2lvl0ycc"]}`))
		val2 := gjson.Get(r2.Body.String(), "message")
		assert.Contains(t, val2.String(), "labels deleted")
		assert.Equal(t, http.StatusOK, r2.Code)

		r3 := PerformRequest(app, "GET", "/api/v1/labels?count=15")
		val3 := gjson.Get(r3.Body.String(), `#(LabelName=="BatchDelete").LabelSlug`)
		assert.Equal(t, val3.String(), "")
	})
	t.Run("no labels selected", func(t *testing.T) {
		app, router, conf := NewApiTest()
		BatchLabelsDelete(router, conf)
		r := PerformRequestWithBody(app, "POST", "/api/v1/batch/labels/delete", `{"labels": []}`)
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, "No labels selected", val.String())
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("invalid request", func(t *testing.T) {
		app, router, conf := NewApiTest()
		BatchLabelsDelete(router, conf)
		r := PerformRequestWithBody(app, "POST", "/api/v1/batch/labels/delete", `{"labels": 123}`)
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
}
