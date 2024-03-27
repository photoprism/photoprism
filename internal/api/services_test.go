package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"

	"github.com/photoprism/photoprism/pkg/i18n"
)

func TestGetService(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		app, router, _ := NewApiTest()
		GetService(router)
		r := PerformRequest(app, "GET", "/api/v1/services/1000000")
		val := gjson.Get(r.Body.String(), "AccName")
		assert.Equal(t, "Test Account", val.String())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("NotFound", func(t *testing.T) {
		app, router, _ := NewApiTest()
		GetService(router)
		r := PerformRequest(app, "GET", "/api/v1/services/999000")
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, i18n.Msg(i18n.ErrAccountNotFound), val.String())
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
}

func TestGetServiceFolders(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		app, router, _ := NewApiTest()
		GetServiceFolders(router)
		r := PerformRequest(app, "GET", "/api/v1/services/1000000/folders")
		count := gjson.Get(r.Body.String(), "#")
		assert.LessOrEqual(t, int64(2), count.Int())
		val := gjson.Get(r.Body.String(), "0.abs")
		assert.Equal(t, "/Photos", val.String())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("NotFound", func(t *testing.T) {
		app, router, _ := NewApiTest()
		GetServiceFolders(router)
		r := PerformRequest(app, "GET", "/api/v1/services/999000/folders")
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, i18n.Msg(i18n.ErrAccountNotFound), val.String())
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
}

func TestCreateService(t *testing.T) {
	t.Run("BadRequest", func(t *testing.T) {
		app, router, _ := NewApiTest()
		AddService(router)
		r := PerformRequest(app, "POST", "/api/v1/services")
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, i18n.Msg(i18n.ErrBadRequest), val.String())
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("ConnectFailed", func(t *testing.T) {
		app, router, _ := NewApiTest()
		AddService(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/services", `{"AccName": "CreateTest1", "AccOwner": "Test", "AccUrl": "http://webdav123/", "AccType": "webdav",
"AccKey": "123", "AccUser": "testuser", "AccPass": "testpasswd", "AccError": "", "AccShare": false, "AccSync": false, "RetryLimit": 3, "SharePath": "", "ShareSize": "", "ShareExpires": 0,
"SyncPath": "", "SyncInterval": 3, "SyncUpload": false, "SyncDownload": false, "SyncFilenames": false, "SyncRaw": false}`)
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, i18n.Msg(i18n.ErrConnectionFailed), val.String())
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("Ok", func(t *testing.T) {
		app, router, _ := NewApiTest()
		AddService(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/services", `{"AccName": "CreateTest", "AccOwner": "Test", "AccUrl": "http://dummy-webdav/", "AccType": "webdav",
"AccKey": "123", "AccUser": "admin", "AccPass": "photoprism", "AccError": "", "AccShare": false, "AccSync": false, "RetryLimit": 3, "SharePath": "", "ShareSize": "", "ShareExpires": 0,
"SyncPath": "", "SyncInterval": 3, "SyncUpload": false, "SyncDownload": false, "SyncFilenames": false, "SyncRaw": false}`)
		val := gjson.Get(r.Body.String(), "AccOwner")
		assert.Equal(t, "Test", val.String())
		assert.Equal(t, http.StatusOK, r.Code)
	})
}

func TestUpdateService(t *testing.T) {
	app, router, _ := NewApiTest()
	AddService(router)
	r := PerformRequestWithBody(app, "POST", "/api/v1/services", `{"AccName": "CreateTest3", "AccOwner": "TestUpdate", "AccUrl": "http://dummy-webdav/", "AccType": "webdav",
"AccKey": "123", "AccUser": "admin", "AccPass": "photoprism", "AccError": "", "AccShare": false, "AccSync": false, "RetryLimit": 3, "SharePath": "", "ShareSize": "", "ShareExpires": 0,
"SyncPath": "", "SyncInterval": 5, "SyncUpload": false, "SyncDownload": false, "SyncFilenames": false, "SyncRaw": false}`)
	val := gjson.Get(r.Body.String(), "AccOwner")
	assert.Equal(t, "TestUpdate", val.String())
	val2 := gjson.Get(r.Body.String(), "SyncInterval")
	assert.Equal(t, int64(5), val2.Int())
	val3 := gjson.Get(r.Body.String(), "AccName")
	assert.Equal(t, "Dummy-Webdav", val3.String())
	assert.Equal(t, http.StatusOK, r.Code)
	id := gjson.Get(r.Body.String(), "ID").String()

	t.Run("Ok", func(t *testing.T) {
		app, router, _ := NewApiTest()
		UpdateService(router)
		r := PerformRequestWithBody(app, "PUT", "/api/v1/services/"+id, `{"AccName": "CreateTestUpdated", "AccOwner": "TestUpdated123", "SyncInterval": 9}`)
		val := gjson.Get(r.Body.String(), "AccOwner")
		assert.Equal(t, "TestUpdated123", val.String())
		val2 := gjson.Get(r.Body.String(), "SyncInterval")
		assert.Equal(t, int64(9), val2.Int())
		val3 := gjson.Get(r.Body.String(), "AccName")
		assert.Equal(t, "CreateTestUpdated", val3.String())
		assert.Equal(t, http.StatusOK, r.Code)
	})

	t.Run("NotFound", func(t *testing.T) {
		app, router, _ := NewApiTest()
		UpdateService(router)
		r := PerformRequestWithBody(app, "PUT", "/api/v1/services/xxx", `{"AccName": "CreateTestUpdated", "AccOwner": "TestUpdated123", "SyncInterval": 9}`)
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, i18n.Msg(i18n.ErrAccountNotFound), val.String())
		assert.Equal(t, http.StatusNotFound, r.Code)
	})

	t.Run("SaveFailed", func(t *testing.T) {
		app, router, _ := NewApiTest()
		UpdateService(router)
		r := PerformRequestWithBody(app, "PUT", "/api/v1/services/"+id, `{"AccName": 6, "AccOwner": "TestUpdated123", "SyncInterval": 9, "AccUrl": "https:xxx.com"}`)
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, i18n.Msg(i18n.ErrBadRequest), val.String())
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
}

func TestDeleteService(t *testing.T) {
	app, router, _ := NewApiTest()
	AddService(router)
	r := PerformRequestWithBody(app, "POST", "/api/v1/services", `{"AccName": "DeleteTest", "AccOwner": "TestDelete", "AccUrl": "http://dummy-webdav/", "AccType": "webdav",
"AccKey": "123", "AccUser": "admin", "AccPass": "photoprism", "AccError": "", "AccShare": false, "AccSync": false, "RetryLimit": 3, "SharePath": "", "ShareSize": "", "ShareExpires": 0,
"SyncPath": "", "SyncInterval": 5, "SyncUpload": false, "SyncDownload": false, "SyncFilenames": false, "SyncRaw": false}`)
	assert.Equal(t, http.StatusOK, r.Code)
	id := gjson.Get(r.Body.String(), "ID").String()

	t.Run("Ok", func(t *testing.T) {
		app, router, _ := NewApiTest()
		DeleteService(router)
		r := PerformRequest(app, "DELETE", "/api/v1/services/"+id)
		val := gjson.Get(r.Body.String(), "AccOwner")
		assert.Equal(t, "TestDelete", val.String())
		assert.Equal(t, http.StatusOK, r.Code)
		GetService(router)
		r2 := PerformRequest(app, "GET", "/api/v1/services/"+id)
		val2 := gjson.Get(r2.Body.String(), "error")
		assert.Equal(t, i18n.Msg(i18n.ErrAccountNotFound), val2.String())
		assert.Equal(t, http.StatusNotFound, r2.Code)
	})

	t.Run("NotFound", func(t *testing.T) {
		app, router, _ := NewApiTest()
		DeleteService(router)
		r := PerformRequest(app, "DELETE", "/api/v1/services/xxx")
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, i18n.Msg(i18n.ErrAccountNotFound), val.String())
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
}
