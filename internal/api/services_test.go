package api

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/i18n"
)

func TestGetService(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
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
	t.Run("Success", func(t *testing.T) {
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
	t.Run("Success", func(t *testing.T) {
		app, router, _ := NewApiTest()
		AddService(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/services", `{"AccName": "CreateTest", "AccOwner": "Test", "AccUrl": "http://dummy-webdav/", "AccType": "webdav",
"AccKey": "123", "AccUser": "admin", "AccPass": "photoprism", "AccError": "", "AccShare": false, "AccSync": false, "RetryLimit": 3, "SharePath": "", "ShareSize": "", "ShareExpires": 0,
"SyncPath": "", "SyncInterval": 3, "SyncUpload": false, "SyncDownload": false, "SyncFilenames": false, "SyncRaw": false}`)
		val := gjson.Get(r.Body.String(), "AccOwner")
		assert.Equal(t, "Test", val.String())
		id := gjson.Get(r.Body.String(), "ID").Int()
		assert.NotZero(t, id)
		assert.Equal(t, fmt.Sprintf("/api/v1/services/%d", id), r.Header().Get("Location"))
		assert.Equal(t, http.StatusCreated, r.Code)
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
	assert.Equal(t, http.StatusCreated, r.Code)
	id := gjson.Get(r.Body.String(), "ID").String()

	t.Run("Success", func(t *testing.T) {
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
	t.Run("FolderTooLong", func(t *testing.T) {
		// A sync folder that leaves no room for file names is rejected, and the account is not changed.
		app, router, _ := NewApiTest()
		UpdateService(router)
		folder := strings.Repeat("/"+strings.Repeat("f", 200), 5) + "/" + strings.Repeat("g", 18)
		require.Len(t, folder, entity.RemoteNameMaxLen)
		r := PerformRequestWithBody(app, "PUT", "/api/v1/services/"+id, fmt.Sprintf(`{"SyncPath": %q}`, folder))
		assert.Equal(t, http.StatusBadRequest, r.Code)
		assert.Equal(t, i18n.Msg(i18n.ErrFolderTooLong), gjson.Get(r.Body.String(), "error").String())
		m, err := query.AccountByID(clean.IdUint(id))
		require.NoError(t, err)
		assert.NotEqual(t, folder, m.SyncPath)
	})
	t.Run("AccError", func(t *testing.T) {
		// The account error cannot be set through the API, only cleared.
		app, router, _ := NewApiTest()
		UpdateService(router)
		m, err := query.AccountByID(clean.IdUint(id))
		require.NoError(t, err)
		require.NoError(t, m.Updates(entity.Values{"acc_error": "worker error"}))
		t.Cleanup(func() { _ = m.Updates(entity.Values{"acc_error": ""}) })

		r := PerformRequestWithBody(app, "PUT", "/api/v1/services/"+id, `{"AccError": "client value"}`)
		assert.Equal(t, http.StatusOK, r.Code, r.Body.String())
		assert.Equal(t, "worker error", gjson.Get(r.Body.String(), "AccError").String())

		r = PerformRequestWithBody(app, "PUT", "/api/v1/services/"+id, `{"AccError": ""}`)
		assert.Equal(t, http.StatusOK, r.Code, r.Body.String())
		m, err = query.AccountByID(clean.IdUint(id))
		require.NoError(t, err)
		assert.Empty(t, m.AccError)
	})
	t.Run("StoredFolderTooLong", func(t *testing.T) {
		// A save that leaves a stored folder unchanged is not refused because of it.
		app, router, _ := NewApiTest()
		UpdateService(router)
		folder := strings.Repeat("/"+strings.Repeat("f", 200), 5) + "/" + strings.Repeat("g", 18)
		m, err := query.AccountByID(clean.IdUint(id))
		require.NoError(t, err)
		prev := m.SyncPath
		require.NoError(t, m.Updates(entity.Values{"sync_path": folder, "acc_sync": true}))
		t.Cleanup(func() { _ = m.Updates(entity.Values{"sync_path": prev, "acc_sync": false}) })

		r := PerformRequestWithBody(app, "PUT", "/api/v1/services/"+id, `{"AccSync": false, "AccPass": "changed"}`)
		assert.Equal(t, http.StatusOK, r.Code, r.Body.String())
		m, err = query.AccountByID(clean.IdUint(id))
		require.NoError(t, err)
		assert.False(t, m.AccSync)
		assert.Equal(t, folder, m.SyncPath)

		r = PerformRequestWithBody(app, "PUT", "/api/v1/services/"+id, fmt.Sprintf(`{"SyncPath": %q}`, folder+"x"))
		assert.Equal(t, http.StatusBadRequest, r.Code)
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
	assert.Equal(t, http.StatusCreated, r.Code)
	id := gjson.Get(r.Body.String(), "ID").String()
	assert.NotEmpty(t, id)
	assert.Equal(t, "/api/v1/services/"+id, r.Header().Get("Location"))

	t.Run("Success", func(t *testing.T) {
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
