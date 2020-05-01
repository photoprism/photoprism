package api

import (
	"github.com/tidwall/gjson"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAccounts(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		app, router, conf := NewApiTest()
		GetAccounts(router, conf)
		r := PerformRequest(app, "GET", "/api/v1/accounts?count=10")
		val := gjson.Get(r.Body.String(), "#(AccName=\"Test Account\").AccURL")
		len := gjson.Get(r.Body.String(), "#")
		assert.LessOrEqual(t, int64(1), len.Int())
		assert.Equal(t, "http://webdav-dummy/", val.String())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("invalid request", func(t *testing.T) {
		app, router, conf := NewApiTest()
		GetAccounts(router, conf)
		r := PerformRequest(app, "GET", "/api/v1/accounts?xxx=10")
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
}

func TestGetAccount(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		app, router, conf := NewApiTest()
		GetAccount(router, conf)
		r := PerformRequest(app, "GET", "/api/v1/accounts/1000000")
		val := gjson.Get(r.Body.String(), "AccName")
		assert.Equal(t, "Test Account", val.String())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("account not found", func(t *testing.T) {
		app, router, conf := NewApiTest()
		GetAccount(router, conf)
		r := PerformRequest(app, "GET", "/api/v1/accounts/999000")
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, "Account not found", val.String())
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
}

func TestGetAccountDirs(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		app, router, conf := NewApiTest()
		GetAccountDirs(router, conf)
		r := PerformRequest(app, "GET", "/api/v1/accounts/1000000/dirs")
		len := gjson.Get(r.Body.String(), "#")
		assert.LessOrEqual(t, int64(2), len.Int())
		val := gjson.Get(r.Body.String(), "0.abs")
		assert.Equal(t, "/Photos", val.String())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("account not found", func(t *testing.T) {
		app, router, conf := NewApiTest()
		GetAccountDirs(router, conf)
		r := PerformRequest(app, "GET", "/api/v1/accounts/999000/dirs")
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, "Account not found", val.String())
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
}

func TestShareWithAccount(t *testing.T) {
	t.Run("invalid request", func(t *testing.T) {
		app, router, conf := NewApiTest()
		ShareWithAccount(router, conf)
		r := PerformRequest(app, "POST", "/api/v1/accounts/1000000/share")
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, "Invalid request", val.String())
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("account not found", func(t *testing.T) {
		app, router, conf := NewApiTest()
		ShareWithAccount(router, conf)
		r := PerformRequest(app, "POST", "/api/v1/accounts/999000/share")
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, "Account not found", val.String())
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
}

func TestCreateAccount(t *testing.T) {
	t.Run("invalid request", func(t *testing.T) {
		app, router, conf := NewApiTest()
		CreateAccount(router, conf)
		r := PerformRequest(app, "POST", "/api/v1/accounts")
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, "Invalid request", val.String())
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("could not connect", func(t *testing.T) {
		app, router, conf := NewApiTest()
		CreateAccount(router, conf)
		r := PerformRequestWithBody(app, "POST", "/api/v1/accounts", `{"AccName": "CreateTest1", "AccOwner": "Test", "AccUrl": "http://webdav123/", "AccType": "webdav",
"AccKey": "123", "AccUser": "testuser", "AccPass": "testpasswd", "AccError": "", "AccShare": false, "AccSync": false, "RetryLimit": 3, "SharePath": "", "ShareSize": "", "ShareExpires": 0,
"SyncPath": "", "SyncInterval": 3, "SyncUpload": false, "SyncDownload": false, "SyncFilenames": false, "SyncRaw": false}`)
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, "Could not connect", val.String())
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("successful request", func(t *testing.T) {
		app, router, conf := NewApiTest()
		CreateAccount(router, conf)
		r := PerformRequestWithBody(app, "POST", "/api/v1/accounts", `{"AccName": "CreateTest", "AccOwner": "Test", "AccUrl": "http://webdav-dummy/", "AccType": "webdav",
"AccKey": "123", "AccUser": "admin", "AccPass": "photoprism", "AccError": "", "AccShare": false, "AccSync": false, "RetryLimit": 3, "SharePath": "", "ShareSize": "", "ShareExpires": 0,
"SyncPath": "", "SyncInterval": 3, "SyncUpload": false, "SyncDownload": false, "SyncFilenames": false, "SyncRaw": false}`)
		val := gjson.Get(r.Body.String(), "AccOwner")
		assert.Equal(t, "Test", val.String())
		assert.Equal(t, http.StatusOK, r.Code)
	})
}

func TestUpdateAccount(t *testing.T) {
	app, router, conf := NewApiTest()
	CreateAccount(router, conf)
	r := PerformRequestWithBody(app, "POST", "/api/v1/accounts", `{"AccName": "CreateTest3", "AccOwner": "TestUpdate", "AccUrl": "http://webdav-dummy/", "AccType": "webdav",
"AccKey": "123", "AccUser": "admin", "AccPass": "photoprism", "AccError": "", "AccShare": false, "AccSync": false, "RetryLimit": 3, "SharePath": "", "ShareSize": "", "ShareExpires": 0,
"SyncPath": "", "SyncInterval": 5, "SyncUpload": false, "SyncDownload": false, "SyncFilenames": false, "SyncRaw": false}`)
	val := gjson.Get(r.Body.String(), "AccOwner")
	assert.Equal(t, "TestUpdate", val.String())
	val2 := gjson.Get(r.Body.String(), "SyncInterval")
	assert.Equal(t, int64(5), val2.Int())
	val3 := gjson.Get(r.Body.String(), "AccName")
	assert.Equal(t, "Webdav-Dummy", val3.String())
	assert.Equal(t, http.StatusOK, r.Code)
	id := gjson.Get(r.Body.String(), "ID").String()

	t.Run("successful request", func(t *testing.T) {
		app, router, conf := NewApiTest()
		UpdateAccount(router, conf)
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
		app, router, conf := NewApiTest()
		UpdateAccount(router, conf)
		r := PerformRequestWithBody(app, "PUT", "/api/v1/accounts/xxx", `{"AccName": "CreateTestUpdated", "AccOwner": "TestUpdated123", "SyncInterval": 9}`)
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, "Photo not found", val.String())
		assert.Equal(t, http.StatusNotFound, r.Code)
	})

	t.Run("changes could not be saved", func(t *testing.T) {
		app, router, conf := NewApiTest()
		UpdateAccount(router, conf)
		r := PerformRequestWithBody(app, "PUT", "/api/v1/accounts/"+id, `{"AccName": 6, "AccOwner": "TestUpdated123", "SyncInterval": 9, "AccUrl": "https:xxx.com"}`)
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, "Changes could not be saved", val.String())
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
}

func TestDeleteAccount(t *testing.T) {
	app, router, conf := NewApiTest()
	CreateAccount(router, conf)
	r := PerformRequestWithBody(app, "POST", "/api/v1/accounts", `{"AccName": "DeleteTest", "AccOwner": "TestDelete", "AccUrl": "http://webdav-dummy/", "AccType": "webdav",
"AccKey": "123", "AccUser": "admin", "AccPass": "photoprism", "AccError": "", "AccShare": false, "AccSync": false, "RetryLimit": 3, "SharePath": "", "ShareSize": "", "ShareExpires": 0,
"SyncPath": "", "SyncInterval": 5, "SyncUpload": false, "SyncDownload": false, "SyncFilenames": false, "SyncRaw": false}`)
	assert.Equal(t, http.StatusOK, r.Code)
	id := gjson.Get(r.Body.String(), "ID").String()

	t.Run("successful request", func(t *testing.T) {
		app, router, conf := NewApiTest()
		DeleteAccount(router, conf)
		r := PerformRequest(app, "DELETE", "/api/v1/accounts/"+id)
		val := gjson.Get(r.Body.String(), "AccOwner")
		assert.Equal(t, "TestDelete", val.String())
		assert.Equal(t, http.StatusOK, r.Code)
		GetAccount(router, conf)
		r2 := PerformRequest(app, "GET", "/api/v1/accounts/"+id)
		val2 := gjson.Get(r2.Body.String(), "error")
		assert.Equal(t, "Account not found", val2.String())
		assert.Equal(t, http.StatusNotFound, r2.Code)
	})

	t.Run("not found", func(t *testing.T) {
		app, router, conf := NewApiTest()
		DeleteAccount(router, conf)
		r := PerformRequest(app, "DELETE", "/api/v1/accounts/xxx")
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, "Account not found", val.String())
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
}
