package api

import (
	"encoding/json"
	"github.com/tidwall/gjson"
	"net/http"
	"testing"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/stretchr/testify/assert"
)

func TestLinkAlbum(t *testing.T) {
	t.Run("create share link", func(t *testing.T) {
		app, router, ctx := NewApiTest()

		var album entity.Album

		LinkAlbum(router, ctx)

		result1 := PerformRequestWithBody(app, "POST", "/api/v1/albums/at9lxuqxpogaaba7/link", `{"password": "foobar", "expires": 0, "edit": true}`)
		assert.Equal(t, http.StatusOK, result1.Code)

		if err := json.Unmarshal(result1.Body.Bytes(), &album); err != nil {
			t.Fatal(err)
		}

		if len(album.Links) != 1 {
			t.Fatalf("one link expected: %d, %+v", len(album.Links), album)
		}

		link := album.Links[0]

		assert.Equal(t, "foobar", link.LinkPassword)
		assert.Nil(t, link.LinkExpires)
		assert.False(t, link.CanComment)
		assert.True(t, link.CanEdit)

		result2 := PerformRequestWithBody(app, "POST", "/api/v1/albums/at9lxuqxpogaaba7/link", `{"password": "", "expires": 3600}`)

		assert.Equal(t, http.StatusOK, result2.Code)

		// t.Logf("result1: %s", result1.Body.String())
		// t.Logf("result2: %s", result2.Body.String())

		if err := json.Unmarshal(result2.Body.Bytes(), &album); err != nil {
			t.Fatal(err)
		}

		if len(album.Links) != 2 {
			t.Fatal("two links expected")
		}
	})
	t.Run("album not found", func(t *testing.T) {
		app, router, ctx := NewApiTest()
		LinkAlbum(router, ctx)
		r := PerformRequestWithBody(app, "POST", "/api/v1/albums/xxx/link", `{"password": "foobar", "expires": 0, "edit": true}`)
		assert.Equal(t, http.StatusNotFound, r.Code)
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, "Album not found", val.String())
	})
	t.Run("invalid request", func(t *testing.T) {
		app, router, ctx := NewApiTest()
		LinkAlbum(router, ctx)
		r := PerformRequestWithBody(app, "POST", "/api/v1/albums/at9lxuqxpogaaba7/link", `{"xxx": 123, "expires": 0, "edit": "xxx"}`)
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
}

func TestLinkPhoto(t *testing.T) {
	t.Run("create share link", func(t *testing.T) {
		app, router, ctx := NewApiTest()

		var photo entity.Photo

		LinkPhoto(router, ctx)

		result1 := PerformRequestWithBody(app, "POST", "/api/v1/photos/pt9jtdre2lvl0yh7/link", `{"password": "foobar", "expires": 0, "edit": true}`)
		assert.Equal(t, http.StatusOK, result1.Code)

		if err := json.Unmarshal(result1.Body.Bytes(), &photo); err != nil {
			t.Fatal(err)
		}

		if len(photo.Links) != 1 {
			t.Fatalf("one link expected: %d, %+v", len(photo.Links), photo)
		}

		link := photo.Links[0]

		assert.Equal(t, "foobar", link.LinkPassword)
		assert.Nil(t, link.LinkExpires)
		assert.False(t, link.CanComment)
		assert.True(t, link.CanEdit)

		result2 := PerformRequestWithBody(app, "POST", "/api/v1/photos/pt9jtdre2lvl0yh7/link", `{"password": "", "expires": 3600}`)

		assert.Equal(t, http.StatusOK, result2.Code)

		if err := json.Unmarshal(result2.Body.Bytes(), &photo); err != nil {
			t.Fatal(err)
		}

		if len(photo.Links) != 2 {
			t.Fatal("two links expected")
		}
	})
	t.Run("photo not found", func(t *testing.T) {
		app, router, ctx := NewApiTest()
		LinkPhoto(router, ctx)
		r := PerformRequestWithBody(app, "POST", "/api/v1/photos/xxx/link", `{"password": "foobar", "expires": 0, "edit": true}`)
		assert.Equal(t, http.StatusNotFound, r.Code)
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, "Photo not found", val.String())
	})
	t.Run("invalid request", func(t *testing.T) {
		app, router, ctx := NewApiTest()
		LinkPhoto(router, ctx)
		r := PerformRequestWithBody(app, "POST", "/api/v1/photos/pt9jtdre2lvl0yh7/link", `{"xxx": 123, "expires": 0, "edit": "xxx"}`)
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
}

func TestLinkLabel(t *testing.T) {
	t.Run("create share link", func(t *testing.T) {
		app, router, ctx := NewApiTest()

		var label entity.Label

		LinkLabel(router, ctx)

		result1 := PerformRequestWithBody(app, "POST", "/api/v1/labels/lt9k3pw1wowuy3c2/link", `{"password": "foobar", "expires": 0, "edit": true}`)
		assert.Equal(t, http.StatusOK, result1.Code)

		if err := json.Unmarshal(result1.Body.Bytes(), &label); err != nil {
			t.Fatal(err)
		}

		if len(label.Links) != 1 {
			t.Fatalf("one link expected: %d, %+v", len(label.Links), label)
		}

		link := label.Links[0]

		assert.Equal(t, "foobar", link.LinkPassword)
		assert.Nil(t, link.LinkExpires)
		assert.False(t, link.CanComment)
		assert.True(t, link.CanEdit)

		result2 := PerformRequestWithBody(app, "POST", "/api/v1/labels/lt9k3pw1wowuy3c2/link", `{"password": "", "expires": 3600}`)

		assert.Equal(t, http.StatusOK, result2.Code)

		if err := json.Unmarshal(result2.Body.Bytes(), &label); err != nil {
			t.Fatal(err)
		}

		if len(label.Links) != 2 {
			t.Fatal("two links expected")
		}
	})
	t.Run("label not found", func(t *testing.T) {
		app, router, ctx := NewApiTest()
		LinkLabel(router, ctx)
		r := PerformRequestWithBody(app, "POST", "/api/v1/labels/xxx/link", `{"password": "foobar", "expires": 0, "edit": true}`)
		assert.Equal(t, http.StatusNotFound, r.Code)
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, "Label not found", val.String())
	})
	t.Run("invalid request", func(t *testing.T) {
		app, router, ctx := NewApiTest()
		LinkLabel(router, ctx)
		r := PerformRequestWithBody(app, "POST", "/api/v1/labels/lt9k3pw1wowuy3c2/link", `{"xxx": 123, "expires": 0, "edit": "xxx"}`)
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
}
