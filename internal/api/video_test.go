package api

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/media/video"
)

func TestGetVideo(t *testing.T) {
	t.Run("ContentTypeAVC", func(t *testing.T) {
		assert.Equal(t, video.ContentTypeAVC, fmt.Sprintf("%s; codecs=\"%s\"", "video/mp4", clean.Codec("avc1")))
	})

	t.Run("InvalidHash", func(t *testing.T) {
		app, router, conf := NewApiTest()
		GetVideo(router)
		r := PerformRequest(app, "GET", "/api/v1/videos/xxx/"+conf.PreviewToken()+"/mp4")
		assert.Equal(t, http.StatusOK, r.Code)
	})

	t.Run("InvalidType", func(t *testing.T) {
		app, router, conf := NewApiTest()
		GetVideo(router)
		r := PerformRequest(app, "GET", "/api/v1/videos/acad9168fa6acc5c5c2965ddf6ec465ca42fd831/"+conf.PreviewToken()+"/xxx")
		assert.Equal(t, http.StatusOK, r.Code)
	})

	t.Run("NotFound", func(t *testing.T) {
		app, router, conf := NewApiTest()
		GetVideo(router)
		r := PerformRequest(app, "GET", "/api/v1/videos/acad9168fa6acc5c5c2965ddf6ec465ca42fd831/"+conf.PreviewToken()+"/mp4")
		assert.Equal(t, http.StatusOK, r.Code)
	})

	t.Run("FileError", func(t *testing.T) {
		app, router, conf := NewApiTest()
		GetVideo(router)
		r := PerformRequest(app, "GET", "/api/v1/videos/acad9168fa6acc5c5c2965ddf6ec465ca42fd832/"+conf.PreviewToken()+"/mp4")
		assert.Equal(t, http.StatusOK, r.Code)
	})

	t.Run("InvalidToken", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		GetVideo(router)
		r := PerformRequest(app, "GET", "/api/v1/videos/acad9168fa6acc5c5c2965ddf6ec465ca42fd832/xxx/mp4")
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
	t.Run("NoVideo", func(t *testing.T) {
		app, router, conf := NewApiTest()
		GetVideo(router)
		r := PerformRequest(app, "GET", "/api/v1/videos/ocad9168fa6acc5c5c2965ddf6ec465ca42fd818/"+conf.PreviewToken()+"/mp4")
		assert.Equal(t, http.StatusOK, r.Code)
	})
}
