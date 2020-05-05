package api

import (
	"github.com/tidwall/gjson"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAlbums(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		app, router, conf := NewApiTest()
		GetAlbums(router, conf)
		r := PerformRequest(app, "GET", "/api/v1/albums?count=10")
		len := gjson.Get(r.Body.String(), "#")
		assert.LessOrEqual(t, int64(3), len.Int())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("invalid request", func(t *testing.T) {
		app, router, conf := NewApiTest()
		GetAlbums(router, conf)
		r := PerformRequest(app, "GET", "/api/v1/albums?xxx=10")
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
}

func TestGetAlbum(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		app, router, conf := NewApiTest()
		GetAlbum(router, conf)
		r := PerformRequest(app, "GET", "/api/v1/albums/at9lxuqxpogaaba8")
		val := gjson.Get(r.Body.String(), "AlbumSlug")
		assert.Equal(t, "holiday-2030", val.String())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("invalid request", func(t *testing.T) {
		app, router, conf := NewApiTest()
		GetAlbum(router, conf)
		r := PerformRequest(app, "GET", "/api/v1/albums/999000")
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, "Album not found", val.String())
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
}

func TestCreateAlbum(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		app, router, conf := NewApiTest()
		CreateAlbum(router, conf)
		r := PerformRequestWithBody(app, "POST", "/api/v1/albums", `{"AlbumName": "New created album", "AlbumNotes": "", "AlbumFavorite": true}`)
		val := gjson.Get(r.Body.String(), "AlbumSlug")
		assert.Equal(t, "new-created-album", val.String())
		val2 := gjson.Get(r.Body.String(), "AlbumFavorite")
		assert.Equal(t, "true", val2.String())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("invalid request", func(t *testing.T) {
		app, router, conf := NewApiTest()
		CreateAlbum(router, conf)
		r := PerformRequestWithBody(app, "POST", "/api/v1/albums", `{"AlbumName": 333, "AlbumDescription": "Created via unit test", "AlbumNotes": "", "AlbumFavorite": true}`)
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
}
func TestUpdateAlbum(t *testing.T) {
	app, router, conf := NewApiTest()
	CreateAlbum(router, conf)
	r := PerformRequestWithBody(app, "POST", "/api/v1/albums", `{"AlbumName": "Update", "AlbumDescription": "To be updated", "AlbumNotes": "", "AlbumFavorite": true}`)
	assert.Equal(t, http.StatusOK, r.Code)
	uuid := gjson.Get(r.Body.String(), "AlbumUUID").String()

	t.Run("successful request", func(t *testing.T) {
		app, router, conf := NewApiTest()
		UpdateAlbum(router, conf)
		r := PerformRequestWithBody(app, "PUT", "/api/v1/albums/"+uuid, `{"AlbumName": "Updated01", "AlbumNotes": "", "AlbumFavorite": false}`)
		val := gjson.Get(r.Body.String(), "AlbumSlug")
		assert.Equal(t, "updated01", val.String())
		val2 := gjson.Get(r.Body.String(), "AlbumFavorite")
		assert.Equal(t, "false", val2.String())
		assert.Equal(t, http.StatusOK, r.Code)
	})

	t.Run("invalid request", func(t *testing.T) {
		app, router, conf := NewApiTest()
		UpdateAlbum(router, conf)
		r := PerformRequestWithBody(app, "PUT", "/api/v1/albums"+uuid, `{"AlbumName": 333, "AlbumDescription": "Created via unit test", "AlbumNotes": "", "AlbumFavorite": true}`)
		assert.Equal(t, http.StatusNotFound, r.Code)
	})

	t.Run("not found", func(t *testing.T) {
		app, router, conf := NewApiTest()
		UpdateAlbum(router, conf)
		r := PerformRequestWithBody(app, "PUT", "/api/v1/albums/xxx", `{"AlbumName": "Update03", "AlbumDescription": "Created via unit test", "AlbumNotes": "", "AlbumFavorite": true}`)
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, "Album not found", val.String())
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
}
func TestDeleteAlbum(t *testing.T) {
	app, router, conf := NewApiTest()
	CreateAlbum(router, conf)
	r := PerformRequestWithBody(app, "POST", "/api/v1/albums", `{"AlbumName": "Delete", "AlbumDescription": "To be deleted", "AlbumNotes": "", "AlbumFavorite": true}`)
	assert.Equal(t, http.StatusOK, r.Code)
	uuid := gjson.Get(r.Body.String(), "AlbumUUID").String()

	t.Run("delete existing album", func(t *testing.T) {
		app, router, conf := NewApiTest()
		DeleteAlbum(router, conf)
		r := PerformRequest(app, "DELETE", "/api/v1/albums/"+uuid)
		assert.Equal(t, http.StatusOK, r.Code)
		val := gjson.Get(r.Body.String(), "AlbumSlug")
		assert.Equal(t, "delete", val.String())
		GetAlbums(router, conf)
		r2 := PerformRequest(app, "GET", "/api/v1/albums/"+uuid)
		assert.Equal(t, http.StatusNotFound, r2.Code)
	})
	t.Run("delete not existing album", func(t *testing.T) {
		app, router, conf := NewApiTest()
		DeleteAlbum(router, conf)
		r := PerformRequest(app, "DELETE", "/api/v1/albums/999000")
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, "Album not found", val.String())
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
}

func TestLikeAlbum(t *testing.T) {
	t.Run("like not existing album", func(t *testing.T) {
		app, router, ctx := NewApiTest()

		LikeAlbum(router, ctx)

		r := PerformRequest(app, "POST", "/api/v1/albums/xxx/like")
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("like existing album", func(t *testing.T) {
		app, router, ctx := NewApiTest()

		LikeAlbum(router, ctx)
		r := PerformRequest(app, "POST", "/api/v1/albums/at9lxuqxpogaaba7/like")
		assert.Equal(t, http.StatusOK, r.Code)
		GetAlbum(router, ctx)
		r2 := PerformRequest(app, "GET", "/api/v1/albums/at9lxuqxpogaaba7")
		val := gjson.Get(r2.Body.String(), "AlbumFavorite")
		assert.Equal(t, "true", val.String())
	})
}

func TestDislikeAlbum(t *testing.T) {
	t.Run("dislike not existing album", func(t *testing.T) {
		app, router, conf := NewApiTest()

		DislikeAlbum(router, conf)

		r := PerformRequest(app, "DELETE", "/api/v1/albums/5678/like")
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("dislike existing album", func(t *testing.T) {
		app, router, conf := NewApiTest()

		DislikeAlbum(router, conf)

		r := PerformRequest(app, "DELETE", "/api/v1/albums/at9lxuqxpogaaba8/like")
		assert.Equal(t, http.StatusOK, r.Code)
		GetAlbum(router, conf)
		r2 := PerformRequest(app, "GET", "/api/v1/albums/at9lxuqxpogaaba8")
		val := gjson.Get(r2.Body.String(), "AlbumFavorite")
		assert.Equal(t, "false", val.String())
	})
}

func TestAddPhotosToAlbum(t *testing.T) {
	app, router, conf := NewApiTest()
	CreateAlbum(router, conf)
	r := PerformRequestWithBody(app, "POST", "/api/v1/albums", `{"AlbumName": "Add photos", "AlbumDescription": "", "AlbumNotes": "", "AlbumFavorite": true}`)
	assert.Equal(t, http.StatusOK, r.Code)
	uuid := gjson.Get(r.Body.String(), "AlbumUUID").String()

	t.Run("successful request", func(t *testing.T) {
		app, router, conf := NewApiTest()
		AddPhotosToAlbum(router, conf)
		r := PerformRequestWithBody(app, "POST", "/api/v1/albums/"+uuid+"/photos", `{"photos": ["pt9jtdre2lvl0y12", "pt9jtdre2lvl0y11"]}`)
		val := gjson.Get(r.Body.String(), "message")
		assert.Equal(t, "photos added to album", val.String())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("add one photo to album", func(t *testing.T) {
		app, router, conf := NewApiTest()
		AddPhotosToAlbum(router, conf)
		r := PerformRequestWithBody(app, "POST", "/api/v1/albums/"+uuid+"/photos", `{"photos": ["pt9jtdre2lvl0y12"]}`)
		val := gjson.Get(r.Body.String(), "message")
		assert.Equal(t, "photos added to album", val.String())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("invalid request", func(t *testing.T) {
		app, router, conf := NewApiTest()
		AddPhotosToAlbum(router, conf)
		r := PerformRequestWithBody(app, "POST", "/api/v1/albums/"+uuid+"/photos", `{"photos": [123, "pt9jtdre2lvl0yxx"]}`)
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("not found", func(t *testing.T) {
		app, router, conf := NewApiTest()
		AddPhotosToAlbum(router, conf)
		r := PerformRequestWithBody(app, "POST", "/api/v1/albums/xxx/photos", `{"photos": ["pt9jtdre2lvl0yxx"]}`)
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
}

func TestRemovePhotosFromAlbum(t *testing.T) {
	app, router, conf := NewApiTest()
	CreateAlbum(router, conf)
	r := PerformRequestWithBody(app, "POST", "/api/v1/albums", `{"AlbumName": "Remove photos", "AlbumDescription": "", "AlbumNotes": "", "AlbumFavorite": true}`)
	assert.Equal(t, http.StatusOK, r.Code)
	uuid := gjson.Get(r.Body.String(), "AlbumUUID").String()
	AddPhotosToAlbum(router, conf)
	r2 := PerformRequestWithBody(app, "POST", "/api/v1/albums/"+uuid+"/photos", `{"photos": ["pt9jtdre2lvl0y12", "pt9jtdre2lvl0y11"]}`)
	assert.Equal(t, http.StatusOK, r2.Code)

	t.Run("successful request", func(t *testing.T) {
		app, router, conf := NewApiTest()
		RemovePhotosFromAlbum(router, conf)
		r := PerformRequestWithBody(app, "DELETE", "/api/v1/albums/"+uuid+"/photos", `{"photos": ["pt9jtdre2lvl0y12", "pt9jtdre2lvl0y11"]}`)
		val := gjson.Get(r.Body.String(), "message")
		assert.Equal(t, "photos removed from album", val.String())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("no photos selected", func(t *testing.T) {
		app, router, conf := NewApiTest()
		RemovePhotosFromAlbum(router, conf)
		r := PerformRequestWithBody(app, "DELETE", "/api/v1/albums/at9lxuqxpogaaba7/photos", `{"photos": []}`)
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, "No photos selected", val.String())
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("invalid request", func(t *testing.T) {
		app, router, conf := NewApiTest()
		RemovePhotosFromAlbum(router, conf)
		r := PerformRequestWithBody(app, "DELETE", "/api/v1/albums/"+uuid+"/photos", `{"photos": [123, "pt9jtdre2lvl0yxx"]}`)
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("album not found", func(t *testing.T) {
		app, router, conf := NewApiTest()
		RemovePhotosFromAlbum(router, conf)
		r := PerformRequestWithBody(app, "DELETE", "/api/v1/albums/xxx/photos", `{"photos": ["pt9jtdre2lvl0yxx"]}`)
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
}

func TestDownloadAlbum(t *testing.T) {
	t.Run("download not existing album", func(t *testing.T) {
		app, router, conf := NewApiTest()

		DownloadAlbum(router, conf)

		r := PerformRequest(app, "GET", "/api/v1/albums/5678/download")
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("download existing album", func(t *testing.T) {
		app, router, conf := NewApiTest()

		DownloadAlbum(router, conf)

		r := PerformRequest(app, "GET", "/api/v1/albums/at9lxuqxpogaaba8/download")
		assert.Equal(t, http.StatusOK, r.Code)
	})
}

func TestAlbumThumbnail(t *testing.T) {
	t.Run("invalid type", func(t *testing.T) {
		app, router, ctx := NewApiTest()
		AlbumThumbnail(router, ctx)
		r := PerformRequest(app, "GET", "/api/v1/albums/at9lxuqxpogaaba7/thumbnail/xxx")

		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("album has no photo (because is not existing)", func(t *testing.T) {
		app, router, ctx := NewApiTest()
		AlbumThumbnail(router, ctx)
		r := PerformRequest(app, "GET", "/api/v1/albums/987-986435/thumbnail/tile_500")
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("album: could not find original", func(t *testing.T) {
		app, router, ctx := NewApiTest()
		AlbumThumbnail(router, ctx)
		r := PerformRequest(app, "GET", "/api/v1/albums/at9lxuqxpogaaba8/thumbnail/tile_500")
		assert.Equal(t, http.StatusOK, r.Code)
	})
}
