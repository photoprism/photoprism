package api

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestGetVideo(t *testing.T) {
	t.Run("invalid hash", func(t *testing.T) {
		app, router, conf := NewApiTest()
		GetVideo(router, conf)
		r := PerformRequest(app, "GET", "/api/v1/videos/xxx/mp4")
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("invalid type", func(t *testing.T) {
		app, router, conf := NewApiTest()
		GetVideo(router, conf)
		r := PerformRequest(app, "GET", "/api/v1/videos/acad9168fa6acc5c5c2965ddf6ec465ca42fd831/xxx")
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("file for video not found", func(t *testing.T) {
		app, router, conf := NewApiTest()
		GetVideo(router, conf)
		r := PerformRequest(app, "GET", "/api/v1/videos/acad9168fa6acc5c5c2965ddf6ec465ca42fd831/mp4")
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("file with error", func(t *testing.T) {
		app, router, conf := NewApiTest()
		GetVideo(router, conf)
		r := PerformRequest(app, "GET", "/api/v1/videos/acad9168fa6acc5c5c2965ddf6ec465ca42fd832/mp4")
		assert.Equal(t, http.StatusOK, r.Code)
	})
}
