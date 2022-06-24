package api

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/photoprism/photoprism/pkg/clean"

	"github.com/stretchr/testify/assert"
)

func TestGetVideo(t *testing.T) {
	t.Run("ContentTypeAvc", func(t *testing.T) {
		assert.Equal(t, ContentTypeAvc, fmt.Sprintf("%s; codecs=\"%s\"", "video/mp4", clean.Codec("avc1")))
	})

	t.Run("invalid hash", func(t *testing.T) {
		app, router, conf := NewApiTest()
		GetVideo(router)
		r := PerformRequest(app, "GET", "/api/v1/videos/xxx/"+conf.PreviewToken()+"/mp4")
		assert.Equal(t, http.StatusOK, r.Code)
	})

	t.Run("invalid type", func(t *testing.T) {
		app, router, conf := NewApiTest()
		GetVideo(router)
		r := PerformRequest(app, "GET", "/api/v1/videos/acad9168fa6acc5c5c2965ddf6ec465ca42fd831/"+conf.PreviewToken()+"/xxx")
		assert.Equal(t, http.StatusOK, r.Code)
	})

	t.Run("file for video not found", func(t *testing.T) {
		app, router, conf := NewApiTest()
		GetVideo(router)
		r := PerformRequest(app, "GET", "/api/v1/videos/acad9168fa6acc5c5c2965ddf6ec465ca42fd831/"+conf.PreviewToken()+"/mp4")
		assert.Equal(t, http.StatusOK, r.Code)
	})

	t.Run("file with error", func(t *testing.T) {
		app, router, conf := NewApiTest()
		GetVideo(router)
		r := PerformRequest(app, "GET", "/api/v1/videos/acad9168fa6acc5c5c2965ddf6ec465ca42fd832/"+conf.PreviewToken()+"/mp4")
		assert.Equal(t, http.StatusOK, r.Code)
	})

	t.Run("invalid token", func(t *testing.T) {
		app, router, _ := NewApiTest()
		GetVideo(router)
		r := PerformRequest(app, "GET", "/api/v1/videos/acad9168fa6acc5c5c2965ddf6ec465ca42fd832/xxx/mp4")
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
	t.Run("no video file", func(t *testing.T) {
		app, router, conf := NewApiTest()
		GetVideo(router)
		r := PerformRequest(app, "GET", "/api/v1/videos/ocad9168fa6acc5c5c2965ddf6ec465ca42fd818/"+conf.PreviewToken()+"/mp4")
		assert.Equal(t, http.StatusOK, r.Code)
	})
}
