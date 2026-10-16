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

// TestSync_refreshInvalidNames checks that remote files and folders whose names are not portable are not queued.
func TestSync_refreshInvalidNames(t *testing.T) {
	worker := NewSync(config.TestConfig())
	remote := t.TempDir()
	for _, name := range []string{"valid.jpg", "03 - Março.jpg", "sub/ok.jpg", "invalid-\xff.jpg", "a|b.jpg", "x:y.jpg",
		"a~b.jpg", "dir*/inner.jpg", "$RECYCLE.BIN/deleted.jpg", "~$draft.jpg", "@Family/x.jpg", ".locks/x.jpg"} {
		require.NoError(t, os.MkdirAll(filepath.Dir(filepath.Join(remote, name)), fs.ModeDir))
		require.NoError(t, os.WriteFile(filepath.Join(remote, name), []byte("jpg"), fs.ModeFile))
	}

	server := httptest.NewServer(&webdav.Handler{FileSystem: webdav.Dir(remote), LockSystem: webdav.NewMemLS()})
	t.Cleanup(server.Close)

	a := entity.Service{AccName: "Sync Invalid " + rnd.Base36(8), AccURL: server.URL + "/", AccType: "webdav", AccSync: true,
		SyncDownload: true, SyncPath: "/", AccTimeout: "low", RetryLimit: -1}
	require.NoError(t, entity.Db().Create(&a).Error)
	t.Cleanup(func() {
		assert.NoError(t, entity.UnscopedDb().Unscoped().Delete(&entity.FileSync{}, "service_id = ?", a.ID).Error)
		assert.NoError(t, entity.UnscopedDb().Unscoped().Delete(&entity.Service{}, a.ID).Error)
	})

	// Files queued before their names were checked are removed from the queue.
	for _, name := range []string{"/a|b.jpg", "/dir*/inner.jpg"} {
		queued := entity.NewFileSync(a.ID, name)
		queued.Status = entity.FileSyncNew
		require.NoError(t, queued.Create())
	}

	hook := captureSyncLog(t)
	complete, err := worker.refresh(a)
	require.NoError(t, err)
	assert.True(t, complete)

	var names []string
	require.NoError(t, entity.Db().Model(&entity.FileSync{}).Where("service_id = ?", a.ID).Pluck("remote_name", &names).Error)
	assert.ElementsMatch(t, []string{"/valid.jpg", "/03 - Março.jpg", "/sub/ok.jpg", "/@Family/x.jpg"}, names)
	assert.Zero(t, countLogged(hook, logrus.ErrorLevel, ""))
	assert.Equal(t, 1, countLogged(hook, logrus.WarnLevel, "skipped 7 remote names in"))
	assert.Equal(t, 1, countLogged(hook, logrus.WarnLevel, "with unsupported characters"))
	// A folder whose name is not portable is skipped without listing its files.
	assert.Equal(t, 1, countLogged(hook, logrus.WarnLevel, "/dir*"))
	assert.Zero(t, countLogged(hook, logrus.WarnLevel, "inner.jpg"))

	// Names starting with "$" or "~" are counted as not portable, while dot-directories are excluded without a warning.
	assert.Zero(t, countLogged(hook, logrus.WarnLevel, "locks"))

	// The name of the sync folder itself is checked only when files keep their remote paths in originals.
	for _, filenames := range []bool{false, true} {
		b := entity.Service{AccName: "Sync Invalid " + rnd.Base36(8), AccURL: server.URL + "/", AccType: "webdav", AccSync: true,
			SyncDownload: true, SyncFilenames: filenames, SyncPath: "/dir*", AccTimeout: "low", RetryLimit: -1}
		require.NoError(t, entity.Db().Create(&b).Error)
		t.Cleanup(func() {
			assert.NoError(t, entity.UnscopedDb().Unscoped().Delete(&entity.FileSync{}, "service_id = ?", b.ID).Error)
			assert.NoError(t, entity.UnscopedDb().Unscoped().Delete(&entity.Service{}, b.ID).Error)
		})

		// A file queued below another sync folder is removed from the queue in both modes.
		queued := entity.NewFileSync(b.ID, "/sub/ok.jpg")
		queued.Status = entity.FileSyncNew
		require.NoError(t, queued.Create())

		hook.Reset()
		complete, err = worker.refresh(b)
		assert.True(t, complete)

		if filenames {
			assert.Equal(t, syncNotice("sync folder is not supported for downloads"), err)
		} else {
			require.NoError(t, err)
		}

		names = nil
		require.NoError(t, entity.Db().Model(&entity.FileSync{}).Where("service_id = ?", b.ID).Pluck("remote_name", &names).Error)

		if filenames {
			assert.Empty(t, names)
			assert.Equal(t, 1, countLogged(hook, logrus.WarnLevel, "skipped downloads from"))
			assert.Equal(t, 1, countLogged(hook, logrus.WarnLevel, ""))
		} else {
			assert.Equal(t, []string{"/dir*/inner.jpg"}, names)
			assert.Zero(t, countLogged(hook, logrus.WarnLevel, ""))
		}
	}
}

