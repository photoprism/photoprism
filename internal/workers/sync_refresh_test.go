package workers

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/webdav"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestSync_refreshRemoteChanges(t *testing.T) {
	worker := NewSync(config.TestConfig())
	remote := t.TempDir()

	server := httptest.NewServer(&webdav.Handler{FileSystem: webdav.Dir(remote), LockSystem: webdav.NewMemLS()})
	t.Cleanup(server.Close)

	a := entity.Service{AccName: "Sync Refresh " + rnd.Base36(8), AccURL: server.URL + "/", AccType: "webdav", AccSync: true,
		SyncDownload: true, SyncPath: "/", AccTimeout: "low", RetryLimit: -1}
	require.NoError(t, entity.Db().Create(&a).Error)
	t.Cleanup(func() {
		assert.NoError(t, entity.UnscopedDb().Unscoped().Delete(&entity.FileSync{}, "service_id = ?", a.ID).Error)
		assert.NoError(t, entity.UnscopedDb().Unscoped().Delete(&entity.Service{}, a.ID).Error)
	})

	// write stores a remote file with the specified size and modification time.
	write := func(t *testing.T, name string, size int, modTime time.Time) {
		fileName := filepath.Join(remote, name)
		require.NoError(t, os.WriteFile(fileName, make([]byte, size), fs.ModeFile))
		require.NoError(t, os.Chtimes(fileName, modTime, modTime))
	}

	// stored returns the stored record of a remote file.
	stored := func(t *testing.T, name string) entity.FileSync {
		var f entity.FileSync
		require.NoError(t, entity.Db().Where("service_id = ? AND remote_name = ?", a.ID, name).First(&f).Error)
		return f
	}

	// refresh lists the remote files.
	refresh := func(t *testing.T) {
		complete, err := worker.refresh(a)
		require.NoError(t, err)
		require.True(t, complete)
	}

	listed := time.Now().Add(-time.Hour).Truncate(time.Second)
	changed := listed.Add(time.Minute)

	t.Run("NewSizeChanged", func(t *testing.T) {
		// A replaced file may keep its date, so a size change alone is stored too.
		write(t, "size.jpg", 100, listed)
		refresh(t)
		assert.Equal(t, int64(100), stored(t, "/size.jpg").RemoteSize)

		write(t, "size.jpg", 50, listed)
		refresh(t)
		f := stored(t, "/size.jpg")
		assert.Equal(t, entity.FileSyncNew, f.Status)
		assert.Equal(t, int64(50), f.RemoteSize)
		assert.Equal(t, listed.Unix(), f.RemoteDate.Unix())
	})
	t.Run("NewDateChanged", func(t *testing.T) {
		// The size stays the same, so only the date change can trigger the update.
		write(t, "date.jpg", 100, listed)
		refresh(t)

		write(t, "date.jpg", 100, changed)
		refresh(t)
		f := stored(t, "/date.jpg")
		assert.Equal(t, entity.FileSyncNew, f.Status)
		assert.Equal(t, int64(100), f.RemoteSize)
		assert.Equal(t, changed.Unix(), f.RemoteDate.Unix())
	})
	t.Run("NewErrorsKept", func(t *testing.T) {
		write(t, "errors.jpg", 100, listed)
		refresh(t)
		f := stored(t, "/errors.jpg")
		require.NoError(t, f.Updates(entity.Values{"Errors": 2, "Error": "failed"}))

		write(t, "errors.jpg", 50, changed)
		refresh(t)
		f = stored(t, "/errors.jpg")
		assert.Equal(t, 2, f.Errors)
		assert.Equal(t, "failed", f.Error)
	})
	t.Run("NewUnchanged", func(t *testing.T) {
		// An unchanged listing writes nothing, since refresh runs on every sync.
		write(t, "same.jpg", 100, listed)
		refresh(t)

		var mu sync.Mutex
		updates := 0
		entity.Db().Callback().Update().After("gorm:update").Register("test:count_file_sync_updates", func(scope *gorm.Scope) {
			if scope.TableName() == (entity.FileSync{}).TableName() {
				mu.Lock()
				updates++
				mu.Unlock()
			}
		})
		t.Cleanup(func() { entity.Db().Callback().Update().Remove("test:count_file_sync_updates") })

		refresh(t)
		mu.Lock()
		defer mu.Unlock()
		assert.Equal(t, 0, updates)
	})
	t.Run("DownloadedDateChanged", func(t *testing.T) {
		write(t, "done.jpg", 100, listed)
		refresh(t)
		f := stored(t, "/done.jpg")
		require.NoError(t, f.Updates(entity.Values{"Status": entity.FileSyncDownloaded}))

		write(t, "done.jpg", 60, changed)
		refresh(t)
		f = stored(t, "/done.jpg")
		assert.Equal(t, entity.FileSyncNew, f.Status)
		assert.Equal(t, int64(60), f.RemoteSize)
		assert.Equal(t, changed.Unix(), f.RemoteDate.Unix())
	})
	t.Run("FailedChanged", func(t *testing.T) {
		// A failed file stays failed, since refresh only updates new and downloaded files.
		write(t, "failed.jpg", 100, listed)
		refresh(t)
		f := stored(t, "/failed.jpg")
		require.NoError(t, f.Updates(entity.Values{"Status": entity.FileSyncFailed}))

		write(t, "failed.jpg", 60, changed)
		refresh(t)
		f = stored(t, "/failed.jpg")
		assert.Equal(t, entity.FileSyncFailed, f.Status)
		assert.Equal(t, int64(100), f.RemoteSize)
	})
	t.Run("DownloadedSizeChanged", func(t *testing.T) {
		// A downloaded file is queued again only when its date changes.
		write(t, "kept.jpg", 100, listed)
		refresh(t)
		f := stored(t, "/kept.jpg")
		require.NoError(t, f.Updates(entity.Values{"Status": entity.FileSyncDownloaded}))

		write(t, "kept.jpg", 60, listed)
		refresh(t)
		f = stored(t, "/kept.jpg")
		assert.Equal(t, entity.FileSyncDownloaded, f.Status)
		assert.Equal(t, int64(100), f.RemoteSize)
	})
}
