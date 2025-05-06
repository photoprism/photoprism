package api

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
)

func TestCreateAlbumLink(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, router, _ := NewApiTest()

		var link entity.Link

		CreateAlbumLink(router)

		resp := PerformRequestWithBody(app, "POST", "/api/v1/albums/as6sg6bxpogaaba7/links", `{"Password": "foobar", "Expires": 0, "CanEdit": true}`)

		if resp.Code != http.StatusOK {
			t.Fatal(resp.Body.String())
		}

		if err := json.Unmarshal(resp.Body.Bytes(), &link); err != nil {
			t.Fatal(err)
		}

		assert.NotEmpty(t, link.LinkUID)
		assert.NotEmpty(t, link.ShareUID)
		assert.NotEmpty(t, link.LinkToken)
		assert.Equal(t, 0, link.LinkExpires)
	})
	t.Run("UserPasswordAdmin", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		UpdateUser(router)
		sessId := AuthenticateUser(app, router, "alice", "Alice123!")

		var link entity.Link

		CreateAlbumLink(router)

		resp := AuthenticatedRequestWithBody(app, "POST", "/api/v1/albums/as6sg6bxpogaaba7/links", `{"Password": "foobar", "Expires": 0, "CanEdit": true}`, sessId)

		if resp.Code != http.StatusOK {
			t.Fatal(resp.Body.String())
		}

		if err := json.Unmarshal(resp.Body.Bytes(), &link); err != nil {
			t.Fatal(err)
		}

		assert.NotEmpty(t, link.LinkUID)
		assert.NotEmpty(t, link.ShareUID)
		assert.NotEmpty(t, link.LinkToken)
		assert.Equal(t, 0, link.LinkExpires)
	})
	t.Run("UserPasswordGuest", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		UpdateUser(router)
		sessId := AuthenticateUser(app, router, "gandalf", "Gandalf123!")

		CreateAlbumLink(router)

		resp := AuthenticatedRequestWithBody(app, "POST", "/api/v1/albums/as6sg6bxpogaaba7/links", `{"Password": "foobar", "Expires": 0, "CanEdit": true}`, sessId)

		if resp.Code != http.StatusForbidden {
			t.Fatal(resp.Body.String())
		}

		val := gjson.Get(resp.Body.String(), "error")
		assert.Equal(t, "Permission denied", val.String())
	})
	t.Run("AliceAppPassword", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		var link entity.Link

		CreateAlbumLink(router)

		resp := AuthenticatedRequestWithBody(app, "POST", "/api/v1/albums/as6sg6bxpogaaba7/links", `{"Password": "foobar", "Expires": 0, "CanEdit": true}`, "X3B6IU-hfeLG5-HpVxkT-ctCY3M")

		if resp.Code != http.StatusOK {
			t.Fatal(resp.Body.String())
		}

		if err := json.Unmarshal(resp.Body.Bytes(), &link); err != nil {
			t.Fatal(err)
		}

		assert.NotEmpty(t, link.LinkUID)
		assert.NotEmpty(t, link.ShareUID)
		assert.NotEmpty(t, link.LinkToken)
		assert.Equal(t, 0, link.LinkExpires)
	})
	t.Run("UlfAppPassword", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		var link entity.Link

		CreateAlbumLink(router)

		resp := AuthenticatedRequestWithBody(app, "POST", "/api/v1/albums/as6sg6bxpogaaba7/links", `{"Password": "foobar", "Expires": 0, "CanEdit": true}`, "zlbkWC-vL4ORw-MvpDD7-mHrPW6")

		if resp.Code != http.StatusOK {
			t.Fatal(resp.Body.String())
		}

		if err := json.Unmarshal(resp.Body.Bytes(), &link); err != nil {
			t.Fatal(err)
		}

		assert.NotEmpty(t, link.LinkUID)
		assert.NotEmpty(t, link.ShareUID)
		assert.NotEmpty(t, link.LinkToken)
		assert.Equal(t, 0, link.LinkExpires)
	})
	t.Run("AliceAppPasswordWebdav", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		CreateAlbumLink(router)

		resp := AuthenticatedRequestWithBody(app, "POST", "/api/v1/albums/as6sg6bxpogaaba7/links", `{"Password": "foobar", "Expires": 0, "CanEdit": true}`, "v2wS72-OkqEzm-MQ63Z2-TEhU0w")

		if resp.Code != http.StatusForbidden {
			t.Fatal(resp.Body.String())
		}

		val := gjson.Get(resp.Body.String(), "error")
		assert.Equal(t, "Permission denied", val.String())
	})
	t.Run("GandalfAppPasswordFullAccess", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		CreateAlbumLink(router)

		resp := AuthenticatedRequestWithBody(app, "POST", "/api/v1/albums/as6sg6bxpogaaba7/links", `{"Password": "foobar", "Expires": 0, "CanEdit": true}`, "5hKyGD-FZUP9p-z8dYit-SYhkGI")

		if resp.Code != http.StatusForbidden {
			t.Fatal(resp.Body.String())
		}

		val := gjson.Get(resp.Body.String(), "error")
		assert.Equal(t, "Permission denied", val.String())
	})
	t.Run("AccessToken", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		var link entity.Link

		CreateAlbumLink(router)

		resp := AuthenticatedRequestWithBody(app, "POST", "/api/v1/albums/as6sg6bxpogaaba7/links", `{"Password": "foobar", "Expires": 0, "CanEdit": true}`, "8e154d323800393faf5177ce7392116feebbf674e6c2d39e")

		if resp.Code != http.StatusOK {
			t.Fatal(resp.Body.String())
		}

		if err := json.Unmarshal(resp.Body.Bytes(), &link); err != nil {
			t.Fatal(err)
		}

		assert.NotEmpty(t, link.LinkUID)
		assert.NotEmpty(t, link.ShareUID)
		assert.NotEmpty(t, link.LinkToken)
		assert.Equal(t, 0, link.LinkExpires)
	})
	t.Run("AlbumDoesNotExist", func(t *testing.T) {
		app, router, _ := NewApiTest()
		CreateAlbumLink(router)
		resp := PerformRequestWithBody(app, "POST", "/api/v1/albums/xxx/links", `{"Password": "foobar", "Expires": 0, "CanEdit": true}`)

		if resp.Code != http.StatusNotFound {
			t.Fatal(resp.Body.String())
		}

		val := gjson.Get(resp.Body.String(), "error")
		assert.Equal(t, "Album not found", val.String())
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		app, router, _ := NewApiTest()

		CreateAlbumLink(router)

		resp := PerformRequestWithBody(app, "POST", "/api/v1/albums/as6sg6bxpogaaba7/links", `{"Password": "foobar", "Expires": "abc", "CanEdit": true}`)

		if resp.Code != http.StatusBadRequest {
			t.Fatal(resp.Body.String())
		}
	})
}

