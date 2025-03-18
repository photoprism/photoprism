package api

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/media/http/header"
	"github.com/photoprism/photoprism/pkg/media/video"
)

func TestGetVideo(t *testing.T) {
	t.Run("ContentTypeAVC", func(t *testing.T) {
		assert.Equal(t, header.ContentTypeMp4AvcMain, clean.ContentType("video/mp4; codecs=\"avc1\""))
		mimeType := fmt.Sprintf("video/mp4; codecs=\"%s\"", clean.Codec("avc1"))
		assert.Equal(t, header.ContentTypeMp4AvcMain, video.ContentType(mimeType, "mp4", "avc1", false))
	})

	t.Run("NoHash", func(t *testing.T) {
		app, router, conf := NewApiTest()
		GetVideo(router)
		r := PerformRequest(app, "GET", "/api/v1/videos//"+conf.PreviewToken()+"/mp4")
		assert.Equal(t, http.StatusOK, r.Code)
	})

	t.Run("InvalidHash", func(t *testing.T) {
		app, router, conf := NewApiTest()
		GetVideo(router)
		r := PerformRequest(app, "GET", "/api/v1/videos/acad9168fa6/"+conf.PreviewToken()+"/mp4")
		assert.Equal(t, http.StatusOK, r.Code)
	})

	t.Run("NoType", func(t *testing.T) {
		app, router, conf := NewApiTest()
		GetVideo(router)
		r := PerformRequest(app, "GET", "/api/v1/videos/acad9168fa6acc5c5c2965ddf6ec465ca42fd831/"+conf.PreviewToken()+"/")
		assert.Equal(t, http.StatusMovedPermanently, r.Code)
	})

	t.Run("InvalidType", func(t *testing.T) {
		app, router, conf := NewApiTest()
		GetVideo(router)
		r := PerformRequest(app, "GET", "/api/v1/videos/acad9168fa6acc5c5c2965ddf6ec465ca42fd831/"+conf.PreviewToken()+"/xxx")
		assert.Equal(t, http.StatusBadRequest, r.Code)
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
