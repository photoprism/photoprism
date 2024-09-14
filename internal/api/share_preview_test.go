package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
)

func TestGetPreview(t *testing.T) {
	t.Run("NotFound", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		SharePreview(router)
		r := PerformRequest(app, "GET", "api/v1/s/1jxf3jfn2k/ss6sg6bxpogaaba7/preview")
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("InvalidToken", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		SharePreview(router)
		r := PerformRequest(app, "GET", "api/v1/s/xxx/ss6sg6bxpogaaba7/preview")
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
}