func TestUpdateAlbumLink(t *testing.T) {
	app, router, _ := NewApiTest()

	CreateAlbumLink(router)

	r := PerformRequestWithBody(app, "POST", "/api/v1/albums/as6sg6bxpogaaba7/links", `{"Password": "foobar", "Expires": 0, "CanEdit": true}`)

	if r.Code != http.StatusOK {
		t.Fatal(r.Body.String())
	}
	val2 := gjson.Get(r.Body.String(), "Expires")
	assert.Equal(t, "0", val2.String())
	uid := gjson.Get(r.Body.String(), "UID").String()

	t.Run("Success", func(t *testing.T) {
		app, router, _ := NewApiTest()
		UpdateAlbumLink(router)
		r := PerformRequestWithBody(app, "PUT", "/api/v1/albums/as6sg6bxpogaaba7/links/"+uid, `{"Token": "newToken", "Expires": 8000, "Password": "1234nhfhfd"}`)
		val := gjson.Get(r.Body.String(), "Token")
		assert.Equal(t, "newtoken", val.String())
		val2 := gjson.Get(r.Body.String(), "Expires")
		assert.Equal(t, "8000", val2.String())
		assert.Equal(t, http.StatusOK, r.Code)
	})

	t.Run("bad request", func(t *testing.T) {
		app, router, _ := NewApiTest()
		UpdateAlbumLink(router)
		r := PerformRequestWithBody(app, "PUT", "/api/v1/albums/as6sg6bxpogaaba7/links/"+uid, `{"Token": "newToken", "Expires": "vgft", "xxx": "xxx"}`)
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
}

func TestDeleteAlbumLink(t *testing.T) {
	app, router, _ := NewApiTest()

	CreateAlbumLink(router)

	r := PerformRequestWithBody(app, "POST", "/api/v1/albums/as6sg6bxpogaaba7/links", `{"Password": "foobar", "Expires": 0, "CanEdit": true}`)

	if r.Code != http.StatusOK {
		t.Fatal(r.Body.String())
	}
	uid := gjson.Get(r.Body.String(), "UID").String()

	GetAlbumLinks(router)
	r2 := PerformRequest(app, "GET", "/api/v1/albums/as6sg6bxpogaaba7/links")
	len := gjson.Get(r2.Body.String(), "#")

	t.Run("successful deletion", func(t *testing.T) {
		app, router, _ := NewApiTest()
		DeleteAlbumLink(router)
		r := PerformRequest(app, "DELETE", "/api/v1/albums/as6sg6bxpogaaba7/links/"+uid)
		assert.Equal(t, http.StatusOK, r.Code)
		GetAlbumLinks(router)
		r2 := PerformRequest(app, "GET", "/api/v1/albums/as6sg6bxpogaaba7/links")
		len2 := gjson.Get(r2.Body.String(), "#")
		assert.Greater(t, len.Int(), len2.Int())
	})
}

func TestGetAlbumLinks(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, router, _ := NewApiTest()

		CreateAlbumLink(router)

		r := PerformRequestWithBody(app, "POST", "/api/v1/albums/as6sg6bxpogaaba7/links", `{"Password": "foobar", "Expires": 0, "CanEdit": true}`)

		if r.Code != http.StatusOK {
			t.Fatal(r.Body.String())
		}
		GetAlbumLinks(router)
		r2 := PerformRequest(app, "GET", "/api/v1/albums/as6sg6bxpogaaba7/links")
		len := gjson.Get(r2.Body.String(), "#")
		assert.GreaterOrEqual(t, len.Int(), int64(1))
		assert.Equal(t, http.StatusOK, r2.Code)
	})

	t.Run("NotFound", func(t *testing.T) {
		app, router, _ := NewApiTest()
		GetAlbumLinks(router)
		r := PerformRequest(app, "GET", "/api/v1/albums/xxx/links")
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
}

/*
func TestCreatePhotoLink(t *testing.T) {
	t.Run("create share link", func(t *testing.T) {
		app, router, _ := NewApiTest()

		var link entity.Link

		CreatePhotoLink(router)

		resp := PerformRequestWithBody(app, "POST", "/api/v1/photos/ps6sg6be2lvl0yh7/links", `{"Password":"foobar","Expires":0,"CanEdit":true}`)
		log.Debugf("BODY: %s", resp.Body.String())
		assert.Equal(t, http.StatusOK, resp.Code)

		if err := json.Unmarshal(resp.Body.Bytes(), &link); err != nil {
			t.Fatal(err)
		}

		assert.NotEmpty(t, link.LinkUID)
		assert.NotEmpty(t, link.ShareUID)
		assert.NotEmpty(t, link.LinkToken)
		assert.Equal(t, 0, link.LinkExpires)
		assert.False(t, link.CanComment)
		assert.True(t, link.CanEdit)
	})
	t.Run("photo not found", func(t *testing.T) {
		app, router, _ := NewApiTest()

		CreatePhotoLink(router)

		resp := PerformRequestWithBody(app, "POST", "/api/v1/photos/xxx/links", `{"Password": "foobar", "Expires": 0, "CanEdit": true}`)

		if resp.Code != http.StatusNotFound {
			t.Fatal(resp.Body.String())
		}
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		app, router, _ := NewApiTest()
		CreatePhotoLink(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/photos/ps6sg6be2lvl0yh7/links", `{"xxx": 123, "Expires": "abc", "CanEdit": "xxx"}`)
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
}

func TestUpdatePhotoLink(t *testing.T) {
	app, router, _ := NewApiTest()

	CreatePhotoLink(router)

	r := PerformRequestWithBody(app, "POST", "/api/v1/photos/ps6sg6be2lvl0yh7/links", `{"Password": "foobar", "Expires": 0, "CanEdit": true}`)

	if r.Code != http.StatusOK {
		t.Fatal(r.Body.String())
	}
	val2 := gjson.Get(r.Body.String(), "Expires")
	assert.Equal(t, "0", val2.String())
	uid := gjson.Get(r.Body.String(), "UID").String()

	t.Run("Success", func(t *testing.T) {
		app, router, _ := NewApiTest()
		UpdatePhotoLink(router)
		r := PerformRequestWithBody(app, "PUT", "/api/v1/photos/ps6sg6be2lvl0yh7/links/"+uid, `{"Token": "newToken", "Expires": 8000, "Password": "1234nhfhfd"}`)
		val := gjson.Get(r.Body.String(), "Token")
		assert.Equal(t, "newtoken", val.String())
		val2 := gjson.Get(r.Body.String(), "Expires")
		assert.Equal(t, "8000", val2.String())
		assert.Equal(t, http.StatusOK, r.Code)
	})

	t.Run("bad request", func(t *testing.T) {
		app, router, _ := NewApiTest()
		UpdatePhotoLink(router)
		r := PerformRequestWithBody(app, "PUT", "/api/v1/photos/ps6sg6be2lvl0yh7/links/"+uid, `{"Token": "newToken", "Expires": "vgft", "xxx": "xxx"}`)
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
}

// TODO Fully assert once functionality exists
func TestDeletePhotoLink(t *testing.T) {
	app, router, _ := NewApiTest()

	CreatePhotoLink(router)

	r := PerformRequestWithBody(app, "POST", "/api/v1/photos/ps6sg6be2lvl0yh7/links", `{"Password": "foobar", "Expires": 0, "CanEdit": true}`)

	if r.Code != http.StatusOK {
		t.Fatal(r.Body.String())
	}
	uid := gjson.Get(r.Body.String(), "UID").String()

	//GetPhotoLinks(router)
	//r2 := PerformRequest(app, "GET", "/api/v1/photos/ps6sg6be2lvl0yh7/links")
	//len := gjson.Get(r2.Body.String(), "#")

	t.Run("successful deletion", func(t *testing.T) {
		app, router, _ := NewApiTest()
		DeletePhotoLink(router)
		r := PerformRequest(app, "DELETE", "/api/v1/photos/ps6sg6be2lvl0yh7/links/"+uid)
		assert.Equal(t, http.StatusOK, r.Code)
		GetPhotoLinks(router)
		r2 := PerformRequest(app, "GET", "/api/v1/photos/ps6sg6be2lvl0yh7/links")
		t.Log(r2)
		//len2 := gjson.Get(r2.Body.String(), "#")
		//assert.Greater(t, len.Int(), len2.Int())
	})
}

func TestGetPhotoLinks(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, router, _ := NewApiTest()

		CreatePhotoLink(router)

		r := PerformRequestWithBody(app, "POST", "/api/v1/photos/ps6sg6be2lvl0yh7/links", `{"Password": "foobar", "Expires": 0, "CanEdit": true}`)
		if r.Code != http.StatusOK {
			t.Fatal(r.Body.String())
		}
		GetPhotoLinks(router)
		r2 := PerformRequest(app, "GET", "/api/v1/photos/ps6sg6be2lvl0yh7/links")
		//len := gjson.Get(r2.Body.String(), "#")
		//assert.GreaterOrEqual(t, len.Int(), int64(1))
		assert.Equal(t, http.StatusOK, r2.Code)
	})

	t.Run("NotFound", func(t *testing.T) {
		app, router, _ := NewApiTest()
		GetPhotoLinks(router)
		r := PerformRequest(app, "GET", "/api/v1/photos/xxx/links")
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
}

func TestCreateLabelLink(t *testing.T) {
	t.Run("create share link", func(t *testing.T) {
		app, router, _ := NewApiTest()

		var link entity.Link

		CreateLabelLink(router)

		resp := PerformRequestWithBody(app, "POST", "/api/v1/labels/ls6sg6b1wowuy3c2/links", `{"Password": "foobar", "Expires": 0, "CanEdit": true}`)
		assert.Equal(t, http.StatusOK, resp.Code)

		if err := json.Unmarshal(resp.Body.Bytes(), &link); err != nil {
			t.Fatal(err)
		}

		assert.NotEmpty(t, link.LinkUID)
		assert.NotEmpty(t, link.ShareUID)
		assert.NotEmpty(t, link.LinkToken)
		assert.Equal(t, 0, link.LinkExpires)
		assert.False(t, link.CanComment)
		assert.True(t, link.CanEdit)
	})
	t.Run("label not found", func(t *testing.T) {
		app, router, _ := NewApiTest()
		CreateLabelLink(router)
		resp := PerformRequestWithBody(app, "POST", "/api/v1/labels/xxx/links", `{"Password": "foobar", "Expires": 0, "CanEdit": true}`)

		if resp.Code != http.StatusNotFound {
			t.Fatal(resp.Body.String())
		}
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		app, router, _ := NewApiTest()
		CreateLabelLink(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/labels/ls6sg6b1wowuy3c2/links", `{"xxx": 123, "Expires": "abc", "CanEdit": "xxx"}`)
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
}

func TestUpdateLabelLink(t *testing.T) {
	app, router, _ := NewApiTest()

	CreateLabelLink(router)

	r := PerformRequestWithBody(app, "POST", "/api/v1/labels/ls6sg6b1wowuy3c2/links", `{"Password": "foobar", "Expires": 0, "CanEdit": true}`)

	if r.Code != http.StatusOK {
		t.Fatal(r.Body.String())
	}
	val2 := gjson.Get(r.Body.String(), "Expires")
	assert.Equal(t, "0", val2.String())
	uid := gjson.Get(r.Body.String(), "UID").String()

	t.Run("Success", func(t *testing.T) {
		app, router, _ := NewApiTest()
		UpdateLabelLink(router)
		r := PerformRequestWithBody(app, "PUT", "/api/v1/labels/ls6sg6b1wowuy3c2/links/"+uid, `{"Token": "newToken", "Expires": 8000, "Password": "1234nhfhfd"}`)
		val := gjson.Get(r.Body.String(), "Token")
		assert.Equal(t, "newtoken", val.String())
		val2 := gjson.Get(r.Body.String(), "Expires")
		assert.Equal(t, "8000", val2.String())
		assert.Equal(t, http.StatusOK, r.Code)
	})

	t.Run("bad request", func(t *testing.T) {
		app, router, _ := NewApiTest()
		UpdateLabelLink(router)
		r := PerformRequestWithBody(app, "PUT", "/api/v1/labels/ls6sg6b1wowuy3c2/links/"+uid, `{"Token": "newToken", "Expires": "vgft", "xxx": "xxx"}`)
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
}

func TestDeleteLabelLink(t *testing.T) {
	app, router, _ := NewApiTest()

	CreateLabelLink(router)

	r := PerformRequestWithBody(app, "POST", "/api/v1/labels/ls6sg6b1wowuy3c2/links", `{"Password": "foobar", "Expires": 0, "CanEdit": true}`)

	if r.Code != http.StatusOK {
		t.Fatal(r.Body.String())
	}
	uid := gjson.Get(r.Body.String(), "UID").String()

	//GetLabelLinks(router)
	//r2 := PerformRequest(app, "GET", "/api/v1/labels/ls6sg6b1wowuy3c2/links")
	//len := gjson.Get(r2.Body.String(), "#")

	t.Run("successful deletion", func(t *testing.T) {
		app, router, _ := NewApiTest()
		DeleteLabelLink(router)
		r := PerformRequest(app, "DELETE", "/api/v1/labels/ls6sg6b1wowuy3c2/links/"+uid)
		assert.Equal(t, http.StatusOK, r.Code)
		//GetLabelLinks(router)
		//r2 := PerformRequest(app, "GET", "/api/v1/labels/ls6sg6b1wowuy3c2/links")
		//len2 := gjson.Get(r2.Body.String(), "#")
		//assert.Greater(t, len.Int(), len2.Int())
	})
}

func TestGetLabelLinks(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, router, _ := NewApiTest()

		CreateLabelLink(router)

		r := PerformRequestWithBody(app, "POST", "/api/v1/labels/ls6sg6b1wowuy3c2/links", `{"Password": "foobar", "Expires": 0, "CanEdit": true}`)

		if r.Code != http.StatusOK {
			t.Fatal(r.Body.String())
		}
		GetLabelLinks(router)
		r2 := PerformRequest(app, "GET", "/api/v1/labels/ls6sg6b1wowuy3c2/links")
		//len := gjson.Get(r2.Body.String(), "#")
		//assert.GreaterOrEqual(t, len.Int(), int64(1))
		assert.Equal(t, http.StatusOK, r2.Code)
	})

	t.Run("NotFound", func(t *testing.T) {
		app, router, _ := NewApiTest()
		GetLabelLinks(router)
		r := PerformRequest(app, "GET", "/api/v1/labels/xxx/links")
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
}
*/
