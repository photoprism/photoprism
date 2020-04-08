package api

import (
	"encoding/json"
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

		result1 := PerformRequestWithBody(app, "POST", "/api/v1/albums/3/link", `{"password": "foobar", "expires": 0, "edit": true}`)

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

		result2 := PerformRequestWithBody(app, "POST", "/api/v1/albums/3/link", `{"password": "", "expires": 3600}`)

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
}
