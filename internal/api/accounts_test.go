package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"

	"github.com/photoprism/photoprism/internal/i18n"
)

func TestGetAccount(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		app, router, _ := NewApiTest()
		GetAccount(router)
		r := PerformRequest(app, "GET", "/api/v1/accounts/1000000")
		val := gjson.Get(r.Body.String(), "AccName")
		assert.Equal(t, "Test Account", val.String())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("account not found", func(t *testing.T) {
		app, router, _ := NewApiTest()
		GetAccount(router)
		r := PerformRequest(app, "GET", "/api/v1/accounts/999000")
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, i18n.Msg(i18n.ErrAccountNotFound), val.String())
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
}

func TestGetAccountFolders(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		app, router, _ := NewApiTest()
		GetAccountFolders(router)
		r := PerformRequest(app, "GET", "/api/v1/accounts/1000000/folders")
		count := gjson.Get(r.Body.String(), "#")
		assert.LessOrEqual(t, int64(2), count.Int())
		val := gjson.Get(r.Body.String(), "0.abs")
		assert.Equal(t, "/Photos", val.String())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("account not found", func(t *testing.T) {
		app, router, _ := NewApiTest()
		GetAccountFolders(router)
		r := PerformRequest(app, "GET", "/api/v1/accounts/999000/folders")
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, i18n.Msg(i18n.ErrAccountNotFound), val.String())
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
}

func TestShareWithAccount(t *testing.T) {
	t.Run("invalid request", func(t *testing.T) {
		app, router, _ := NewApiTest()
		ShareWithAccount(router)
		r := PerformRequest(app, "POST", "/api/v1/accounts/1000000/share")
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, i18n.Msg(i18n.ErrBadRequest), val.String())
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("account not found", func(t *testing.T) {
		app, router, _ := NewApiTest()
		ShareWithAccount(router)
		r := PerformRequest(app, "POST", "/api/v1/accounts/999000/share")
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, i18n.Msg(i18n.ErrAccountNotFound), val.String())
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
}

func TestCreateAccount(t *testing.T) {
	t.Run("invalid request", func(t *testing.T) {
		app, router, _ := NewApiTest()
		CreateAccount(router)
		r := PerformRequest(app, "POST", "/api/v1/accounts")
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, i18n.Msg(i18n.ErrBadRequest), val.String())
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("could not connect", func(t *testing.T) {
		app, router, _ := NewApiTest()
		CreateAccount(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/accounts", `{"AccName": "CreateTest1", "AccOwner": "Test", "AccUrl": "http://webdav123/", "AccType": "webdav",
"AccKey": "123", "AccUser": "testuser", "AccPass": "testpasswd", "AccError": "", "AccShare": false, "AccSync": false, "RetryLimit": 3, "SharePath": "", "ShareSize": "", "ShareExpires": 0,
"SyncPath": "", "SyncInterval": 3, "SyncUpload": false, "SyncDownload": false, "SyncFilenames": false, "SyncRaw": false}`)
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, i18n.Msg(i18n.ErrConnectionFailed), val.String())
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("successful request", func(t *testing.T) {
		app, router, _ := NewApiTest()
		CreateAccount(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/accounts", `{"AccName": "CreateTest", "AccOwner": "Test", "AccUrl": "http://dummy-webdav/", "AccType": "webdav",
"AccKey": "123", "AccUser": "admin", "AccPass": "photoprism", "AccError": "", "AccShare": false, "AccSync": false, "RetryLimit": 3, "SharePath": "", "ShareSize": "", "ShareExpires": 0,
"SyncPath": "", "SyncInterval": 3, "SyncUpload": false, "SyncDownload": false, "SyncFilenames": false, "SyncRaw": false}`)
		val := gjson.Get(r.Body.String(), "AccOwner")
		assert.Equal(t, "Test", val.String())
		assert.Equal(t, http.StatusOK, r.Code)
	})
}

func TestUpdateAccount(t *testing.T) {
	app, router, _ := NewApiTest()
	CreateAccount(router)
	r := PerformRequestWithBody(app, "POST", "/api/v1/accounts", `{"AccName": "CreateTest3", "AccOwner": "TestUpdate", "AccUrl": "http://dummy-webdav/", "AccType": "webdav",
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

	t.Run("successful request", func(t *testing.T) {
		app, router, _ := NewApiTest()
		UpdateAccount(router)
		r := PerformRequestWithBody(app, "PUT", "/api/v1/accounts/"+id, `{"AccName": "CreateTestUpdated", "AccOwner": "TestUpdated123", "SyncInterval": 9}`)
		val := gjson.Get(r.Body.String(), "AccOwner")
		assert.Equal(t, "TestUpdated123", val.String())
		val2 := gjson.Get(r.Body.String(), "SyncInterval")
		assert.Equal(t, int64(9), val2.Int())
		val3 := gjson.Get(r.Body.String(), "AccName")
		assert.Equal(t, "CreateTestUpdated", val3.String())
		assert.Equal(t, http.StatusOK, r.Code)
	})

	t.Run("not found", func(t *testing.T) {
		app, router, _ := NewApiTest()
		UpdateAccount(router)
		r := PerformRequestWithBody(app, "PUT", "/api/v1/accounts/xxx", `{"AccName": "CreateTestUpdated", "AccOwner": "TestUpdated123", "SyncInterval": 9}`)
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, i18n.Msg(i18n.ErrAccountNotFound), val.String())
		assert.Equal(t, http.StatusNotFound, r.Code)
	})

	t.Run("changes could not be saved", func(t *testing.T) {
		app, router, _ := NewApiTest()
		UpdateAccount(router)
		r := PerformRequestWithBody(app, "PUT", "/api/v1/accounts/"+id, `{"AccName": 6, "AccOwner": "TestUpdated123", "SyncInterval": 9, "AccUrl": "https:xxx.com"}`)
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, i18n.Msg(i18n.ErrBadRequest), val.String())
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
}

func TestDeleteAccount(t *testing.T) {
	app, router, _ := NewApiTest()
	CreateAccount(router)
	r := PerformRequestWithBody(app, "POST", "/api/v1/accounts", `{"AccName": "DeleteTest", "AccOwner": "TestDelete", "AccUrl": "http://dummy-webdav/", "AccType": "webdav",
"AccKey": "123", "AccUser": "admin", "AccPass": "photoprism", "AccError": "", "AccShare": false, "AccSync": false, "RetryLimit": 3, "SharePath": "", "ShareSize": "", "ShareExpires": 0,
"SyncPath": "", "SyncInterval": 5, "SyncUpload": false, "SyncDownload": false, "SyncFilenames": false, "SyncRaw": false}`)
	assert.Equal(t, http.StatusOK, r.Code)
	id := gjson.Get(r.Body.String(), "ID").String()

	t.Run("successful request", func(t *testing.T) {
		app, router, _ := NewApiTest()
		DeleteAccount(router)
		r := PerformRequest(app, "DELETE", "/api/v1/accounts/"+id)
		val := gjson.Get(r.Body.String(), "AccOwner")
		assert.Equal(t, "TestDelete", val.String())
		assert.Equal(t, http.StatusOK, r.Code)
		GetAccount(router)
		r2 := PerformRequest(app, "GET", "/api/v1/accounts/"+id)
		val2 := gjson.Get(r2.Body.String(), "error")
		assert.Equal(t, i18n.Msg(i18n.ErrAccountNotFound), val2.String())
		assert.Equal(t, http.StatusNotFound, r2.Code)
	})

	t.Run("not found", func(t *testing.T) {
		app, router, _ := NewApiTest()
		DeleteAccount(router)
		r := PerformRequest(app, "DELETE", "/api/v1/accounts/xxx")
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, i18n.Msg(i18n.ErrAccountNotFound), val.String())
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
}
