package api

import (
	"encoding/json"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/tidwall/gjson"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFile(t *testing.T) {
	t.Run("search for existing file", func(t *testing.T) {
		app, router, ctx := NewApiTest()
		GetFile(router, ctx)
		r := PerformRequest(app, "GET", "/api/v1/files/2cad9168fa6acc5c5c2965ddf6ec465ca42fd818")
		assert.Equal(t, http.StatusOK, r.Code)

		val := gjson.Get(r.Body.String(), "Name")
		assert.Equal(t, "exampleFileName.jpg", val.String())
	})
	t.Run("search for not existing file", func(t *testing.T) {
		app, router, ctx := NewApiTest()
		GetFile(router, ctx)
		r := PerformRequest(app, "GET", "/api/v1/files/111")
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
}

func TestLinkFile(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		app, router, ctx := NewApiTest()
		LinkFile(router, ctx)
		r := PerformRequestWithBody(app, "POST", "/api/v1/files/ft9es39w45bnlqdw/link", `{"Password": "foobar123", "Expires": 0, "CanEdit": true}`)

		var label entity.Label

		if err := json.Unmarshal(r.Body.Bytes(), &label); err != nil {
			t.Fatal(err)
		}

		if len(label.Links) != 1 {
			t.Fatalf("one link expected: %d, %+v", len(label.Links), label)
		}

		link := label.Links[0]

		assert.Equal(t, "foobar123", link.LinkPassword)
		assert.Nil(t, link.LinkExpires)
		assert.False(t, link.CanComment)
		assert.True(t, link.CanEdit)
	})
	t.Run("file not found", func(t *testing.T) {
		app, router, ctx := NewApiTest()
		LinkFile(router, ctx)
		r := PerformRequestWithBody(app, "POST", "/api/v1/files/xxx/link", `{"Password": "foobar", "Expires": 0, "CanEdit": true}`)
		assert.Equal(t, http.StatusNotFound, r.Code)
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, "File not found", val.String())
	})
	t.Run("invalid request", func(t *testing.T) {
		app, router, ctx := NewApiTest()
		LinkFile(router, ctx)
		r := PerformRequestWithBody(app, "POST", "/api/v1/files/ft9es39w45bnlqdw/link", `{"xxx": 123, "Expires": 0, "CanEdit": "xxx"}`)
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
}
