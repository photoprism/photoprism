package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
)

func TestGetFolderCover(t *testing.T) {
	t.Run("NoCover", func(t *testing.T) {
		app, router, conf := NewApiTest()
		FolderCover(router)
		r := PerformRequest(app, "GET", "/api/v1/folders/t/dqo63pn35k2d495z/"+conf.PreviewToken()+"/tile_500")
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("InvalidType", func(t *testing.T) {
		app, router, conf := NewApiTest()
		FolderCover(router)
		r := PerformRequest(app, "GET", "/api/v1/folders/t/dqo63pn35k2d495z/"+conf.PreviewToken()+"/xxx")
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("InvalidToken", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		FolderCover(router)
		r := PerformRequest(app, "GET", "/api/v1/folders/t/dqo63pn35k2d495z/xxx/tile_500")
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
	t.Run("NotFound", func(t *testing.T) {
		app, router, conf := NewApiTest()
		FolderCover(router)
		r := PerformRequest(app, "GET", "/api/v1/folders/t/dqo63pn2f87f02oi/"+conf.PreviewToken()+"/fit_7680")
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("LimitsSize", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.Options().ThumbUncached = true
		defer func() { conf.Options().ThumbUncached = false }()
		CreateTestFolderCover(t, "dqo63pn35k2d495z", "1990/Photo16.jpg")
		FolderCover(router)
		small := PerformRequest(app, "GET", "/api/v1/folders/t/dqo63pn35k2d495z/"+conf.PreviewToken()+"/fit_720")
		large := PerformRequest(app, "GET", "/api/v1/folders/t/dqo63pn35k2d495z/"+conf.PreviewToken()+"/fit_1920")
		assert.Equal(t, http.StatusOK, large.Code)
		assert.Equal(t, "image/jpeg", large.Header().Get("Content-Type"))
		assert.Equal(t, small.Body.Bytes(), large.Body.Bytes())
	})
	t.Run("ServedInline", func(t *testing.T) {
		// Cover responses are always inline.
		app, router, conf := NewApiTest()
		conf.Options().ThumbUncached = true
		defer func() { conf.Options().ThumbUncached = false }()
		CreateTestFolderCover(t, "dqo63pn35k2d495z", "1990/Photo16.jpg")
		FolderCover(router)
		// Repeated so that the cache hit is covered as well as the render path.
		for i := 0; i < 2; i++ {
			r := PerformRequest(app, "GET", "/api/v1/folders/t/dqo63pn35k2d495z/"+conf.PreviewToken()+"/tile_500?download=1&name=original")
			assert.Equal(t, http.StatusOK, r.Code)
			assert.Equal(t, "image/jpeg", r.Header().Get("Content-Type"))
			assert.Empty(t, r.Header().Get("Content-Disposition"))
		}
	})
	t.Run("InvalidUID", func(t *testing.T) {
		app, router, conf := NewApiTest()
		FolderCover(router)
		r := PerformRequest(app, "GET", "/api/v1/folders/t/xxx/"+conf.PreviewToken()+"/tile_500")
		assert.Equal(t, http.StatusOK, r.Code)
		assert.Equal(t, "image/svg+xml", r.Header().Get("Content-Type"))
	})
	t.Run("SizeExceedsLimit", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.Options().ThumbUncached = true
		defer func() { conf.Options().ThumbUncached = false }()
		original := CreateTestFolderCover(t, "dqo63pn35k2d495z", "1990/Photo16.jpg")
		FolderCover(router)
		r := PerformRequest(app, "GET", "/api/v1/folders/t/dqo63pn35k2d495z/"+conf.PreviewToken()+"/fit_15360")
		assert.Equal(t, http.StatusOK, r.Code)
		assert.Equal(t, "image/jpeg", r.Header().Get("Content-Type"))
		assert.NotEqual(t, original, r.Body.Bytes())
	})
}
