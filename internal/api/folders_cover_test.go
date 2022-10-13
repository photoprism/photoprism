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
}
