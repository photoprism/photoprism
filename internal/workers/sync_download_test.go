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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
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
}

func TestSync_downloadRetryLimit(t *testing.T) {
	conf := config.TestConfig()
	worker := NewSync(conf)

	var mu sync.Mutex
	requested := make(map[string]int)

	// The remote answers every request with an error and counts the downloads, so nothing is
	// imported and a skipped file is one that was never requested.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			mu.Lock()
			requested[r.URL.Path]++
			mu.Unlock()
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
		// The related file sorts after the skipped one in the same group, so it shows that the
		// remaining files are still attempted.
		a := newAccount(t, 3)
		newFile(t, a, "/below.jpg", 2)
		newFile(t, a, "/at.jpg", 3)
		newFile(t, a, "/above.jpg", 4)
		newFile(t, a, "/above.png", 2)

		_, err := worker.download(a)
		require.NoError(t, err)

		assert.Equal(t, 1, attempts("/below.jpg"))
		assert.Equal(t, 1, attempts("/at.jpg"))
		assert.Equal(t, 1, attempts("/above.png"))

		above := stored(t, a, "/above.jpg")
		assert.Equal(t, 0, attempts("/above.jpg"))
		assert.Equal(t, 4, above.Errors)
		assert.Equal(t, entity.FileSyncNew, above.Status)
	})
	t.Run("LowestLimit", func(t *testing.T) {
		a := newAccount(t, 1)
		newFile(t, a, "/lowest.jpg", 2)

		_, err := worker.download(a)
		require.NoError(t, err)

		assert.Equal(t, 0, attempts("/lowest.jpg"))
		assert.Equal(t, entity.FileSyncNew, stored(t, a, "/lowest.jpg").Status)
	})
	t.Run("HigherLimit", func(t *testing.T) {
		a := newAccount(t, 5)
		newFile(t, a, "/higher.jpg", 4)

		_, err := worker.download(a)
		require.NoError(t, err)

		assert.Equal(t, 1, attempts("/higher.jpg"))
	})
	t.Run("NoLimit", func(t *testing.T) {
		// A limit of -1 disables the check, so the file is requested regardless of its error count.
		a := newAccount(t, -1)
		newFile(t, a, "/unlimited.jpg", 100)

		_, err := worker.download(a)
		require.NoError(t, err)

		assert.Equal(t, 1, attempts("/unlimited.jpg"))
	})
}
