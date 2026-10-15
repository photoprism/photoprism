package workers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	xwebdav "golang.org/x/net/webdav"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
)

// TestSync_StartSettingsChanged checks that a run never advances a stage with a folder or direction that an admin
// saved during the run, and that a stage whose direction is off ends without transfers.
func TestSync_StartSettingsChanged(t *testing.T) {
	isolateSyncAccounts(t)

	worker := NewSync(config.TestConfig())

	// newSavingAccount creates an account on a WebDAV server that, on the first request, saves the account with
	// the changes made by update, as an admin saving it during the run does.
	newSavingAccount := func(t *testing.T, status string, setup func(m *entity.Service), update func(f *form.Service)) (*entity.Service, *atomic.Int32) {
		var a *entity.Service
		var once sync.Once
		var requests atomic.Int32

		handler := &xwebdav.Handler{FileSystem: xwebdav.Dir(t.TempDir()), LockSystem: xwebdav.NewMemLS()}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requests.Add(1)

			if update != nil {
				once.Do(func() {
					stored := storedSyncAccount(t, a.ID)
					frm, err := form.NewService(stored)
					assert.NoError(t, err)
					update(&frm)
					assert.NoError(t, stored.SaveForm(frm))
				})
			}

			handler.ServeHTTP(w, r)
		}))
		t.Cleanup(server.Close)

		a = newSyncAccount(t, func(m *entity.Service) {
			m.AccURL, m.AccTimeout, m.SyncStatus, m.SyncPath = server.URL+"/", "low", status, "/old"
			m.RetryLimit = -1

			if setup != nil {
				setup(m)
			}
		})
		t.Cleanup(func() { _ = entity.UnscopedDb().Delete(&entity.FileSync{}, "service_id = ?", a.ID).Error })

		if status == entity.SyncStatusDownload {
			require.NoError(t, entity.NewFileSync(a.ID, "/old/changing.jpg").Create())
		}

		return a, &requests
	}

	download := func(m *entity.Service) { m.SyncDownload = true }
	upload := func(m *entity.Service) { m.SyncUpload = true }

	t.Run("PathChangedDuringRefresh", func(t *testing.T) {
		a, _ := newSavingAccount(t, entity.SyncStatusRefresh, download, func(f *form.Service) { f.SyncPath = "/new" })
		require.NoError(t, worker.Start())
		stored := storedSyncAccount(t, a.ID)
		assert.Equal(t, "/new", stored.SyncPath)
		assert.Equal(t, entity.SyncStatusRefresh, stored.SyncStatus)
	})
	t.Run("PathChangedDuringDownload", func(t *testing.T) {
		a, _ := newSavingAccount(t, entity.SyncStatusDownload, download, func(f *form.Service) { f.SyncPath = "/new" })
		require.NoError(t, worker.Start())
		stored := storedSyncAccount(t, a.ID)
		assert.Equal(t, "/new", stored.SyncPath)
		assert.Equal(t, entity.SyncStatusRefresh, stored.SyncStatus)
	})
	t.Run("PathChangedDuringUpload", func(t *testing.T) {
		a, _ := newSavingAccount(t, entity.SyncStatusUpload, upload, func(f *form.Service) { f.SyncPath = "/new" })
		require.NoError(t, worker.Start())
		stored := storedSyncAccount(t, a.ID)
		assert.Equal(t, "/new", stored.SyncPath)
		assert.Equal(t, entity.SyncStatusRefresh, stored.SyncStatus)
	})
	t.Run("DownloadOffDuringDownload", func(t *testing.T) {
		a, _ := newSavingAccount(t, entity.SyncStatusDownload, download, func(f *form.Service) { f.SyncDownload = false })
		require.NoError(t, worker.Start())
		assert.Equal(t, entity.SyncStatusRefresh, storedSyncAccount(t, a.ID).SyncStatus)
	})
	t.Run("DownloadOff", func(t *testing.T) {
		// A download stage whose direction is off ends without a request and moves on.
		a, requests := newSavingAccount(t, entity.SyncStatusDownload, upload, nil)
		require.NoError(t, worker.Start())
		assert.Zero(t, requests.Load())
		assert.Equal(t, entity.SyncStatusUpload, storedSyncAccount(t, a.ID).SyncStatus)
	})
	t.Run("UploadOff", func(t *testing.T) {
		a, requests := newSavingAccount(t, entity.SyncStatusUpload, download, nil)
		require.NoError(t, worker.Start())
		assert.Zero(t, requests.Load())
		assert.Equal(t, entity.SyncStatusSynced, storedSyncAccount(t, a.ID).SyncStatus)
	})
	t.Run("UploadOnDuringRefresh", func(t *testing.T) {
		// Without downloads, the refresh would complete the cycle, but the direction saved meanwhile keeps it in refresh.
		a, _ := newSavingAccount(t, entity.SyncStatusRefresh, nil, func(f *form.Service) { f.SyncUpload = true })
		require.NoError(t, worker.Start())
		stored := storedSyncAccount(t, a.ID)
		assert.True(t, stored.SyncUpload)
		assert.Equal(t, entity.SyncStatusRefresh, stored.SyncStatus)
	})
	t.Run("DownloadOnDuringRefresh", func(t *testing.T) {
		// Without a direction, the refresh would complete the cycle, but downloads turned on meanwhile keep it in refresh.
		a, _ := newSavingAccount(t, entity.SyncStatusRefresh, nil, func(f *form.Service) { f.SyncDownload = true })
		require.NoError(t, worker.Start())
		stored := storedSyncAccount(t, a.ID)
		assert.True(t, stored.SyncDownload)
		assert.Equal(t, entity.SyncStatusRefresh, stored.SyncStatus)
	})
	t.Run("RefreshUploadOffResetsSetAside", func(t *testing.T) {
		// A completed refresh resets the count of an account error about set-aside uploads once uploads are off.
		a, _ := newSavingAccount(t, entity.SyncStatusRefresh, func(m *entity.Service) {
			m.SyncDownload, m.AccErrors, m.AccError = true, 2, uploadSetAsideError(1)
		}, nil)
		require.NoError(t, worker.Start())
		stored := storedSyncAccount(t, a.ID)
		assert.Equal(t, entity.SyncStatusDownload, stored.SyncStatus)
		assert.Zero(t, stored.AccErrors)
		assert.Empty(t, stored.AccError)
	})
	t.Run("NullSyncPath", func(t *testing.T) {
		// A row of an earlier release may hold NULL as its sync folder, which is loaded as empty.
		a, _ := newSavingAccount(t, entity.SyncStatusRefresh, download, nil)
		require.NoError(t, entity.Db().Exec("UPDATE services SET sync_path = NULL WHERE id = ?", a.ID).Error)
		require.NoError(t, worker.Start())
		assert.Equal(t, entity.SyncStatusDownload, storedSyncAccount(t, a.ID).SyncStatus)
	})
	t.Run("UploadOffClearsSetAside", func(t *testing.T) {
		// An upload stage whose direction is off does not report uploads set aside earlier.
		a, _ := newSavingAccount(t, entity.SyncStatusUpload, func(m *entity.Service) {
			m.SyncDownload, m.AccErrors, m.AccError = true, 1, uploadSetAsideError(1)
		}, nil)
		require.NoError(t, entity.FailFileSyncUpload(a.ID, 1, "failed"))
		require.NoError(t, worker.Start())
		stored := storedSyncAccount(t, a.ID)
		assert.Equal(t, entity.SyncStatusSynced, stored.SyncStatus)
		assert.Zero(t, stored.AccErrors)
		assert.Empty(t, stored.AccError)
	})
	t.Run("UnlistedQueueDropped", func(t *testing.T) {
		// A completed refresh removes queued files it did not list, such as those of a previous sync folder.
		a, _ := newSavingAccount(t, entity.SyncStatusRefresh, download, nil)
		require.NoError(t, entity.NewFileSync(a.ID, "/previous/queued.jpg").Create())
		failed := entity.NewFileSync(a.ID, "/previous/failed.jpg")
		failed.Status = entity.FileSyncFailed
		require.NoError(t, failed.Create())
		require.NoError(t, worker.Start())
		assert.Equal(t, entity.SyncStatusDownload, storedSyncAccount(t, a.ID).SyncStatus)
		var names []string
		require.NoError(t, entity.Db().Model(&entity.FileSync{}).Where("service_id = ?", a.ID).Pluck("remote_name", &names).Error)
		assert.Equal(t, []string{"/previous/failed.jpg"}, names)
	})
	t.Run("OtherSettingsKeepStage", func(t *testing.T) {
		// A save that changes neither the folder nor a direction lets the run advance.
		a, _ := newSavingAccount(t, entity.SyncStatusRefresh, download, func(f *form.Service) { f.SyncInterval = 60 })
		require.NoError(t, worker.Start())
		stored := storedSyncAccount(t, a.ID)
		assert.Equal(t, 60, stored.SyncInterval)
		assert.Equal(t, entity.SyncStatusDownload, stored.SyncStatus)
	})
}