func TestPortableSyncName(t *testing.T) {
	t.Run("Portable", func(t *testing.T) {
		assert.True(t, portableSyncName("/", "/a/b.jpg"))
		assert.True(t, portableSyncName("/", "/"))
		assert.True(t, portableSyncName("/Photos", "/Photos"))
		assert.True(t, portableSyncName("/A|B", "/A|B/ok.jpg"))
		assert.True(t, portableSyncName("/A|B/", "A|B/sub/ok.jpg"))
		assert.True(t, portableSyncName("", "/03 - Março/#1 & Friends.jpg"))
	})
	t.Run("NotPortable", func(t *testing.T) {
		assert.False(t, portableSyncName("/", "/a|b.jpg"))
		assert.False(t, portableSyncName("/Photos", "/Photos/dir*/x.jpg"))
		assert.False(t, portableSyncName("/A|B", "/A|B/x:y.jpg"))
		assert.False(t, portableSyncName("/", "/caf\xe9.jpg"))
		assert.False(t, portableSyncName("/", "/A|B/ok.jpg"))
		// A path equal to the root, and a sibling that starts with its name, are checked in full.
		assert.False(t, portableSyncName("/A|B", "/A|B"))
		assert.False(t, portableSyncName("/A|B", "/A|Bx/ok.jpg"))
	})
}

func TestSyncNameRoot(t *testing.T) {
	t.Run("SyncFolder", func(t *testing.T) {
		assert.Equal(t, "/A|B", syncNameRoot(entity.Service{SyncPath: "/A|B"}))
	})
	t.Run("Filenames", func(t *testing.T) {
		assert.Equal(t, "/", syncNameRoot(entity.Service{SyncPath: "/A|B", SyncFilenames: true}))
	})
}

// TestSync_refreshExcludedSyncFolder checks that an account whose sync folder is excluded is reported once per refresh.
func TestSync_refreshExcludedSyncFolder(t *testing.T) {
	worker := NewSync(config.TestConfig())
	a := entity.Service{AccName: "Sync Excluded " + rnd.Base36(8), AccURL: "http://127.0.0.1:1/", AccType: "webdav", AccSync: true,
		SyncDownload: true, SyncPath: "/.photos", AccTimeout: "low", RetryLimit: -1}
	require.NoError(t, entity.Db().Create(&a).Error)
	t.Cleanup(func() {
		assert.NoError(t, entity.UnscopedDb().Unscoped().Delete(&entity.Service{}, a.ID).Error)
	})

	hook := captureSyncLog(t)
	complete, err := worker.refresh(a)
	assert.Equal(t, syncNotice("sync folder is not supported"), err)
	assert.True(t, complete)
	assert.Equal(t, 1, countLogged(hook, logrus.WarnLevel, "is excluded"))
}

