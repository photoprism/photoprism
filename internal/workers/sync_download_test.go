package workers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestSync_download(t *testing.T) {
	t.Run("NotFound", func(t *testing.T) {
		conf := config.TestConfig()

		t.Logf("database-dsn: %s", conf.DatabaseDSN())

		worker := NewSync(conf)

		assert.IsType(t, &Sync{}, worker)
		account := entity.ServiceFixtureWebdavDummy

		if complete, err := worker.download(account); err != nil {
			t.Fatal(err)
		} else {
			assert.False(t, complete)

		}
	})
	t.Run("QuotaExceeded", func(t *testing.T) {
		conf := config.TestConfig()
		conf.Options().FilesQuota = 1

		t.Logf("database-dsn: %s", conf.DatabaseDSN())

		worker := NewSync(conf)

		assert.IsType(t, &Sync{}, worker)
		account := entity.ServiceFixtureWebdavDummy

		if complete, err := worker.download(account); err != nil {
			t.Fatal(err)
		} else {
			assert.False(t, complete)

		}
		conf.Options().FilesQuota = 0
	})
}

func TestSync_downloadPath(t *testing.T) {
	conf := config.TestConfig()

	worker := NewSync(conf)

	assert.IsType(t, &Sync{}, worker)
	assert.True(t, strings.HasSuffix(worker.downloadPath(), "testdata/temp/sync"))
}

func TestSync_relatedDownloads(t *testing.T) {
	conf := config.TestConfig()

	worker := NewSync(conf)
	account := entity.ServiceFixtureWebdavDummy

	assert.IsType(t, &Sync{}, worker)

	if result, err := worker.relatedDownloads(account); err != nil {
		t.Fatal(err)
	} else {
		assert.IsType(t, Downloads{}, result)
	}

	t.Run("FailedFilesLast", func(t *testing.T) {
		// More failed files than one batch holds sort before the new file by name.
		a := entity.Service{AccName: "Sync Related " + rnd.Base36(8), AccType: "webdav", AccSync: true, SyncDownload: true, RetryLimit: -1}
		require.NoError(t, entity.Db().Create(&a).Error)
		t.Cleanup(func() {
			assert.NoError(t, entity.UnscopedDb().Unscoped().Delete(&entity.FileSync{}, "service_id = ?", a.ID).Error)
			assert.NoError(t, entity.UnscopedDb().Unscoped().Delete(&entity.Service{}, a.ID).Error)
		})

		tx := entity.Db().Begin()
		for i := 0; i < 960; i++ {
			f := entity.NewFileSync(a.ID, fmt.Sprintf("/failed-%04d.jpg", i))
			f.Status, f.Errors = entity.FileSyncNew, 1
			if err := tx.Create(f).Error; err != nil {
				tx.Rollback()
				t.Fatal(err)
			}
		}
		require.NoError(t, tx.Commit().Error)

		f := entity.NewFileSync(a.ID, "/zz-new.jpg")
		f.Status = entity.FileSyncNew
		require.NoError(t, f.Create())

		result, err := worker.relatedDownloads(a)
		require.NoError(t, err)
		assert.Contains(t, result, "/zz-new")
	})
}