func TestSync_downloadIfEnabled(t *testing.T) {
	worker := NewSync(config.TestConfig())
	t.Run("Off", func(t *testing.T) {
		// The remote is never contacted.
		complete, err := worker.downloadIfEnabled(entity.Service{AccURL: "http://127.0.0.1:1/", SyncDownload: false})
		require.NoError(t, err)
		assert.True(t, complete)
	})
	t.Run("NothingQueued", func(t *testing.T) {
		complete, err := worker.downloadIfEnabled(entity.Service{ID: 4741, AccURL: "http://127.0.0.1:1/", SyncDownload: true})
		require.NoError(t, err)
		assert.True(t, complete)
	})
}

func TestQueuedFileSyncNames(t *testing.T) {
	const serviceID = 4743
	t.Cleanup(func() { _ = entity.UnscopedDb().Delete(&entity.FileSync{}, "service_id = ?", serviceID).Error })

	names, err := queuedFileSyncNames(serviceID)
	require.NoError(t, err)
	assert.Empty(t, names)

	require.NoError(t, entity.NewFileSync(serviceID, "/a.jpg").Create())
	downloaded := entity.NewFileSync(serviceID, "/b.jpg")
	downloaded.Status = entity.FileSyncDownloaded
	require.NoError(t, downloaded.Create())

	names, err = queuedFileSyncNames(serviceID)
	require.NoError(t, err)
	assert.Equal(t, map[string]struct{}{"/a.jpg": {}}, names)
}

func TestDropQueuedFileSyncs(t *testing.T) {
	const serviceID = 4742
	t.Cleanup(func() { _ = entity.UnscopedDb().Delete(&entity.FileSync{}, "service_id = ?", serviceID).Error })

	for i := range 1001 {
		require.NoError(t, entity.NewFileSync(serviceID, fmt.Sprintf("/%04d.jpg", i)).Create())
	}
	downloaded := entity.NewFileSync(serviceID, "/d.jpg")
	downloaded.Status = entity.FileSyncDownloaded
	require.NoError(t, downloaded.Create())

	// Every queued name except one is dropped, in batches, and other records are kept.
	drop, err := queuedFileSyncNames(serviceID)
	require.NoError(t, err)
	delete(drop, "/0500.jpg")
	drop["/d.jpg"] = struct{}{}
	require.NoError(t, dropQueuedFileSyncs(serviceID, drop))
	require.NoError(t, dropQueuedFileSyncs(serviceID, nil))

	var names []string
	require.NoError(t, entity.Db().Model(&entity.FileSync{}).Where("service_id = ?", serviceID).Pluck("remote_name", &names).Error)
	assert.ElementsMatch(t, []string{"/0500.jpg", "/d.jpg"}, names)
}
