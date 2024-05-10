package api

import (
	"net/http"
	"testing"

	"github.com/photoprism/photoprism/internal/query"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"

	"github.com/photoprism/photoprism/pkg/i18n"
)

func TestGetAlbum(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, router, _ := NewApiTest()
		GetAlbum(router)
		r := PerformRequest(app, "GET", "/api/v1/albums/as6sg6bxpogaaba8")
		val := gjson.Get(r.Body.String(), "Slug")
		assert.Equal(t, "holiday-2030", val.String())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("Invalid", func(t *testing.T) {
		app, router, _ := NewApiTest()
		GetAlbum(router)
		r := PerformRequest(app, "GET", "/api/v1/albums/999000")
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, "Album not found", val.String())
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
}

func TestCreateAlbum(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, router, _ := NewApiTest()
		CreateAlbum(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/albums", `{"Title": "New created album", "Notes": "", "Favorite": true}`)
		val := gjson.Get(r.Body.String(), "Slug")
		assert.Equal(t, "new-created-album", val.String())
		val2 := gjson.Get(r.Body.String(), "Favorite")
		assert.Equal(t, "true", val2.String())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("Invalid", func(t *testing.T) {
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

	t.Run("Successful", func(t *testing.T) {
		app, router, _ := NewApiTest()
		UpdateAlbum(router)
		r := PerformRequestWithBody(app, "PUT", "/api/v1/albums/"+uid, `{"Title": "Updated01", "Notes": "", "Favorite": false}`)
		val := gjson.Get(r.Body.String(), "Slug")
		assert.Equal(t, "updated01", val.String())
		val2 := gjson.Get(r.Body.String(), "Favorite")
		assert.Equal(t, "false", val2.String())
		assert.Equal(t, http.StatusOK, r.Code)
	})

	t.Run("Invalid", func(t *testing.T) {
		app, router, _ := NewApiTest()
		UpdateAlbum(router)
		r := PerformRequestWithBody(app, "PUT", "/api/v1/albums"+uid, `{"Title": 333, "Description": "Created via unit test", "Notes": "", "Favorite": true}`)
		assert.Equal(t, http.StatusNotFound, r.Code)
	})

	t.Run("NotFound", func(t *testing.T) {
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

	t.Run("ExistingAlbum", func(t *testing.T) {
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
	t.Run("NotExistingAlbum", func(t *testing.T) {
		app, router, _ := NewApiTest()
		DeleteAlbum(router)
		r := PerformRequest(app, "DELETE", "/api/v1/albums/999000")
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, "Album not found", val.String())
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("ExistingMoment", func(t *testing.T) {
		app, router, _ = NewApiTest()
		DeleteAlbum(router)
		r = PerformRequest(app, "DELETE", "/api/v1/albums/as6sg6bipotaaj10")
		assert.Equal(t, http.StatusOK, r.Code)
		val := gjson.Get(r.Body.String(), "Slug")
		assert.Equal(t, "mexico", val.String())
		SearchAlbums(router)
		r2 := PerformRequest(app, "GET", "/api/v1/albums/as6sg6bipotaaj10")
		assert.Equal(t, http.StatusNotFound, r2.Code)
	})
}

func TestLikeAlbum(t *testing.T) {
	t.Run("NotExistingAlbum", func(t *testing.T) {
		app, router, _ := NewApiTest()

		LikeAlbum(router)

		r := PerformRequest(app, "POST", "/api/v1/albums/xxx/like")
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("ExistingAlbum", func(t *testing.T) {
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
	t.Run("NotExistingAlbum", func(t *testing.T) {
		app, router, _ := NewApiTest()

		DislikeAlbum(router)

		r := PerformRequest(app, "DELETE", "/api/v1/albums/5678/like")
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("ExistingAlbum", func(t *testing.T) {
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

	t.Run("AddMultiplePhotos", func(t *testing.T) {
		app, router, _ := NewApiTest()
		AddPhotosToAlbum(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/albums/"+uid+"/photos", `{"photos": ["ps6sg6be2lvl0y12", "ps6sg6be2lvl0y11"]}`)
		val := gjson.Get(r.Body.String(), "message")
		assert.Equal(t, i18n.Msg(i18n.MsgChangesSaved), val.String())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("AddSinglePhoto", func(t *testing.T) {
		app, router, _ := NewApiTest()
		AddPhotosToAlbum(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/albums/"+uid+"/photos", `{"photos": ["ps6sg6be2lvl0y12"]}`)
		val := gjson.Get(r.Body.String(), "message")
		assert.Equal(t, i18n.Msg(i18n.MsgChangesSaved), val.String())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("AddPhotoFromReview", func(t *testing.T) {
		p, err := query.PhotoByUID("ps6sg6byk7wrbk44")

		if err != nil {
			t.Fatal(err)
		}

		assert.Less(t, p.PhotoQuality, 3)

		app, router, _ := NewApiTest()
		AddPhotosToAlbum(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/albums/"+uid+"/photos", `{"photos": ["ps6sg6byk7wrbk44"]}`)
		val := gjson.Get(r.Body.String(), "message")
		assert.Equal(t, i18n.Msg(i18n.MsgChangesSaved), val.String())
		assert.Equal(t, http.StatusOK, r.Code)

		p, err = query.PhotoByUID("ps6sg6byk7wrbk44")

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, p.PhotoQuality, 3)
	})
	t.Run("NoPhotosSelected", func(t *testing.T) {
		app, router, _ := NewApiTest()
		AddPhotosToAlbum(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/albums/"+uid+"/photos", `{"photos": []}`)
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("Invalid", func(t *testing.T) {
		app, router, _ := NewApiTest()
		AddPhotosToAlbum(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/albums/"+uid+"/photos", `{"photos": [123, "ps6sg6be2lvl0yxx"]}`)
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("NotFound", func(t *testing.T) {
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

	t.Run("RemoveMultiplePhotos", func(t *testing.T) {
		app, router, _ := NewApiTest()
		RemovePhotosFromAlbum(router)
		r := PerformRequestWithBody(app, "DELETE", "/api/v1/albums/"+uid+"/photos", `{"photos": ["ps6sg6be2lvl0y12", "ps6sg6be2lvl0y11"]}`)
		val := gjson.Get(r.Body.String(), "message")
		assert.Equal(t, i18n.Msg(i18n.MsgChangesSaved), val.String())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("RemoveSinglePhoto", func(t *testing.T) {
		app, router, _ := NewApiTest()
		RemovePhotosFromAlbum(router)
		r := PerformRequestWithBody(app, "DELETE", "/api/v1/albums/"+uid+"/photos", `{"photos": ["ps6sg6be2lvl0y12"]}`)
		val := gjson.Get(r.Body.String(), "message")
		assert.Equal(t, i18n.Msg(i18n.MsgChangesSaved), val.String())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("NoPhotosSelected", func(t *testing.T) {
		app, router, _ := NewApiTest()
		RemovePhotosFromAlbum(router)
		r := PerformRequestWithBody(app, "DELETE", "/api/v1/albums/as6sg6bxpogaaba7/photos", `{"photos": []}`)
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, "No items selected", val.String())
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("Invalid", func(t *testing.T) {
		app, router, _ := NewApiTest()
		RemovePhotosFromAlbum(router)
		r := PerformRequestWithBody(app, "DELETE", "/api/v1/albums/"+uid+"/photos", `{"photos": [123, "ps6sg6be2lvl0yxx"]}`)
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("AlbumNotFound", func(t *testing.T) {
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

	t.Run("CloneEmptyAlbum", func(t *testing.T) {
		app, router, _ := NewApiTest()
		CloneAlbums(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/albums/"+uid+"/clone", `{"albums": ["`+uid+`"]}`)
		val := gjson.Get(r.Body.String(), "message")
		assert.Equal(t, "Album contents cloned", val.String())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("CloneAlbum", func(t *testing.T) {
		app, router, _ := NewApiTest()
		CloneAlbums(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/albums/as6sg6bxpogaaba9/clone", `{"albums": ["as6sg6bxpogaaba9"]}`)
		val := gjson.Get(r.Body.String(), "message")
		assert.Equal(t, "Album contents cloned", val.String())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("CloneMoment", func(t *testing.T) {
		app, router, _ := NewApiTest()
		CloneAlbums(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/albums/as6sg6bipotaah64/clone", `{"albums": ["as6sg6bipotaah64"]}`)
		val := gjson.Get(r.Body.String(), "message")
		assert.Equal(t, "Album contents cloned", val.String())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("NotFound", func(t *testing.T) {
		app, router, _ := NewApiTest()
		CloneAlbums(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/albums/123/clone", `{albums: ["123"]}`)
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("BadRequest", func(t *testing.T) {
		app, router, _ := NewApiTest()
		CloneAlbums(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/albums/"+uid+"/clone", `{albums: ["`+uid+`"]}`)
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, "Unable to do that", val.String())
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
}
