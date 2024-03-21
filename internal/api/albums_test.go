package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"

	"github.com/photoprism/photoprism/pkg/i18n"
)

func TestGetAlbum(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		app, router, _ := NewApiTest()
		GetAlbum(router)
		r := PerformRequest(app, "GET", "/api/v1/albums/as6sg6bxpogaaba8")
		val := gjson.Get(r.Body.String(), "Slug")
		assert.Equal(t, "holiday-2030", val.String())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("invalid request", func(t *testing.T) {
		app, router, _ := NewApiTest()
		GetAlbum(router)
		r := PerformRequest(app, "GET", "/api/v1/albums/999000")
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, "Album not found", val.String())
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
}

func TestCreateAlbum(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		app, router, _ := NewApiTest()
		CreateAlbum(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/albums", `{"Title": "New created album", "Notes": "", "Favorite": true}`)
		val := gjson.Get(r.Body.String(), "Slug")
		assert.Equal(t, "new-created-album", val.String())
		val2 := gjson.Get(r.Body.String(), "Favorite")
		assert.Equal(t, "true", val2.String())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("invalid request", func(t *testing.T) {
		app, router, _ := NewApiTest()
		CreateAlbum(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/albums", `{"Title": 333, "Description": "Created via unit test", "Notes": "", "Favorite": true}`)
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
}
func TestUpdateAlbum(t *testing.T) {
	app, router, _ := NewApiTest()
	CreateAlbum(router)
	r := PerformRequestWithBody(app, "POST", "/api/v1/albums", `{"Title": "Update", "Description": "To be updated", "Notes": "", "Favorite": true}`)
	assert.Equal(t, http.StatusOK, r.Code)
	uid := gjson.Get(r.Body.String(), "UID").String()

	t.Run("successful request", func(t *testing.T) {
		app, router, _ := NewApiTest()
		UpdateAlbum(router)
		r := PerformRequestWithBody(app, "PUT", "/api/v1/albums/"+uid, `{"Title": "Updated01", "Notes": "", "Favorite": false}`)
		val := gjson.Get(r.Body.String(), "Slug")
		assert.Equal(t, "updated01", val.String())
		val2 := gjson.Get(r.Body.String(), "Favorite")
		assert.Equal(t, "false", val2.String())
		assert.Equal(t, http.StatusOK, r.Code)
	})

	t.Run("invalid request", func(t *testing.T) {
		app, router, _ := NewApiTest()
		UpdateAlbum(router)
		r := PerformRequestWithBody(app, "PUT", "/api/v1/albums"+uid, `{"Title": 333, "Description": "Created via unit test", "Notes": "", "Favorite": true}`)
		assert.Equal(t, http.StatusNotFound, r.Code)
	})

	t.Run("not found", func(t *testing.T) {
		app, router, _ := NewApiTest()
		UpdateAlbum(router)
		r := PerformRequestWithBody(app, "PUT", "/api/v1/albums/xxx", `{"Title": "Update03", "Description": "Created via unit test", "Notes": "", "Favorite": true}`)
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, "Album not found", val.String())
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
}
func TestDeleteAlbum(t *testing.T) {
	app, router, _ := NewApiTest()
	CreateAlbum(router)
	r := PerformRequestWithBody(app, "POST", "/api/v1/albums", `{"Title": "Delete", "Description": "To be deleted", "Notes": "", "Favorite": true}`)
	assert.Equal(t, http.StatusOK, r.Code)
	uid := gjson.Get(r.Body.String(), "UID").String()

	t.Run("delete existing album", func(t *testing.T) {
		app, router, _ := NewApiTest()
		DeleteAlbum(router)
		r := PerformRequest(app, "DELETE", "/api/v1/albums/"+uid)
		assert.Equal(t, http.StatusOK, r.Code)
		val := gjson.Get(r.Body.String(), "Slug")
		assert.Equal(t, "delete", val.String())
		SearchAlbums(router)
		r2 := PerformRequest(app, "GET", "/api/v1/albums/"+uid)
		assert.Equal(t, http.StatusNotFound, r2.Code)
	})
	t.Run("delete not existing album", func(t *testing.T) {
		app, router, _ := NewApiTest()
		DeleteAlbum(router)
		r := PerformRequest(app, "DELETE", "/api/v1/albums/999000")
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, "Album not found", val.String())
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
}

func TestLikeAlbum(t *testing.T) {
	t.Run("like not existing album", func(t *testing.T) {
		app, router, _ := NewApiTest()

		LikeAlbum(router)

		r := PerformRequest(app, "POST", "/api/v1/albums/xxx/like")
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("like existing album", func(t *testing.T) {
		app, router, _ := NewApiTest()

		LikeAlbum(router)
		r := PerformRequest(app, "POST", "/api/v1/albums/as6sg6bxpogaaba7/like")
		assert.Equal(t, http.StatusOK, r.Code)
		GetAlbum(router)
		r2 := PerformRequest(app, "GET", "/api/v1/albums/as6sg6bxpogaaba7")
		val := gjson.Get(r2.Body.String(), "Favorite")
		assert.Equal(t, "true", val.String())
	})
}

func TestDislikeAlbum(t *testing.T) {
	t.Run("dislike not existing album", func(t *testing.T) {
		app, router, _ := NewApiTest()

		DislikeAlbum(router)

		r := PerformRequest(app, "DELETE", "/api/v1/albums/5678/like")
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("dislike existing album", func(t *testing.T) {
		app, router, _ := NewApiTest()

		DislikeAlbum(router)

		r := PerformRequest(app, "DELETE", "/api/v1/albums/as6sg6bxpogaaba8/like")
		assert.Equal(t, http.StatusOK, r.Code)
		GetAlbum(router)
		r2 := PerformRequest(app, "GET", "/api/v1/albums/as6sg6bxpogaaba8")
		val := gjson.Get(r2.Body.String(), "Favorite")
		assert.Equal(t, "false", val.String())
	})
}

func TestAddPhotosToAlbum(t *testing.T) {
	app, router, _ := NewApiTest()
	CreateAlbum(router)
	r := PerformRequestWithBody(app, "POST", "/api/v1/albums", `{"Title": "Add photos", "Description": "", "Notes": "", "Favorite": true}`)
	assert.Equal(t, http.StatusOK, r.Code)
	uid := gjson.Get(r.Body.String(), "UID").String()

	t.Run("successful request", func(t *testing.T) {
		app, router, _ := NewApiTest()
		AddPhotosToAlbum(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/albums/"+uid+"/photos", `{"photos": ["ps6sg6be2lvl0y12", "ps6sg6be2lvl0y11"]}`)
		val := gjson.Get(r.Body.String(), "message")
		assert.Equal(t, i18n.Msg(i18n.MsgChangesSaved), val.String())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("add one photo to album", func(t *testing.T) {
		app, router, _ := NewApiTest()
		AddPhotosToAlbum(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/albums/"+uid+"/photos", `{"photos": ["ps6sg6be2lvl0y12"]}`)
		val := gjson.Get(r.Body.String(), "message")
		assert.Equal(t, i18n.Msg(i18n.MsgChangesSaved), val.String())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("invalid request", func(t *testing.T) {
		app, router, _ := NewApiTest()
		AddPhotosToAlbum(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/albums/"+uid+"/photos", `{"photos": [123, "ps6sg6be2lvl0yxx"]}`)
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("not found", func(t *testing.T) {
		app, router, _ := NewApiTest()
		AddPhotosToAlbum(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/albums/xxx/photos", `{"photos": ["ps6sg6be2lvl0yxx"]}`)
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
}

func TestRemovePhotosFromAlbum(t *testing.T) {
	app, router, _ := NewApiTest()

	// Register routes.
	CreateAlbum(router)
	AddPhotosToAlbum(router)

	r := PerformRequestWithBody(app, "POST", "/api/v1/albums", `{"Title": "Remove photos", "Description": "", "Notes": "", "Favorite": true}`)
	assert.Equal(t, http.StatusOK, r.Code)
	uid := gjson.Get(r.Body.String(), "UID").String()

	r2 := PerformRequestWithBody(app, "POST", "/api/v1/albums/"+uid+"/photos", `{"photos": ["ps6sg6be2lvl0y12", "ps6sg6be2lvl0y11"]}`)
	assert.Equal(t, http.StatusOK, r2.Code)

	t.Run("successful request", func(t *testing.T) {
		app, router, _ := NewApiTest()
		RemovePhotosFromAlbum(router)
		r := PerformRequestWithBody(app, "DELETE", "/api/v1/albums/"+uid+"/photos", `{"photos": ["ps6sg6be2lvl0y12", "ps6sg6be2lvl0y11"]}`)
		val := gjson.Get(r.Body.String(), "message")
		assert.Equal(t, i18n.Msg(i18n.MsgChangesSaved), val.String())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("no items selected", func(t *testing.T) {
		app, router, _ := NewApiTest()
		RemovePhotosFromAlbum(router)
		r := PerformRequestWithBody(app, "DELETE", "/api/v1/albums/as6sg6bxpogaaba7/photos", `{"photos": []}`)
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, "No items selected", val.String())
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("invalid request", func(t *testing.T) {
		app, router, _ := NewApiTest()
		RemovePhotosFromAlbum(router)
		r := PerformRequestWithBody(app, "DELETE", "/api/v1/albums/"+uid+"/photos", `{"photos": [123, "ps6sg6be2lvl0yxx"]}`)
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("album not found", func(t *testing.T) {
		app, router, _ := NewApiTest()
		RemovePhotosFromAlbum(router)
		r := PerformRequestWithBody(app, "DELETE", "/api/v1/albums/xxx/photos", `{"photos": ["ps6sg6be2lvl0yxx"]}`)
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
}

func TestCloneAlbums(t *testing.T) {
	app, router, _ := NewApiTest()
	CreateAlbum(router)
	r := PerformRequestWithBody(app, "POST", "/api/v1/albums", `{"Title": "Update", "Description": "To be updated", "Notes": "", "Favorite": true}`)
	assert.Equal(t, http.StatusOK, r.Code)
	uid := gjson.Get(r.Body.String(), "UID").String()

	t.Run("successful request", func(t *testing.T) {
		app, router, _ := NewApiTest()
		CloneAlbums(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/albums/"+uid+"/clone", `{"albums": ["`+uid+`"]}`)
		val := gjson.Get(r.Body.String(), "message")
		assert.Equal(t, "Album contents cloned", val.String())
		assert.Equal(t, http.StatusOK, r.Code)
	})

	t.Run("not found", func(t *testing.T) {
		app, router, _ := NewApiTest()
		CloneAlbums(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/albums/123/clone", `{albums: ["123"]}`)
		assert.Equal(t, http.StatusNotFound, r.Code)
	})

	t.Run("bad request", func(t *testing.T) {
		app, router, _ := NewApiTest()
		CloneAlbums(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/albums/"+uid+"/clone", `{albums: ["`+uid+`"]}`)
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, "Unable to do that", val.String())
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
}
