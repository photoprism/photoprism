package workers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// TestSync_downloadFaceMigration pins that downloads wait for a face migration and stay queued, so the
// next sync run downloads them, while a run outside a migration downloads as before.
func TestSync_downloadFaceMigration(t *testing.T) {
	conf := config.TestConfig()
	worker := NewSync(conf)

	var mu sync.Mutex
	requested := 0

	// lockOnRequest, when set, is called on the first download request, as a migration starting then would.
	var lockOnRequest func()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			mu.Lock()
			requested++
			onRequest := lockOnRequest
			lockOnRequest = nil
			mu.Unlock()

			if onRequest != nil {
				onRequest()
			}
		}

		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	t.Cleanup(server.Close)

	requests := func() int {
		mu.Lock()
		defer mu.Unlock()
		return requested
	}

	// newAccount creates a download account on the test server with the specified queued remote files.
	newAccount := func(t *testing.T, names ...string) entity.Service {
		a := entity.Service{AccName: "Sync Migration " + rnd.Base36(8), AccURL: server.URL + "/", AccType: "webdav",
			AccSync: true, SyncDownload: true, SyncPath: "/", AccTimeout: "low"}
		require.NoError(t, entity.Db().Create(&a).Error)

		t.Cleanup(func() {
			_ = entity.UnscopedDb().Unscoped().Delete(&entity.FileSync{}, "service_id = ?", a.ID).Error
			_ = entity.UnscopedDb().Unscoped().Delete(&entity.Service{}, a.ID).Error
			_ = os.RemoveAll(filepath.Join(worker.downloadPath(), strconv.Itoa(int(a.ID))))
		})

		for _, name := range names {
			f := entity.NewFileSync(a.ID, name)
			f.Status = entity.FileSyncNew
			require.NoError(t, f.Create())
		}

		return a
	}

	// stubWorkers replaces the index and import workers with ones that only drain their jobs.
	stubWorkers := func(t *testing.T) {
		prevIndex, prevImport := syncIndexWorker, syncImportWorker
		t.Cleanup(func() { syncIndexWorker, syncImportWorker = prevIndex, prevImport })

		syncIndexWorker = func(jobs <-chan photoprism.IndexJob) {
			for range jobs {
			}
		}
		syncImportWorker = func(jobs <-chan photoprism.ImportJob) {
			for range jobs {
			}
		}
	}

	t.Run("LockedBeforeTheRun", func(t *testing.T) {
		lock, err := mutex.AcquireFileLock(conf.FacesLockFile(), "faces migration")
		require.NoError(t, err)
		t.Cleanup(lock.Release)

		stubWorkers(t)
		hook := captureSyncLog(t)
		before := requests()

		complete, err := worker.download(newAccount(t, "/queued.jpg"))
		require.NoError(t, err)
		assert.False(t, complete, "the account stays in the download stage")
		assert.Equal(t, before, requests(), "nothing is downloaded")

		// Refused before the run starts, rather than by the check between groups.
		assert.Zero(t, countLogged(hook, logrus.InfoLevel, "sync: downloading from"))
		assert.Equal(t, 1, countLogged(hook, logrus.InfoLevel, "waiting for the faces migration"))
	})
	t.Run("LockedDuringTheRun", func(t *testing.T) {
		// A migration that starts with the first download stops the run before the next group.
		var held atomic.Pointer[mutex.FileLock]
		t.Cleanup(func() { held.Load().Release() })

		mu.Lock()
		lockOnRequest = func() {
			lock, err := mutex.AcquireFileLock(conf.FacesLockFile(), "faces migration")
			assert.NoError(t, err)
			held.Store(lock)
		}
		mu.Unlock()

		stubWorkers(t)
		a := newAccount(t, "/first.jpg", "/second.jpg")
		before := requests()

		complete, err := worker.download(a)
		require.NoError(t, err)
		assert.False(t, complete)
		assert.Equal(t, before+1, requests(), "only the first group is requested")
	})
	t.Run("NothingQueued", func(t *testing.T) {
		// An account with nothing to download completes the stage, since nothing is indexed.
		lock, err := mutex.AcquireFileLock(conf.FacesLockFile(), "faces migration")
		require.NoError(t, err)
		t.Cleanup(lock.Release)

		stubWorkers(t)

		complete, err := worker.download(newAccount(t))
		require.NoError(t, err)
		assert.True(t, complete)
	})
	t.Run("NotLocked", func(t *testing.T) {
		require.Empty(t, conf.FacesLocked())
		stubWorkers(t)
		before := requests()

		_, err := worker.download(newAccount(t, "/first.jpg", "/second.jpg"))
		require.NoError(t, err)
		assert.Equal(t, before+2, requests(), "both groups are requested")
	})
}
