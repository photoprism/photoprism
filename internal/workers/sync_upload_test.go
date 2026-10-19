package workers

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	xwebdav "golang.org/x/net/webdav"
	"gorm.io/gorm"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/service/webdav"
	"github.com/photoprism/photoprism/pkg/clean"
)

// failingServer returns a WebDAV server that refuses uploads of names containing "big" with 413, fails those
// containing "fail" with 500, and stalls those containing "stall", unless healed, and counts the uploads per name.
func failingServer(t *testing.T) (*httptest.Server, func(name string) int) {
	t.Helper()

	var mu sync.Mutex
	puts := make(map[string]int)
	healed.Store(false)
	onPut.Store(nil)
	t.Cleanup(func() { healed.Store(false); onPut.Store(nil) })
	t.Cleanup(webdav.SetTransferStallTimeout(100 * time.Millisecond))
	handler := &xwebdav.Handler{FileSystem: xwebdav.Dir(t.TempDir()), LockSystem: xwebdav.NewMemLS()}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			mu.Lock()
			puts[r.URL.Path]++
			mu.Unlock()

			if f := onPut.Load(); f != nil {
				(*f)()
			}

			switch {
			case healed.Load():
			case strings.Contains(r.URL.Path, "stall"):
				_, _ = io.Copy(io.Discard, r.Body)
				<-r.Context().Done()
				return
			case strings.Contains(r.URL.Path, "big"):
				w.WriteHeader(http.StatusRequestEntityTooLarge)
				return
			case strings.Contains(r.URL.Path, "fail"):
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		handler.ServeHTTP(w, r)
	}))
	t.Cleanup(server.Close)

	return server, func(name string) int {
		mu.Lock()
		defer mu.Unlock()
		return puts[name]
	}
}

// healed makes the failing server accept every upload.
var healed atomic.Bool

// onPut is called by the failing server on every upload, if set.
var onPut atomic.Pointer[func()]

// newFailingUploadAccount creates an upload account on a failing server with local files of the specified names.
func newFailingUploadAccount(t *testing.T, retryLimit int, names ...string) (*Sync, entity.Service, func(name string) int) {
	t.Helper()

	conf := config.TestConfig()
	options := *conf.Options()
	t.Cleanup(func() { *conf.Options() = options })
	conf.Options().OriginalsPath = t.TempDir()

	server, puts := failingServer(t)
	a, _ := newStallUploadAccount(t, conf, server, false, names...)
	require.NoError(t, entity.Db().Model(&entity.Service{}).Where("id = ?", a.ID).UpdateColumns(entity.Values{
		"retry_limit": retryLimit, "sync_interval": 0, "sync_status": entity.SyncStatusRefresh}).Error)

	return NewSync(conf), storedSyncAccount(t, a.ID), puts
}

// syncCycle runs the sync worker until the account is synced, starting a new cycle first if it already is.
func syncCycle(t *testing.T, worker *Sync, id uint) entity.Service {
	t.Helper()

	if storedSyncAccount(t, id).SyncStatus == entity.SyncStatusSynced {
		require.NoError(t, entity.Db().Model(&entity.Service{}).Where("id = ?", id).
			UpdateColumn("sync_status", entity.SyncStatusRefresh).Error)
	}

	for range 10 {
		require.NoError(t, worker.Start())

		if a := storedSyncAccount(t, id); !a.AccSync || a.SyncStatus == entity.SyncStatusSynced {
			return a
		}
	}

	t.Fatalf("account %d was not synced", id)

	return entity.Service{}
}

