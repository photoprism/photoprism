package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/photoprism/photoprism/internal/i18n"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func TestBatchPhotosArchive(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		app, router, _ := NewApiTest()
		GetPhoto(router)
		r := PerformRequest(app, "GET", "/api/v1/photos/pt9jtdre2lvl0yh7")
		assert.Equal(t, http.StatusOK, r.Code)
		val := gjson.Get(r.Body.String(), "DeletedAt")
		assert.Empty(t, val.String())

		BatchPhotosArchive(router)
		r2 := PerformRequestWithBody(app, "POST", "/api/v1/batch/photos/archive", `{"photos": ["pt9jtdre2lvl0yh7", "pt9jtdre2lvl0ycc"]}`)
		val2 := gjson.Get(r2.Body.String(), "message")
		assert.Contains(t, val2.String(), "Selection archived")
		assert.Equal(t, http.StatusOK, r2.Code)

		r3 := PerformRequest(app, "GET", "/api/v1/photos/pt9jtdre2lvl0yh7")
		assert.Equal(t, http.StatusOK, r3.Code)
		val3 := gjson.Get(r3.Body.String(), "DeletedAt")
		assert.NotEmpty(t, val3.String())
	})
	t.Run("no items selected", func(t *testing.T) {
		app, router, _ := NewApiTest()
		BatchPhotosArchive(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/batch/photos/archive", `{"photos": []}`)
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, i18n.Msg(i18n.ErrNoItemsSelected), val.String())
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("invalid request", func(t *testing.T) {
		app, router, _ := NewApiTest()
		BatchPhotosArchive(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/batch/photos/archive", `{"photos": 123}`)
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
}

func TestBatchPhotosRestore(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		app, router, _ := NewApiTest()

		// Register routes.
		BatchPhotosArchive(router)
		GetPhoto(router)
		BatchPhotosRestore(router)

		r2 := PerformRequestWithBody(app, "POST", "/api/v1/batch/photos/archive", `{"photos": ["pt9jtdre2lvl0yh8", "pt9jtdre2lvl0ycc"]}`)
		val2 := gjson.Get(r2.Body.String(), "message")
		assert.Contains(t, val2.String(), "Selection archived")
		assert.Equal(t, http.StatusOK, r2.Code)

		r3 := PerformRequest(app, "GET", "/api/v1/photos/pt9jtdre2lvl0yh8")
		assert.Equal(t, http.StatusOK, r3.Code)
		val3 := gjson.Get(r3.Body.String(), "DeletedAt")
		assert.NotEmpty(t, val3.String())

		r := PerformRequestWithBody(app, "POST", "/api/v1/batch/photos/restore", `{"photos": ["pt9jtdre2lvl0yh8", "pt9jtdre2lvl0ycc"]}`)
		val := gjson.Get(r.Body.String(), "message")
		assert.Contains(t, val.String(), "Selection restored")
		assert.Equal(t, http.StatusOK, r.Code)

		r4 := PerformRequest(app, "GET", "/api/v1/photos/pt9jtdre2lvl0yh8")
		assert.Equal(t, http.StatusOK, r4.Code)
		val4 := gjson.Get(r4.Body.String(), "DeletedAt")
		assert.Empty(t, val4.String())
	})
	t.Run("no items selected", func(t *testing.T) {
		app, router, _ := NewApiTest()
		BatchPhotosRestore(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/batch/photos/restore", `{"photos": []}`)
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, i18n.Msg(i18n.ErrNoItemsSelected), val.String())
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("invalid request", func(t *testing.T) {
		app, router, _ := NewApiTest()
		BatchPhotosRestore(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/batch/photos/restore", `{"photos": 123}`)
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
}

func TestBatchAlbumsDelete(t *testing.T) {
	app, router, _ := NewApiTest()
	CreateAlbum(router)
	r := PerformRequestWithBody(app, "POST", "/api/v1/albums", `{"Title": "BatchDelete", "Description": "To be deleted", "Notes": "", "Favorite": true}`)
	assert.Equal(t, http.StatusOK, r.Code)
	uid := gjson.Get(r.Body.String(), "UID").String()

	t.Run("successful request", func(t *testing.T) {
		app, router, _ := NewApiTest()

		// Register routes.
		GetAlbum(router)
		BatchAlbumsDelete(router)

		r := PerformRequest(app, "GET", "/api/v1/albums/"+uid)
		val := gjson.Get(r.Body.String(), "Slug")
		assert.Equal(t, "batchdelete", val.String())

		r2 := PerformRequestWithBody(app, "POST", "/api/v1/batch/albums/delete", fmt.Sprintf(`{"albums": ["%s", "pt9jtdre2lvl0ycc"]}`, uid))
		val2 := gjson.Get(r2.Body.String(), "message")
		assert.Contains(t, val2.String(), i18n.Msg(i18n.MsgAlbumsDeleted))
		assert.Equal(t, http.StatusOK, r2.Code)

		r3 := PerformRequest(app, "GET", "/api/v1/albums/"+uid)
		val3 := gjson.Get(r3.Body.String(), "error")
		assert.Equal(t, i18n.Msg(i18n.ErrAlbumNotFound), val3.String())
		assert.Equal(t, http.StatusNotFound, r3.Code)
	})
	t.Run("no albums selected", func(t *testing.T) {
		app, router, _ := NewApiTest()
		BatchAlbumsDelete(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/batch/albums/delete", `{"albums": []}`)
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, i18n.Msg(i18n.ErrNoAlbumsSelected), val.String())
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("invalid request", func(t *testing.T) {
		app, router, _ := NewApiTest()
		BatchAlbumsDelete(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/batch/albums/delete", `{"albums": 123}`)
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
}

func TestBatchPhotosPrivate(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		app, router, _ := NewApiTest()

		// Register routes.
		GetPhoto(router)
		BatchPhotosPrivate(router)

		r := PerformRequest(app, "GET", "/api/v1/photos/pt9jtdre2lvl0yh8")
		assert.Equal(t, http.StatusOK, r.Code)
		val := gjson.Get(r.Body.String(), "Private")
		assert.Equal(t, "false", val.String())

		r2 := PerformRequestWithBody(app, "POST", "/api/v1/batch/photos/private", `{"photos": ["pt9jtdre2lvl0yh8", "pt9jtdre2lvl0ycc"]}`)
		val2 := gjson.Get(r2.Body.String(), "message")
		assert.Contains(t, val2.String(), "Selection marked as private")
		assert.Equal(t, http.StatusOK, r2.Code)

		r3 := PerformRequest(app, "GET", "/api/v1/photos/pt9jtdre2lvl0yh8")
		assert.Equal(t, http.StatusOK, r3.Code)
		val3 := gjson.Get(r3.Body.String(), "Private")
		assert.Equal(t, "true", val3.String())
	})
	t.Run("no items selected", func(t *testing.T) {
		app, router, _ := NewApiTest()
		BatchPhotosPrivate(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/batch/photos/private", `{"photos": []}`)
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, i18n.Msg(i18n.ErrNoItemsSelected), val.String())
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("invalid request", func(t *testing.T) {
		app, router, _ := NewApiTest()
		BatchPhotosPrivate(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/batch/photos/private", `{"photos": 123}`)
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
}

func TestBatchLabelsDelete(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		app, router, _ := NewApiTest()

		// Register routes.
		SearchLabels(router)
		BatchLabelsDelete(router)

		r := PerformRequest(app, "GET", "/api/v1/labels?count=15")
		val := gjson.Get(r.Body.String(), `#(Name=="Batch Delete").Slug`)
		assert.Equal(t, val.String(), "batch-delete")

		r2 := PerformRequestWithBody(app, "POST", "/api/v1/batch/labels/delete", fmt.Sprintf(`{"labels": ["lt9k3pw1wowuy3c6", "pt9jtdre2lvl0ycc"]}`))

		var resp i18n.Response

		if err := json.Unmarshal(r2.Body.Bytes(), &resp); err != nil {
			t.Fatal(err)
		}

		assert.True(t, resp.Success())
		assert.Equal(t, i18n.Msg(i18n.MsgLabelsDeleted), resp.Msg)
		assert.Equal(t, i18n.Msg(i18n.MsgLabelsDeleted), resp.String())
		assert.Equal(t, http.StatusOK, r2.Code)
		assert.Equal(t, http.StatusOK, resp.Code)

		r3 := PerformRequest(app, "GET", "/api/v1/labels?count=15")
		val3 := gjson.Get(r3.Body.String(), `#(Name=="BatchDelete").Slug`)
		assert.Equal(t, val3.String(), "")
	})
	t.Run("no labels selected", func(t *testing.T) {
		app, router, _ := NewApiTest()
		BatchLabelsDelete(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/batch/labels/delete", `{"labels": []}`)
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, i18n.Msg(i18n.ErrNoLabelsSelected), val.String())
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("invalid request", func(t *testing.T) {
		app, router, _ := NewApiTest()
		BatchLabelsDelete(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/batch/labels/delete", `{"labels": 123}`)
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
}

func TestBatchPhotosApprove(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		app, router, _ := NewApiTest()

		// Register routes.
		GetPhoto(router)
		BatchPhotosApprove(router)

		r := PerformRequest(app, "GET", "/api/v1/photos/pt9jtdre2lvl0y50")
		assert.Equal(t, http.StatusOK, r.Code)
		val := gjson.Get(r.Body.String(), "Quality")
		assert.Equal(t, "1", val.String())
		val4 := gjson.Get(r.Body.String(), "EditedAt")
		assert.Empty(t, val4.String())

		r2 := PerformRequestWithBody(app, "POST", "/api/v1/batch/photos/approve", `{"photos": ["pt9jtdre2lvl0y50", "pt9jtdre2lvl0y90"]}`)
		val2 := gjson.Get(r2.Body.String(), "message")
		assert.Contains(t, val2.String(), "Selection approved")
		assert.Equal(t, http.StatusOK, r2.Code)

		r3 := PerformRequest(app, "GET", "/api/v1/photos/pt9jtdre2lvl0y50")
		assert.Equal(t, http.StatusOK, r3.Code)
		val5 := gjson.Get(r3.Body.String(), "Quality")
		assert.Equal(t, "7", val5.String())
		val6 := gjson.Get(r3.Body.String(), "EditedAt")
		assert.NotEmpty(t, val6.String())
	})
	t.Run("no items selected", func(t *testing.T) {
		app, router, _ := NewApiTest()
		BatchPhotosApprove(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/batch/photos/approve", `{"photos": []}`)
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, i18n.Msg(i18n.ErrNoItemsSelected), val.String())
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("invalid request", func(t *testing.T) {
		app, router, _ := NewApiTest()
		BatchPhotosApprove(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/batch/photos/approve", `{"photos": 123}`)
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
}

func TestBatchPhotosDelete(t *testing.T) {
	t.Run("ErrNoItemsSelected", func(t *testing.T) {
		app, router, _ := NewApiTest()
		BatchPhotosDelete(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/batch/photos/delete", `{"photos": []}`)
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, i18n.Msg(i18n.ErrNoItemsSelected), val.String())
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
}
