package api

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/fs/disk"
	"github.com/photoprism/photoprism/pkg/i18n"
)

func TestCancelIndex(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, router, _ := NewApiTest()
		CancelIndexing(router)
		r := PerformRequest(app, "DELETE", "/api/v1/index")

		var resp i18n.Response

		if err := json.Unmarshal(r.Body.Bytes(), &resp); err != nil {
			t.Fatal(err)
		}

		assert.True(t, resp.Success())
		assert.Equal(t, i18n.Msg(i18n.MsgIndexingCanceled), resp.Message)
		assert.Equal(t, i18n.Msg(i18n.MsgIndexingCanceled), resp.String())
		assert.Equal(t, http.StatusOK, r.Code)
		assert.Equal(t, http.StatusOK, resp.Code)
	})
}

func TestStartIndexing(t *testing.T) {
	t.Run("InsufficientStorage", func(t *testing.T) {
		app, router, conf := NewApiTest()

		disk.FlushFree()
		t.Cleanup(disk.FlushFree)
		disk.SetFree(conf.StoragePath(), 1*disk.MB, 1000*disk.MB)
		// Reset the storage-check latch, which an earlier test may have tripped by probing a
		// temp path that duf cannot resolve to a mount point.
		config.DisableStorageCheck.Store(false)

		StartIndexing(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/index", "{}")

		assert.Equal(t, http.StatusInsufficientStorage, r.Code)
	})
	t.Run("FaceMigrationRunning", func(t *testing.T) {
		app, router, conf := NewApiTest()

		disk.FlushFree()
		t.Cleanup(disk.FlushFree)
		config.DisableStorageCheck.Store(false)

		lock, err := mutex.AcquireFileLock(conf.FacesLockFile(), "faces migration")
		require.NoError(t, err)
		t.Cleanup(lock.Release)

		StartIndexing(router)
		sub := event.Subscribe("notify.info")
		defer event.Unsubscribe(sub)

		r := PerformRequestWithBody(app, "POST", "/api/v1/index", "{}")
		assert.Equal(t, http.StatusConflict, r.Code, r.Body.String())

		var resp i18n.Response
		require.NoError(t, json.Unmarshal(r.Body.Bytes(), &resp))
		assert.Equal(t, i18n.Msg(i18n.ErrMigrationInProgress), resp.Error)

		// The request is refused before indexing is announced.
		select {
		case msg := <-sub.Receiver:
			t.Fatalf("unexpected notification %v", msg.Fields)
		case <-time.After(100 * time.Millisecond):
		}
	})
	t.Run("FaceMigrationAfterAuthorization", func(t *testing.T) {
		// A caller that may not index is refused as such, whether or not a migration is running.
		app, router, conf := NewApiTest()

		prevAuthMode := conf.AuthMode()
		conf.SetAuthMode(config.AuthModePasswd)
		t.Cleanup(func() { conf.SetAuthMode(prevAuthMode) })

		denied, err := entity.AddClientSession("index-lock-test", conf.SessionMaxAge(), "people", authn.GrantClientCredentials, nil)
		require.NoError(t, err)
		t.Cleanup(func() { entity.UnscopedDb().Unscoped().Delete(denied) })

		lock, err := mutex.AcquireFileLock(conf.FacesLockFile(), "faces migration")
		require.NoError(t, err)
		t.Cleanup(lock.Release)

		StartIndexing(router)

		r := PerformRequestWithBody(app, "POST", "/api/v1/index", "{}")
		assert.Equal(t, http.StatusUnauthorized, r.Code, r.Body.String())

		r = AuthenticatedRequestWithBody(app, "POST", "/api/v1/index", "{}", denied.AuthToken())
		assert.Equal(t, http.StatusForbidden, r.Code, r.Body.String())
	})
	t.Run("RejectsPathTraversal", func(t *testing.T) {
		app, router, _ := NewApiTest()

		disk.FlushFree()
		t.Cleanup(disk.FlushFree)
		config.DisableStorageCheck.Store(false)

		StartIndexing(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/index", `{"path":"../../../outside"}`)

		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
}
