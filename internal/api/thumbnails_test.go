package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetThumb(t *testing.T) {
	t.Run("invalid type", func(t *testing.T) {
		app, router, conf := NewApiTest()
		GetThumb(router)
		r := PerformRequest(app, "GET", "/api/v1/t/1/"+conf.PreviewToken()+"/xxx")

		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("invalid hash", func(t *testing.T) {
		app, router, conf := NewApiTest()
		GetThumb(router)
		r := PerformRequest(app, "GET", "/api/v1/t/1/"+conf.PreviewToken()+"/tile_500")

		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("could not find original", func(t *testing.T) {
		app, router, conf := NewApiTest()
		GetThumb(router)
		r := PerformRequest(app, "GET", "/api/v1/t/2cad9168fa6acc5c5c2965ddf6ec465ca42fd818/"+conf.PreviewToken()+"/fit_7680")
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("invalid token", func(t *testing.T) {
		app, router, _ := NewApiTest()
		GetThumb(router)
		r := PerformRequest(app, "GET", "/api/v1/t/2cad9168fa6acc5c5c2965ddf6ec465ca42fd818/xxx/tile_500")
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
	t.Run("no jpg", func(t *testing.T) {
		app, router, conf := NewApiTest()
		GetThumb(router)
		r := PerformRequest(app, "GET", "/api/v1/t/pcad9168fa6acc5c5ba965adf6ec465ca42fd819/"+conf.PreviewToken()+"/fit_7680")
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("file error", func(t *testing.T) {
		app, router, conf := NewApiTest()
		GetThumb(router)
		r := PerformRequest(app, "GET", "/api/v1/t/acad9168fa6acc5c5c2965ddf6ec465ca42fd832/"+conf.PreviewToken()+"/fit_7680")
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("NotFound", func(t *testing.T) {
		app, router, conf := NewApiTest()
		GetThumb(router)
		r := PerformRequest(app, "GET", "/api/v1/t/46f5b5c0c027f0c1b15136644f404c57210bf20c-016014058037/"+conf.PreviewToken()+"/tile_160")

		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("InvalidType", func(t *testing.T) {
		app, router, conf := NewApiTest()
		GetThumb(router)
		r := PerformRequest(app, "GET", "/api/v1/t/1-016014058037/"+conf.PreviewToken()+"/xxx")

		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("InvalidHash", func(t *testing.T) {
		app, router, conf := NewApiTest()
		GetThumb(router)
		r := PerformRequest(app, "GET", "/api/v1/t/1-016014058037/"+conf.PreviewToken()+"/tile_500")

		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("invalid token", func(t *testing.T) {
		app, router, _ := NewApiTest()
		GetThumb(router)
		r := PerformRequest(app, "GET", "/api/v1/t/2cad9168fa6acc5c5c2965ddf6ec465ca42fd818-016014058037/xxx/tile_500")
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
}
