package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetSvg(t *testing.T) {
	t.Run("photo", func(t *testing.T) {
		app, router, conf := NewApiTest()
		t.Log(conf)
		GetSvg(router)
		r := PerformRequest(app, "GET", "/api/v1/svg/photo")
		assert.Equal(t, photoIconSvg, r.Body.Bytes())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("raw", func(t *testing.T) {
		app, router, conf := NewApiTest()
		t.Log(conf)
		GetSvg(router)
		r := PerformRequest(app, "GET", "/api/v1/svg/raw")
		assert.Equal(t, rawIconSvg, r.Body.Bytes())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("file", func(t *testing.T) {
		app, router, conf := NewApiTest()
		t.Log(conf)
		GetSvg(router)
		r := PerformRequest(app, "GET", "/api/v1/svg/file")
		assert.Equal(t, fileIconSvg, r.Body.Bytes())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("video", func(t *testing.T) {
		app, router, conf := NewApiTest()
		t.Log(conf)
		GetSvg(router)
		r := PerformRequest(app, "GET", "/api/v1/svg/video")
		assert.Equal(t, videoIconSvg, r.Body.Bytes())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("label", func(t *testing.T) {
		app, router, conf := NewApiTest()
		t.Log(conf)
		GetSvg(router)
		r := PerformRequest(app, "GET", "/api/v1/svg/label")
		assert.Equal(t, labelIconSvg, r.Body.Bytes())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("album", func(t *testing.T) {
		app, router, conf := NewApiTest()
		t.Log(conf)
		GetSvg(router)
		r := PerformRequest(app, "GET", "/api/v1/svg/album")
		assert.Equal(t, albumIconSvg, r.Body.Bytes())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("folder", func(t *testing.T) {
		app, router, conf := NewApiTest()
		t.Log(conf)
		GetSvg(router)
		r := PerformRequest(app, "GET", "/api/v1/svg/folder")
		assert.Equal(t, folderIconSvg, r.Body.Bytes())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("broken", func(t *testing.T) {
		app, router, conf := NewApiTest()
		t.Log(conf)
		GetSvg(router)
		r := PerformRequest(app, "GET", "/api/v1/svg/broken")
		assert.Equal(t, brokenIconSvg, r.Body.Bytes())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("uncached", func(t *testing.T) {
		app, router, conf := NewApiTest()
		t.Log(conf)
		GetSvg(router)
		r := PerformRequest(app, "GET", "/api/v1/svg/uncached")
		assert.Equal(t, uncachedIconSvg, r.Body.Bytes())
		assert.Equal(t, http.StatusOK, r.Code)
	})
}