// TestSync_refreshUnsupportedDest checks that refresh reports an import destination that gives unsupported names.
func TestSync_refreshUnsupportedDest(t *testing.T) {
	conf := config.TestConfig()
	worker := NewSync(conf)
	dest := conf.Options().ImportDest
	t.Cleanup(func() { conf.Options().ImportDest = dest })
	conf.Options().ImportDest = "a:b"

	remote := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(remote, "x.jpg"), []byte("jpg"), fs.ModeFile))
	server := httptest.NewServer(&webdav.Handler{FileSystem: webdav.Dir(remote), LockSystem: webdav.NewMemLS()})
	t.Cleanup(server.Close)

	for _, filenames := range []bool{false, true} {
		a := entity.Service{AccName: "Sync Dest " + rnd.Base36(8), AccURL: server.URL + "/", AccType: "webdav", AccSync: true,
			SyncDownload: true, SyncFilenames: filenames, SyncPath: "/", AccTimeout: "low", RetryLimit: -1}
		require.NoError(t, entity.Db().Create(&a).Error)
		t.Cleanup(func() {
			assert.NoError(t, entity.UnscopedDb().Unscoped().Delete(&entity.FileSync{}, "service_id = ?", a.ID).Error)
			assert.NoError(t, entity.UnscopedDb().Unscoped().Delete(&entity.Service{}, a.ID).Error)
		})

		complete, err := worker.refresh(a)
		assert.True(t, complete)

		var names []string
		require.NoError(t, entity.Db().Model(&entity.FileSync{}).Where("service_id = ?", a.ID).Pluck("remote_name", &names).Error)

		if filenames {
			// Files that keep their remote paths are not imported, so the account is listed as usual.
			assert.NoError(t, err)
			assert.Equal(t, []string{"/x.jpg"}, names)
		} else {
			assert.Equal(t, syncNotice("import destination is not supported"), err)
			assert.Empty(t, names)
		}
	}
}

// TestSync_StartNotice checks that a notice is stored as account error without counting toward the retry limit,
// and cleared once the sync folder is supported.
func TestSync_StartNotice(t *testing.T) {
	worker := NewSync(config.TestConfig())
	remote := t.TempDir()
	server := httptest.NewServer(&webdav.Handler{FileSystem: webdav.Dir(remote), LockSystem: webdav.NewMemLS()})
	t.Cleanup(server.Close)

	a := entity.Service{AccName: "Sync Notice " + rnd.Base36(8), AccURL: server.URL + "/", AccType: "webdav", AccSync: true,
		SyncDownload: true, SyncPath: "/.photos", SyncStatus: entity.SyncStatusRefresh, AccTimeout: "low", RetryLimit: 1}
	require.NoError(t, entity.Db().Create(&a).Error)
	t.Cleanup(func() {
		assert.NoError(t, entity.UnscopedDb().Unscoped().Delete(&entity.FileSync{}, "service_id = ?", a.ID).Error)
		assert.NoError(t, entity.UnscopedDb().Unscoped().Delete(&entity.Service{}, a.ID).Error)
	})

	// stored returns the stored account.
	stored := func() (m entity.Service) {
		require.NoError(t, entity.UnscopedDb().First(&m, a.ID).Error)
		return m
	}

	// Several cycles keep the notice without counting it, so the retry limit never turns sync off.
	for i := 0; i < 3; i++ {
		require.NoError(t, entity.Db().Model(&a).UpdateColumn("sync_status", entity.SyncStatusRefresh).Error)
		require.NoError(t, worker.Start())

		m := stored()
		assert.Equal(t, "sync folder is not supported", m.AccError, i)
		assert.Equal(t, 0, m.AccErrors, i)
		assert.True(t, m.AccSync, i)
	}

	// A supported sync folder clears it.
	require.NoError(t, entity.Db().Model(&a).UpdateColumns(entity.Values{"sync_path": "/", "sync_status": entity.SyncStatusRefresh}).Error)
	require.NoError(t, worker.Start())

	m := stored()
	assert.Equal(t, "", m.AccError)
	assert.Equal(t, 0, m.AccErrors)
}