func TestSync_uploadFailures(t *testing.T) {
	t.Run("SomeFail", func(t *testing.T) {
		isolateSyncAccounts(t)
		worker, a, puts := newFailingUploadAccount(t, 3, "a-big.png", "b-fail.png", "c-ok.png")
		hook := captureSyncLog(t)

		// The stage completes with the failed files set aside and the other file uploaded.
		sub := event.Subscribe("sync.synced")
		t.Cleanup(func() { event.Unsubscribe(sub) })
		a = syncCycle(t, worker, a.ID)
		assert.True(t, a.AccSync)
		assert.Equal(t, "2 files could not be uploaded", a.AccError)
		assert.Zero(t, a.AccErrors)
		assert.Equal(t, 1, puts("/a-big.png"))
		assert.Equal(t, 1, puts("/b-fail.png"))
		assert.Equal(t, 1, puts("/c-ok.png"))
		assert.True(t, loggedWarning(hook, "failed to upload 2 files to "+clean.Log(a.AccName)))
		assert.True(t, loggedWarning(hook, "a-big.png"))
		select {
		case msg := <-sub.Receiver:
			t.Fatalf("unexpected event %s", msg.Name)
		case <-time.After(200 * time.Millisecond):
		}

		// Later runs in the same cycle send nothing, and the records are never queued for download.
		require.NoError(t, worker.Start())
		require.NoError(t, worker.Start())
		assert.Equal(t, 1, puts("/a-big.png"))
		queued, err := query.CountQueuedFileSyncs(a.ID, a.RetryLimit)
		require.NoError(t, err)
		assert.Zero(t, queued.Queued)

		// The next cycle tries them again, up to the retry limit, and the account stays enabled.
		for cycle := 2; cycle <= 6; cycle++ {
			a = syncCycle(t, worker, a.ID)
			assert.True(t, a.AccSync, "cycle %d", cycle)
			assert.Equal(t, "2 files could not be uploaded", a.AccError, "cycle %d", cycle)
		}
		assert.Equal(t, 4, puts("/a-big.png"))
		assert.Equal(t, 4, puts("/b-fail.png"))
		assert.Equal(t, 1, puts("/c-ok.png"))

		// Saving the account resets the errors, and once the remote accepts the files, the error is cleared,
		// the placeholders are removed, and the cycle is reported as synced.
		stored := storedSyncAccount(t, a.ID)
		require.NoError(t, stored.ResetErrors(false, true))
		healed.Store(true)
		a = syncCycle(t, worker, a.ID)
		assert.Equal(t, 5, puts("/a-big.png"))
		assert.Empty(t, a.AccError)
		setAside, err := entity.CountFailedFileSyncUploads(a.ID)
		require.NoError(t, err)
		assert.Zero(t, setAside)
		var placeholders int64
		require.NoError(t, entity.Db().Model(&entity.FileSync{}).Where("service_id = ? AND remote_name LIKE ?",
			a.ID, entity.FileSyncUploadPrefix+"%").Count(&placeholders).Error)
		assert.Zero(t, placeholders)
		select {
		case msg := <-sub.Receiver:
			assert.Equal(t, "sync.synced", msg.Name)
		case <-time.After(2 * time.Second):
			t.Fatal("expected one sync.synced event")
		}
	})
	t.Run("AllFail", func(t *testing.T) {
		isolateSyncAccounts(t)
		worker, a, puts := newFailingUploadAccount(t, 2, "a-fail.png", "b-fail.png")

		// Each stage in which every upload failed because of the remote counts once, so the retry limit disables sync.
		for cycle := 1; cycle <= 3; cycle++ {
			a = syncCycle(t, worker, a.ID)
			assert.Equal(t, cycle, a.AccErrors, "cycle %d", cycle)
			assert.Equal(t, "2 files could not be uploaded", a.AccError)
		}

		require.NoError(t, entity.Db().Model(&entity.Service{}).Where("id = ?", a.ID).
			UpdateColumn("sync_status", entity.SyncStatusRefresh).Error)
		require.NoError(t, worker.Start())
		assert.False(t, storedSyncAccount(t, a.ID).AccSync)
		assert.Equal(t, 3, puts("/a-fail.png"))
	})
	t.Run("Restart", func(t *testing.T) {
		isolateSyncAccounts(t)
		worker, a, _ := newFailingUploadAccount(t, 3, "a-fail.png")

		a = syncCycle(t, worker, a.ID)
		assert.Equal(t, 1, a.AccErrors)

		// A stage started before a restart neither counts nor resets the count.
		require.NoError(t, entity.Db().Model(&entity.Service{}).Where("id = ?", a.ID).
			UpdateColumn("sync_status", entity.SyncStatusRefresh).Error)
		require.NoError(t, worker.Start())
		require.Equal(t, entity.SyncStatusUpload, storedSyncAccount(t, a.ID).SyncStatus)
		uploadStages.End(a.ID)
		a = syncCycle(t, worker, a.ID)
		assert.Equal(t, 1, a.AccErrors)
		assert.Equal(t, "1 file could not be uploaded", a.AccError)

		a = syncCycle(t, worker, a.ID)
		assert.Equal(t, 2, a.AccErrors)
	})
	t.Run("UploadResetsCount", func(t *testing.T) {
		isolateSyncAccounts(t)
		worker, a, _ := newFailingUploadAccount(t, 3, "a-fail.png")

		a = syncCycle(t, worker, a.ID)
		assert.Equal(t, 1, a.AccErrors)

		// A stage that uploaded a file resets the count, while the failed file stays set aside.
		require.NoError(t, os.WriteFile(filepath.Join(worker.conf.OriginalsPath(), "b-ok.png"), []byte("b"), 0o600))
		file := &entity.File{PhotoID: 1, PhotoUID: "ps6sg6be2lvl0yh7", FileRoot: entity.RootOriginals, FileName: "b-ok.png",
			FileType: "png", MediaType: entity.MediaImage, FileHash: fmt.Sprintf("%040d", 1299)}
		require.NoError(t, file.Create())
		t.Cleanup(func() { entity.UnscopedDb().Unscoped().Delete(file) })
		a = syncCycle(t, worker, a.ID)
		assert.Zero(t, a.AccErrors)
		assert.Equal(t, "1 file could not be uploaded", a.AccError)
	})
	t.Run("RefusedOnly", func(t *testing.T) {
		isolateSyncAccounts(t)
		worker, a, puts := newFailingUploadAccount(t, 2, "a-big.png")

		// A remote that refuses a file does not count against the account; the file stops at its own limit.
		for cycle := 1; cycle <= 5; cycle++ {
			a = syncCycle(t, worker, a.ID)
			assert.True(t, a.AccSync, "cycle %d", cycle)
			assert.Zero(t, a.AccErrors, "cycle %d", cycle)
		}
		assert.Equal(t, 3, puts("/a-big.png"))
		assert.Equal(t, "1 file could not be uploaded", a.AccError)
	})
	t.Run("MissingFile", func(t *testing.T) {
		isolateSyncAccounts(t)
		worker, a, puts := newFailingUploadAccount(t, 2, "a-gone.png")
		require.NoError(t, os.Remove(filepath.Join(worker.conf.OriginalsPath(), "a-gone.png")))

		// A local file that cannot be read does not count against the account.
		for cycle := 1; cycle <= 5; cycle++ {
			a = syncCycle(t, worker, a.ID)
			assert.True(t, a.AccSync, "cycle %d", cycle)
			assert.Zero(t, a.AccErrors, "cycle %d", cycle)
		}
		assert.Zero(t, puts("/a-gone.png"))
		assert.Equal(t, "1 file could not be uploaded", a.AccError)
	})
	t.Run("Stalls", func(t *testing.T) {
		isolateSyncAccounts(t)
		worker, a, puts := newFailingUploadAccount(t, 3, "a-stall.png")

		// A stalled upload is caused by the remote, so the stage counts.
		a = syncCycle(t, worker, a.ID)
		assert.Equal(t, 1, a.AccErrors)
		assert.Equal(t, 1, puts("/a-stall.png"))
	})
	t.Run("KeepsAdminReset", func(t *testing.T) {
		isolateSyncAccounts(t)
		worker, a, _ := newFailingUploadAccount(t, 3, "a-big.png", "b-big.png")

		// The stage tried both files, which the remote refused, and has nothing left to upload.
		files, err := query.AccountUploads(a, 10)
		require.NoError(t, err)
		require.Len(t, files, 2)
		for _, f := range files {
			require.NoError(t, entity.FailFileSyncUpload(a.ID, f.ID, "refused"))
		}
		require.NoError(t, entity.Db().Model(&entity.Service{}).Where("id = ?", a.ID).UpdateColumns(entity.Values{
			"sync_status": entity.SyncStatusUpload, "acc_errors": 2, "acc_error": uploadSetAsideError(1)}).Error)
		uploadStages.Set(a.ID, uploadStage{tried: true})
		t.Cleanup(func() { uploadStages.End(a.ID) })

		// An admin saves the account while the stage completes, which resets the count.
		require.NoError(t, entity.Db().Callback().Query().Before("gorm:query").Register("test:reset_acc_errors", func(tx *gorm.DB) {
			if tx.Statement.Table == (entity.FileSync{}).TableName() {
				_ = entity.UnscopedDb().Exec("UPDATE services SET acc_errors = 0 WHERE id = ?", a.ID).Error
			}
		}))
		t.Cleanup(func() { _ = entity.Db().Callback().Query().Remove("test:reset_acc_errors") })

		// A stage that keeps the count stores only its message, so the reset is kept.
		require.NoError(t, worker.Start())
		stored := storedSyncAccount(t, a.ID)
		assert.Equal(t, entity.SyncStatusSynced, stored.SyncStatus)
		assert.Zero(t, stored.AccErrors)
		assert.Equal(t, "2 files could not be uploaded", stored.AccError)
	})
	t.Run("AllUploaded", func(t *testing.T) {
		isolateSyncAccounts(t)
		worker, a, puts := newFailingUploadAccount(t, 2, "a-ok.png", "b-ok.png")

		a = syncCycle(t, worker, a.ID)
		assert.Empty(t, a.AccError)
		assert.Zero(t, a.AccErrors)
		assert.Equal(t, 1, puts("/a-ok.png"))
		assert.Equal(t, 1, puts("/b-ok.png"))
	})
}

