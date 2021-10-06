package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAlbumCover(t *testing.T) {
	t.Run("invalid type", func(t *testing.T) {
		app, router, conf := NewApiTest()
		AlbumCover(router)
		r := PerformRequest(app, "GET", "/api/v1/albums/at9lxuqxpogaaba7/t/"+conf.PreviewToken()+"/xxx")

		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("album contains no photos (because is not existing)", func(t *testing.T) {
		app, router, conf := NewApiTest()
		AlbumCover(router)
		r := PerformRequest(app, "GET", "/api/v1/albums/987-986435/t/"+conf.PreviewToken()+"/tile_500")
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("album: could not find original", func(t *testing.T) {
		app, router, conf := NewApiTest()
		AlbumCover(router)
		r := PerformRequest(app, "GET", "/api/v1/albums/at9lxuqxpogaaba9/t/"+conf.PreviewToken()+"/tile_500")
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("invalid token", func(t *testing.T) {
		app, router, _ := NewApiTest()
		AlbumCover(router)
		r := PerformRequest(app, "GET", "/api/v1/albums/at9lxuqxpogaaba8/t/xxx/tile_500")
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
}

func TestLabelCover(t *testing.T) {
	t.Run("invalid type", func(t *testing.T) {
		app, router, conf := NewApiTest()
		LabelCover(router)
		r := PerformRequest(app, "GET", "/api/v1/labels/lt9k3pw1wowuy3c2/t/"+conf.PreviewToken()+"/xxx")
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("invalid label", func(t *testing.T) {
		app, router, conf := NewApiTest()
		LabelCover(router)
		r := PerformRequest(app, "GET", "/api/v1/labels/xxx/t/"+conf.PreviewToken()+"/tile_500")

		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("could not find original", func(t *testing.T) {
		app, router, conf := NewApiTest()
		LabelCover(router)
		//r := PerformRequest(app, "GET", "/api/v1/labels/lt9k3pw1wowuy3c3/t/"+conf.PreviewToken()+"/tile_500")
		//lt9k3pw1wowuy3c2
		r := PerformRequest(app, "GET", "/api/v1/labels/lt9k3pw1wowuy3c2/t/"+conf.PreviewToken()+"/tile_500")
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("invalid token", func(t *testing.T) {
		app, router, _ := NewApiTest()
		LabelCover(router)
		r := PerformRequest(app, "GET", "/api/v1/labels/lt9k3pw1wowuy3c3/t/xxx/tile_500")
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
}
