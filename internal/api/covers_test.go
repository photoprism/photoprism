package api

import (
	"net/http"
	"testing"

	"github.com/photoprism/photoprism/internal/config"

	"github.com/stretchr/testify/assert"
)

func TestAlbumCover(t *testing.T) {
	t.Run("InvalidType", func(t *testing.T) {
		app, router, conf := NewApiTest()
		AlbumCover(router)
		r := PerformRequest(app, "GET", "/api/v1/albums/as6sg6bxpogaaba7/t/"+conf.PreviewToken()+"/xxx")

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
		r := PerformRequest(app, "GET", "/api/v1/albums/as6sg6bxpogaaba9/t/"+conf.PreviewToken()+"/tile_500")
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("InvalidToken", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		AlbumCover(router)
		r := PerformRequest(app, "GET", "/api/v1/albums/as6sg6bxpogaaba8/t/xxx/tile_500")
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
}

func TestLabelCover(t *testing.T) {
	t.Run("InvalidType", func(t *testing.T) {
		app, router, conf := NewApiTest()
		LabelCover(router)
		r := PerformRequest(app, "GET", "/api/v1/labels/ls6sg6b1wowuy3c2/t/"+conf.PreviewToken()+"/xxx")
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
		//r := PerformRequest(app, "GET", "/api/v1/labels/ls6sg6b1wowuy3c3/t/"+conf.PreviewToken()+"/tile_500")
		//ls6sg6b1wowuy3c2
		r := PerformRequest(app, "GET", "/api/v1/labels/ls6sg6b1wowuy3c2/t/"+conf.PreviewToken()+"/tile_500")
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("InvalidToken", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		LabelCover(router)
		r := PerformRequest(app, "GET", "/api/v1/labels/ls6sg6b1wowuy3c3/t/xxx/tile_500")
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
}
