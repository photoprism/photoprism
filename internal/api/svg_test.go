package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetSvg(t *testing.T) {
	t.Run("Photo", func(t *testing.T) {
		app, router, conf := NewApiTest()
		t.Log(conf)
		GetSvg(router)
		r := PerformRequest(app, "GET", "/api/v1/svg/photo")
		assert.Equal(t, photoIconSvg, r.Body.Bytes())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("Raw", func(t *testing.T) {
		app, router, conf := NewApiTest()
		t.Log(conf)
		GetSvg(router)
		r := PerformRequest(app, "GET", "/api/v1/svg/raw")
		assert.Equal(t, rawIconSvg, r.Body.Bytes())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("File", func(t *testing.T) {
		app, router, conf := NewApiTest()
		t.Log(conf)
		GetSvg(router)
		r := PerformRequest(app, "GET", "/api/v1/svg/file")
		assert.Equal(t, fileIconSvg, r.Body.Bytes())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("Video", func(t *testing.T) {
		app, router, conf := NewApiTest()
		t.Log(conf)
		GetSvg(router)
		r := PerformRequest(app, "GET", "/api/v1/svg/video")
		assert.Equal(t, videoIconSvg, r.Body.Bytes())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("Label", func(t *testing.T) {
		app, router, conf := NewApiTest()
		t.Log(conf)
		GetSvg(router)
		r := PerformRequest(app, "GET", "/api/v1/svg/label")
		assert.Equal(t, labelIconSvg, r.Body.Bytes())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("Album", func(t *testing.T) {
		app, router, conf := NewApiTest()
		t.Log(conf)
		GetSvg(router)
		r := PerformRequest(app, "GET", "/api/v1/svg/album")
		assert.Equal(t, albumIconSvg, r.Body.Bytes())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("Folder", func(t *testing.T) {
		app, router, conf := NewApiTest()
		t.Log(conf)
		GetSvg(router)
		r := PerformRequest(app, "GET", "/api/v1/svg/folder")
		assert.Equal(t, folderIconSvg, r.Body.Bytes())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("Broken", func(t *testing.T) {
		app, router, conf := NewApiTest()
		t.Log(conf)
		GetSvg(router)
		r := PerformRequest(app, "GET", "/api/v1/svg/broken")
		assert.Equal(t, brokenIconSvg, r.Body.Bytes())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("Uncached", func(t *testing.T) {
		app, router, conf := NewApiTest()
		t.Log(conf)
		GetSvg(router)
		r := PerformRequest(app, "GET", "/api/v1/svg/uncached")
		assert.Equal(t, uncachedIconSvg, r.Body.Bytes())
		assert.Equal(t, http.StatusOK, r.Code)
	})
}