func TestUploadStages(t *testing.T) {
	t.Run("Started", func(t *testing.T) {
		var m = uploadStageMap{stages: make(map[uint]uploadStage)}
		m.Start(1)
		assert.Equal(t, uploadStage{}, m.Get(1))
		m.Set(1, uploadStage{failed: true})
		stage, known := m.End(1)
		assert.True(t, known)
		assert.True(t, stage.failed)
		_, known = m.End(1)
		assert.False(t, known, "an ended stage is unknown")
	})
	t.Run("Unknown", func(t *testing.T) {
		// A stage started before a restart never turns off YAML sync.
		var m = uploadStageMap{stages: make(map[uint]uploadStage)}
		stage := m.Get(2)
		assert.Equal(t, uploadStage{otherRefused: true, unknown: true}, stage)
		stage.failed = true
		m.Set(2, stage)
		_, known := m.End(2)
		assert.False(t, known, "a stage started before a restart stays unknown")
	})
}

func TestUploadSetAsideError(t *testing.T) {
	assert.Equal(t, "1 file could not be uploaded", uploadSetAsideError(1))
	assert.Equal(t, "3 files could not be uploaded", uploadSetAsideError(3))
	assert.True(t, isUploadSetAsideError(uploadSetAsideError(2)))
	assert.False(t, isUploadSetAsideError("failed to list remote folders"))
	assert.False(t, isUploadSetAsideError(""))
}