func TestSync_downloadRetryLimit(t *testing.T) {
	conf := config.TestConfig()
	worker := NewSync(conf)

	var mu sync.Mutex
	requested := make(map[string]int)

	// The remote answers every request except text files below /ok with an error and counts the
	// downloads, so nothing is imported and a skipped file is one that was never requested.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			mu.Lock()
			requested[r.URL.Path]++
			mu.Unlock()

			if strings.HasPrefix(r.URL.Path, "/ok") {
				_, _ = w.Write([]byte("ok"))
				return
			}
		}

		w.WriteHeader(http.StatusServiceUnavailable)
	}))

	t.Cleanup(server.Close)

	// newAccount creates a download account with the specified retry limit on the test server.
	newAccount := func(t *testing.T, limit int) entity.Service {
		a := entity.Service{AccName: "Sync Download " + rnd.Base36(8), AccURL: server.URL + "/", AccType: "webdav",
			AccSync: true, SyncDownload: true, SyncPath: "/", AccTimeout: "low", RetryLimit: limit}
		require.NoError(t, entity.Db().Create(&a).Error)

		t.Cleanup(func() {
			if err := entity.UnscopedDb().Unscoped().Delete(&entity.FileSync{}, "service_id = ?", a.ID).Error; err != nil {
				t.Errorf("delete files of service %d: %s", a.ID, err)
			}

			if err := entity.UnscopedDb().Unscoped().Delete(&entity.Service{}, a.ID).Error; err != nil {
				t.Errorf("delete service %d: %s", a.ID, err)
			}

			_ = os.RemoveAll(filepath.Join(worker.downloadPath(), fmt.Sprintf("%d", a.ID)))
		})

		return a
	}

	// newFile adds a remote file in the "new" state with the specified error count.
	newFile := func(t *testing.T, a entity.Service, name string, count int) {
		f := entity.NewFileSync(a.ID, name)
		f.Status = entity.FileSyncNew
		f.Errors = count
		require.NoError(t, f.Create())
	}

	// stored returns the stored record of a remote file.
	stored := func(t *testing.T, a entity.Service, name string) entity.FileSync {
		var f entity.FileSync
		require.NoError(t, entity.Db().Where("service_id = ? AND remote_name = ?", a.ID, name).First(&f).Error)
		return f
	}

	// attempts returns how often a remote file was requested.
	attempts := func(name string) int {
		mu.Lock()
		defer mu.Unlock()
		return requested[name]
	}

	t.Run("Limit", func(t *testing.T) {
		// Files with more errors sort after the skipped one in its group, so the ignored YAML file
		// shows that the remaining files are still visited.
		a := newAccount(t, 3)
		require.NoError(t, a.Update("sync_yaml", -1))
		newFile(t, a, "/below.jpg", 2)
		newFile(t, a, "/at.jpg", 3)
		newFile(t, a, "/above.jpg", 4)
		newFile(t, a, "/above.png", 2)
		newFile(t, a, "/above.yml", 5)

		complete, err := worker.download(a)
		require.NoError(t, err)
		assert.False(t, complete)

		assert.Equal(t, 1, attempts("/below.jpg"))
		assert.Equal(t, 1, attempts("/at.jpg"))
		assert.Equal(t, 1, attempts("/above.png"))

		above := stored(t, a, "/above.jpg")
		assert.Equal(t, 0, attempts("/above.jpg"))
		assert.Equal(t, 4, above.Errors)
		assert.Equal(t, entity.FileSyncNew, above.Status)
		assert.Equal(t, entity.FileSyncIgnore, stored(t, a, "/above.yml").Status)

		// A failure that stays within the limit keeps the file queued, the next one marks it failed.
		below := stored(t, a, "/below.jpg")
		assert.Equal(t, 3, below.Errors)
		assert.Equal(t, entity.FileSyncNew, below.Status)
		at := stored(t, a, "/at.jpg")
		assert.Equal(t, 4, at.Errors)
		assert.Equal(t, entity.FileSyncFailed, at.Status)
	})
	t.Run("LowestLimit", func(t *testing.T) {
		a := newAccount(t, 1)
		newFile(t, a, "/lowest.jpg", 2)
		newFile(t, a, "/lowest-once.jpg", 1)

		_, err := worker.download(a)
		require.NoError(t, err)

		assert.Equal(t, 0, attempts("/lowest.jpg"))
		assert.Equal(t, entity.FileSyncNew, stored(t, a, "/lowest.jpg").Status)
		assert.Equal(t, 1, attempts("/lowest-once.jpg"))
		assert.Equal(t, entity.FileSyncFailed, stored(t, a, "/lowest-once.jpg").Status)
	})
	t.Run("HigherLimit", func(t *testing.T) {
		// A failure within a positive limit keeps the stage open for the next run.
		a := newAccount(t, 5)
		newFile(t, a, "/higher.jpg", 4)

		complete, err := worker.download(a)
		require.NoError(t, err)
		assert.False(t, complete)

		assert.Equal(t, 1, attempts("/higher.jpg"))
	})
	t.Run("NoLimit", func(t *testing.T) {
		// A limit of -1 disables the check, so the file is requested regardless of its error count
		// and stays queued after each failure.
		a := newAccount(t, -1)
		newFile(t, a, "/unlimited.jpg", 100)

		for run := 1; run <= 2; run++ {
			complete, err := worker.download(a)
			require.NoError(t, err)
			assert.True(t, complete)

			f := stored(t, a, "/unlimited.jpg")
			assert.Equal(t, run, attempts("/unlimited.jpg"))
			assert.Equal(t, 100+run, f.Errors)
			assert.Equal(t, entity.FileSyncNew, f.Status)
		}
	})
	t.Run("ZeroLimit", func(t *testing.T) {
		a := newAccount(t, 0)
		newFile(t, a, "/zero.jpg", 0)

		for run := 1; run <= 2; run++ {
			complete, err := worker.download(a)
			require.NoError(t, err)
			assert.True(t, complete)

			f := stored(t, a, "/zero.jpg")
			assert.Equal(t, run, attempts("/zero.jpg"))
			assert.Equal(t, run, f.Errors)
			assert.Equal(t, entity.FileSyncNew, f.Status)
		}
	})
	t.Run("NoLimitProgress", func(t *testing.T) {
		// A run that downloads a file is not complete, the next one with only failures left is.
		a := newAccount(t, -1)
		newFile(t, a, "/ok-progress.txt", 0)
		newFile(t, a, "/stuck.jpg", 0)
		newFile(t, a, "/stuck.png", 0)

		logger, hook := test.NewNullLogger()
		prev := log
		log = logger
		t.Cleanup(func() { log = prev })

		complete, err := worker.download(a)
		require.NoError(t, err)
		assert.False(t, complete)
		assert.Equal(t, entity.FileSyncDownloaded, stored(t, a, "/ok-progress.txt").Status)
		assert.Equal(t, entity.FileSyncNew, stored(t, a, "/stuck.jpg").Status)
		for _, e := range hook.AllEntries() {
			assert.NotEqual(t, logrus.WarnLevel, e.Level, e.Message)
		}

		complete, err = worker.download(a)
		require.NoError(t, err)
		assert.True(t, complete)
		require.NotNil(t, hook.LastEntry())
		assert.Equal(t, logrus.WarnLevel, hook.LastEntry().Level)
		assert.Contains(t, hook.LastEntry().Message, "failed to download 2 files from")
		assert.Equal(t, 1, attempts("/ok-progress.txt"))
		assert.Equal(t, 2, attempts("/stuck.jpg"))
		assert.Equal(t, entity.FileSyncNew, stored(t, a, "/stuck.jpg").Status)
	})
	t.Run("LowestLimitOpen", func(t *testing.T) {
		a := newAccount(t, 1)
		newFile(t, a, "/lowest-open.jpg", 0)

		logger, hook := test.NewNullLogger()
		prev := log
		log = logger
		t.Cleanup(func() { log = prev })

		complete, err := worker.download(a)
		require.NoError(t, err)
		assert.False(t, complete)
		assert.Equal(t, entity.FileSyncNew, stored(t, a, "/lowest-open.jpg").Status)
		for _, e := range hook.AllEntries() {
			assert.NotEqual(t, logrus.WarnLevel, e.Level, e.Message)
		}
	})
	t.Run("NoLimitIgnored", func(t *testing.T) {
		// An ignored file counts as a change, like a download.
		a := newAccount(t, -1)
		require.NoError(t, a.Update("sync_yaml", -1))
		newFile(t, a, "/ignored.yml", 0)
		newFile(t, a, "/ignored.jpg", 0)

		complete, err := worker.download(a)
		require.NoError(t, err)
		assert.False(t, complete)
		assert.Equal(t, entity.FileSyncIgnore, stored(t, a, "/ignored.yml").Status)

		complete, err = worker.download(a)
		require.NoError(t, err)
		assert.True(t, complete)
		assert.Equal(t, 0, attempts("/ignored.yml"))
		assert.Equal(t, 2, attempts("/ignored.jpg"))
	})
	t.Run("NoLimitSaveKeepsOrder", func(t *testing.T) {
		// A run where every file fails leaves the first window with one error each, and saving the
		// account in between keeps those counts, so the next run reads the files after them.
		a := newAccount(t, -1)
		tx := entity.Db().Begin()
		for i := 0; i < 960; i++ {
			f := entity.NewFileSync(a.ID, fmt.Sprintf("/order-%04d.jpg", i))
			f.Status = entity.FileSyncNew
			if i < 952 {
				f.Errors, f.Error = 1, "failed"
			}
			if err := tx.Create(f).Error; err != nil {
				tx.Rollback()
				t.Fatal(err)
			}
		}
		require.NoError(t, tx.Commit().Error)
		newFile(t, a, "/zz-order.jpg", 0)

		f, err := form.NewService(a)
		require.NoError(t, err)
		require.NoError(t, a.SaveForm(f))
		assert.Equal(t, 1, stored(t, a, "/order-0000.jpg").Errors)
		assert.Equal(t, "", stored(t, a, "/order-0000.jpg").Error)

		result, err := worker.relatedDownloads(a)
		require.NoError(t, err)
		assert.Contains(t, result, "/zz-order")
	})
	t.Run("NoLimitStart", func(t *testing.T) {
		// The worker moves past the download stage while the failed file stays queued.
		isolateSyncAccounts(t)

		a := newAccount(t, -1)
		require.NoError(t, a.Update("sync_status", entity.SyncStatusDownload))
		newFile(t, a, "/start.jpg", 0)

		require.NoError(t, worker.Start())

		s := storedSyncAccount(t, a.ID)
		assert.Equal(t, entity.SyncStatusSynced, s.SyncStatus)
		assert.True(t, s.SyncDate.Valid)
		assert.Equal(t, 1, attempts("/start.jpg"))
		assert.Equal(t, entity.FileSyncNew, stored(t, a, "/start.jpg").Status)
	})
}
