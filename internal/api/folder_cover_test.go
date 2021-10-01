package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFolderCover(t *testing.T) {
	t.Run("no cover yet", func(t *testing.T) {
		app, router, conf := NewApiTest()
		FolderCover(router)
		r := PerformRequest(app, "GET", "/api/v1/folders/t/dqo63pn35k2d495z/"+conf.PreviewToken()+"/tile_500")
		assert.Equal(t, http.StatusOK, r.Code)
	})

	t.Run("invalid thumb type", func(t *testing.T) {
		app, router, conf := NewApiTest()
		FolderCover(router)
		r := PerformRequest(app, "GET", "/api/v1/folders/t/dqo63pn35k2d495z/"+conf.PreviewToken()+"/xxx")
		assert.Equal(t, http.StatusOK, r.Code)
	})

	t.Run("invalid token", func(t *testing.T) {
		app, router, _ := NewApiTest()
		FolderCover(router)
		r := PerformRequest(app, "GET", "/api/v1/folders/t/dqo63pn35k2d495z/xxx/tile_500")
		assert.Equal(t, http.StatusForbidden, r.Code)
	})

	t.Run("could not find original", func(t *testing.T) {
		app, router, conf := NewApiTest()
		FolderCover(router)
		r := PerformRequest(app, "GET", "/api/v1/folders/t/dqo63pn2f87f02oi/"+conf.PreviewToken()+"/fit_7680")
		assert.Equal(t, http.StatusOK, r.Code)
	})
}
