package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/i18n"
)

func TestGetPhoto(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, router, _ := NewApiTest()
		GetPhoto(router)
		r := PerformRequest(app, "GET", "/api/v1/photos/ps6sg6be2lvl0yh7")
		assert.Equal(t, http.StatusOK, r.Code)
		val := gjson.Get(r.Body.String(), "Iso")
		assert.Equal(t, "200", val.String())
	})
	t.Run("AliceAppPassword", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		GetPhoto(router)
		r := AuthenticatedRequest(app, "GET", "/api/v1/photos/ps6sg6be2lvl0yh7", "X3B6IU-hfeLG5-HpVxkT-ctCY3M")
		assert.Equal(t, http.StatusOK, r.Code)
		val := gjson.Get(r.Body.String(), "Iso")
		assert.Equal(t, "200", val.String())
	})
	t.Run("AliceAppPasswordWebdav", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		GetPhoto(router)
		r := AuthenticatedRequest(app, "GET", "/api/v1/photos/ps6sg6be2lvl0yh7", "v2wS72-OkqEzm-MQ63Z2-TEhU0w")
		assert.Equal(t, http.StatusForbidden, r.Code)
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, "Permission denied", val.String())
	})
	t.Run("AccessToken", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		GetPhoto(router)
		r := AuthenticatedRequest(app, "GET", "/api/v1/photos/ps6sg6be2lvl0yh7", "8e154d323800393faf5177ce7392116feebbf674e6c2d39e")
		assert.Equal(t, http.StatusOK, r.Code)
		val := gjson.Get(r.Body.String(), "Iso")
		assert.Equal(t, "200", val.String())
	})
	t.Run("InvalidAppPassword", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		GetPhoto(router)
		r := AuthenticatedRequest(app, "GET", "/api/v1/photos/ps6sg6be2lvl0yh7", "69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7xxx")
		assert.Equal(t, http.StatusUnauthorized, r.Code)
		val := gjson.Get(r.Body.String(), "Iso")
		assert.Equal(t, "", val.String())
	})
	t.Run("NotFound", func(t *testing.T) {
		app, router, _ := NewApiTest()
		GetPhoto(router)
		r := PerformRequest(app, "GET", "/api/v1/photos/xxx")
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
}

func TestUpdatePhoto(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, router, _ := NewApiTest()
		UpdatePhoto(router)
		r := PerformRequestWithBody(app, "PUT", "/api/v1/photos/ps6sg6be2lvl0y13", `{"Title": "Updated01", "Country": "de"}`)
		val := gjson.Get(r.Body.String(), "Title")
		assert.Equal(t, "Updated01", val.String())
		val2 := gjson.Get(r.Body.String(), "Country")
		assert.Equal(t, "de", val2.String())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("BadRequest", func(t *testing.T) {
		app, router, _ := NewApiTest()
		UpdatePhoto(router)
		r := PerformRequestWithBody(app, "PUT", "/api/v1/photos/ps6sg6be2lvl0y13", `{"Name": "Updated01", "Country": 123}`)
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("NotFound", func(t *testing.T) {
		app, router, _ := NewApiTest()
		UpdatePhoto(router)
		r := PerformRequestWithBody(app, "PUT", "/api/v1/photos/xxx", `{"Name": "Updated01", "Country": "de"}`)
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, i18n.Msg(i18n.ErrEntityNotFound), val.String())
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
}

func TestGetPhotoDownload(t *testing.T) {
	t.Run("OriginalMissing", func(t *testing.T) {
		app, router, conf := NewApiTest()
		GetPhotoDownload(router)
		r := PerformRequest(app, "GET", "/api/v1/photos/ps6sg6be2lvl0yh7/dl?t="+conf.DownloadToken())
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("NotFound", func(t *testing.T) {
		app, router, conf := NewApiTest()
		GetPhotoDownload(router)
		r := PerformRequest(app, "GET", "/api/v1/photos/xxx/dl?t="+conf.DownloadToken())
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("InvalidToken", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		GetPhotoDownload(router)
		r := PerformRequest(app, "GET", "/api/v1/photos/ps6sg6be2lvl0yh7/dl?t=xxx")
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
}

func TestLikePhoto(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, router, _ := NewApiTest()
		LikePhoto(router)
		r := PerformRequest(app, "POST", "/api/v1/photos/ps6sg6be2lvl0yh9/like")
		assert.Equal(t, http.StatusOK, r.Code)
		GetPhoto(router)
		r2 := PerformRequest(app, "GET", "/api/v1/photos/ps6sg6be2lvl0yh9")
		val := gjson.Get(r2.Body.String(), "Favorite")
		assert.Equal(t, "true", val.String())
	})
	t.Run("NotFound", func(t *testing.T) {
		app, router, _ := NewApiTest()
		LikePhoto(router)
		r := PerformRequest(app, "POST", "/api/v1/photos/xxx/like")
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
}

func TestDislikePhoto(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, router, _ := NewApiTest()
		DislikePhoto(router)
		r := PerformRequest(app, "DELETE", "/api/v1/photos/ps6sg6be2lvl0yh8/like")
		assert.Equal(t, http.StatusOK, r.Code)
		GetPhoto(router)
		r2 := PerformRequest(app, "GET", "/api/v1/photos/ps6sg6be2lvl0yh8")
		val := gjson.Get(r2.Body.String(), "Favorite")
		assert.Equal(t, "false", val.String())
	})
	t.Run("NotFound", func(t *testing.T) {
		app, router, _ := NewApiTest()
		DislikePhoto(router)
		r := PerformRequest(app, "DELETE", "/api/v1/photos/xxx/like")
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
}

func TestPhotoPrimary(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, router, _ := NewApiTest()
		PhotoPrimary(router)
		r := PerformRequest(app, "POST", "/api/v1/photos/ps6sg6be2lvl0yh8/files/fs6sg6bw45bn0003/primary")
		assert.Equal(t, http.StatusOK, r.Code)
		GetFile(router)
		r2 := PerformRequest(app, "GET", "/api/v1/files/ocad9168fa6acc5c5c2965ddf6ec465ca42fd818")
		val := gjson.Get(r2.Body.String(), "Primary")
		assert.Equal(t, "true", val.String())
		r3 := PerformRequest(app, "GET", "/api/v1/files/3cad9168fa6acc5c5c2965ddf6ec465ca42fd818")
		val2 := gjson.Get(r3.Body.String(), "Primary")
		assert.Equal(t, "false", val2.String())
	})
	t.Run("NotFound", func(t *testing.T) {
		app, router, _ := NewApiTest()
		PhotoPrimary(router)
		r := PerformRequest(app, "POST", "/api/v1/photos/xxx/files/fs6sg6bw45bnlqdw/primary")
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, i18n.Msg(i18n.ErrEntityNotFound), val.String())
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
}

func TestGetPhotoYaml(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, router, _ := NewApiTest()
		GetPhotoYaml(router)
		r := PerformRequest(app, "GET", "/api/v1/photos/ps6sg6be2lvl0yh7/yaml")
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("NotFound", func(t *testing.T) {
		app, router, _ := NewApiTest()
		GetPhotoYaml(router)
		r := PerformRequest(app, "GET", "/api/v1/photos/xxx/yaml")
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
}

func TestApprovePhoto(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, router, _ := NewApiTest()
		GetPhoto(router)
		r3 := PerformRequest(app, "GET", "/api/v1/photos/ps6sg6bexxvl0y20")
		val2 := gjson.Get(r3.Body.String(), "Quality")
		assert.Equal(t, "1", val2.String())
		ApprovePhoto(router)
		r := PerformRequest(app, "POST", "/api/v1/photos/ps6sg6bexxvl0y20/approve")
		assert.Equal(t, http.StatusOK, r.Code)
		r2 := PerformRequest(app, "GET", "/api/v1/photos/ps6sg6bexxvl0y20")
		val := gjson.Get(r2.Body.String(), "Quality")
		assert.Equal(t, "3", val.String())
	})
	t.Run("NotFound", func(t *testing.T) {
		app, router, _ := NewApiTest()
		ApprovePhoto(router)
		r := PerformRequest(app, "POST", "/api/v1/photos/xxx/approve")
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
}
