package workers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/webdav"
	"gorm.io/gorm"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/clean"
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
	t.Run("ControlCharactersSkipped", func(t *testing.T) {
		for _, name := range []string{"ctl\x01name.jpg", "line\nbreak.jpg", "del\x7f.jpg"} {
			write(t, name, 100, listed)
		}
		write(t, "Älter 日本 صور.jpg", 100, listed)
		refresh(t)

		var count int64
		require.NoError(t, entity.Db().Model(&entity.FileSync{}).
			Where("service_id = ? AND remote_name IN (?)", a.ID, []string{"/ctl\x01name.jpg", "/line\nbreak.jpg", "/del\x7f.jpg"}).
			Count(&count).Error)
		assert.Zero(t, count)
		assert.Equal(t, entity.FileSyncNew, stored(t, "/Älter 日本 صور.jpg").Status)
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
		entity.Db().Callback().Update().After("gorm:update").Register("test:count_file_sync_updates", func(tx *gorm.DB) {
			if tx.Statement.Table == (entity.FileSync{}).TableName() {
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

func TestSync_refreshLogs(t *testing.T) {
	worker := NewSync(config.TestConfig())
	remote := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(remote, "sub"), fs.ModeDir))
	require.NoError(t, os.WriteFile(filepath.Join(remote, "sub", "photo.jpg"), []byte("jpg"), fs.ModeFile))
	require.NoError(t, os.WriteFile(filepath.Join(remote, "root.jpg"), []byte("jpg"), fs.ModeFile))

	// The remote lists its folders, but fails to list the files of the subfolder when failSub is set.
	var failSub atomic.Bool
	handler := &webdav.Handler{FileSystem: webdav.Dir(remote), LockSystem: webdav.NewMemLS()}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if failSub.Load() && r.Method == "PROPFIND" && r.Header.Get("Depth") == "1" && strings.HasPrefix(r.URL.Path, "/sub") {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		handler.ServeHTTP(w, r)
	}))
	t.Cleanup(server.Close)

	a := entity.Service{AccName: "Sync Logs " + rnd.Base36(8), AccURL: server.URL + "/", AccType: "webdav", AccSync: true,
		SyncDownload: true, SyncPath: "/", AccTimeout: "low", RetryLimit: -1}
	require.NoError(t, entity.Db().Create(&a).Error)
	t.Cleanup(func() {
		assert.NoError(t, entity.UnscopedDb().Unscoped().Delete(&entity.FileSync{}, "service_id = ?", a.ID).Error)
		assert.NoError(t, entity.UnscopedDb().Unscoped().Delete(&entity.Service{}, a.ID).Error)
	})

	// errorsContaining returns the error messages logged that contain text.
	errorsContaining := func(hook *test.Hook, text string) (result []string) {
		for _, e := range hook.AllEntries() {
			if e.Level == logrus.ErrorLevel && strings.Contains(e.Message, text) {
				result = append(result, e.Message)
			}
		}
		return result
	}

	// The sync folder is listed once, although the folder listing includes it.
	t.Run("RecordFailed", func(t *testing.T) {
		require.NoError(t, entity.Db().Callback().Create().Before("gorm:create").Register("test:fail_file_sync_create", func(tx *gorm.DB) {
			if f, ok := tx.Statement.Dest.(*entity.FileSync); ok && f.ServiceID == a.ID && f.RemoteName == "/root.jpg" {
				_ = tx.AddError(errors.New("create failed"))
			}
		}))
		t.Cleanup(func() { _ = entity.Db().Callback().Create().Remove("test:fail_file_sync_create") })
		hook := captureSyncLog(t)
		complete, err := worker.refresh(a)
		require.NoError(t, err)
		assert.True(t, complete)
		assert.Len(t, errorsContaining(hook, "failed to record "+clean.Log("/root.jpg")+" from "+clean.Log(a.AccName)), 1)
	})
	t.Run("SyncPathSpelling", func(t *testing.T) {
		// A sync path spelled differently from the folder listing is still listed once.
		b := a
		b.SyncPath = "sub/"
		require.NoError(t, os.WriteFile(filepath.Join(remote, "sub", "spelling.jpg"), []byte("jpg"), fs.ModeFile))
		require.NoError(t, entity.Db().Callback().Create().Before("gorm:create").Register("test:fail_file_sync_create", func(tx *gorm.DB) {
			if f, ok := tx.Statement.Dest.(*entity.FileSync); ok && f.ServiceID == b.ID && f.RemoteName == "/sub/spelling.jpg" {
				_ = tx.AddError(errors.New("create failed"))
			}
		}))
		t.Cleanup(func() { _ = entity.Db().Callback().Create().Remove("test:fail_file_sync_create") })
		hook := captureSyncLog(t)
		complete, err := worker.refresh(b)
		require.NoError(t, err)
		assert.True(t, complete)
		assert.Len(t, errorsContaining(hook, "failed to record "+clean.Log("/sub/spelling.jpg")+" from "+clean.Log(b.AccName)), 1)
	})
	t.Run("ListFilesFailed", func(t *testing.T) {
		failSub.Store(true)
		t.Cleanup(func() { failSub.Store(false) })
		hook := captureSyncLog(t)
		_, err := worker.refresh(a)
		require.Error(t, err)
		assert.Len(t, errorsContaining(hook, "(list remote files of "+clean.Log(a.AccName)+")"), 1)
	})
}
