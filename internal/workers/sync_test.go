package workers

import (
	"database/sql"
	"errors"
	"math"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
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
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/txt"
)

func TestNewSync(t *testing.T) {
	conf := config.TestConfig()

	worker := NewSync(conf)

	assert.IsType(t, &Sync{}, worker)
}

func TestSync_Start(t *testing.T) {
	conf := config.TestConfig()

	worker := NewSync(conf)

	assert.IsType(t, &Sync{}, worker)

	if err := mutex.SyncWorker.Start(); err != nil {
		t.Fatal(err)
	}

	if err := worker.Start(); err == nil {
		t.Fatal("error expected")
	}

	mutex.SyncWorker.Stop()

	if err := worker.Start(); err != nil {
		t.Fatal(err)
	}
}

func TestSyncRefresh_RemoteErrorText(t *testing.T) {
	// A remote server chooses its own error body, which reaches the log through the sync worker.
	// The rendered line must stay one line, in one field, whatever the body contains.
	remote := errors.New("500 Internal Server Error: denied\nsync: › admin › granted\u202e \x07")

	rendered := clean.Error(remote)

	assert.NotContains(t, rendered, "\n", "a remote body must not add a line")
	assert.NotContains(t, rendered, "\r", "a remote body must not add a line")
	assert.NotContains(t, rendered, "›", "a remote body must not add a field separator")
	assert.NotContains(t, rendered, "\u202e", "a remote body must not carry a bidi override")
	assert.NotContains(t, rendered, "\x07", "a remote body must not carry a control character")
	assert.Contains(t, rendered, "denied", "the reason must survive")
}

func TestSyncRefresh_RemoteCredentials(t *testing.T) {
	// A sync account URL may carry credentials, and the remote error repeats the URL it failed on.
	//nolint:gosec // G101: Example credential in a fixture URL, which is the subject of the test.
	remote := &url.Error{Op: "PROPFIND", URL: "https://sync-user:notreal@dav.example.com/photos", Err: errors.New("timeout")}

	rendered := clean.Error(remote)

	assert.NotContains(t, rendered, "notreal", "the credential must not survive")
}

// isolateSyncAccounts disables sync for every existing account until the test ends, so Sync.Start
// processes only the accounts the test creates and makes no network requests for the others.
func isolateSyncAccounts(t *testing.T) {
	t.Helper()

	var ids, disabled []uint

	require.NoError(t, entity.Db().Model(&entity.Service{}).Where("acc_sync = TRUE").Pluck("id", &ids).Error)

	t.Cleanup(func() {
		for _, id := range disabled {
			if err := entity.Db().Model(&entity.Service{ID: id}).UpdateColumn("acc_sync", true).Error; err != nil {
				t.Errorf("restore sync for service %d: %s", id, err)
			}
		}
	})

	for _, id := range ids {
		require.NoError(t, entity.Db().Model(&entity.Service{ID: id}).UpdateColumn("acc_sync", false).Error)
		disabled = append(disabled, id)
	}
}

// newSyncAccount creates a WebDAV account with sync enabled in the synced state, applies update,
// and deletes it when the test ends. Its URL is never contacted in that state.
func newSyncAccount(t *testing.T, update func(a *entity.Service)) *entity.Service {
	t.Helper()

	a := &entity.Service{
		AccName:      "Sync Start " + rnd.Base36(8),
		AccURL:       "http://127.0.0.1:1/",
		AccType:      "webdav",
		AccSync:      true,
		RetryLimit:   3,
		SyncStatus:   entity.SyncStatusSynced,
		SyncInterval: 3600,
		SyncDate:     sql.NullTime{Time: time.Now().Add(-10 * time.Minute), Valid: true},
	}

	if update != nil {
		update(a)
	}

	require.NoError(t, entity.Db().Create(a).Error)

	t.Cleanup(func() {
		if err := entity.UnscopedDb().Unscoped().Delete(&entity.Service{}, a.ID).Error; err != nil {
			t.Errorf("delete service %d: %s", a.ID, err)
		}
	})

	return a
}

// loggedWarning reports whether a warning containing text was logged.
func loggedWarning(hook *test.Hook, text string) bool {
	for _, e := range hook.AllEntries() {
		if e.Level == logrus.WarnLevel && strings.Contains(e.Message, text) {
			return true
		}
	}

	return false
}

// storedSyncAccount returns the account as currently stored in the database.
func storedSyncAccount(t *testing.T, id uint) entity.Service {
	t.Helper()

	var m entity.Service

	require.NoError(t, entity.Db().First(&m, id).Error)

	return m
}

func TestSync_StartInterval(t *testing.T) {
	isolateSyncAccounts(t)

	worker := NewSync(config.TestConfig())

	t.Run("OlderThanInterval", func(t *testing.T) {
		a := newSyncAccount(t, func(a *entity.Service) {
			a.SyncDate = sql.NullTime{Time: time.Now().Add(-2 * time.Hour), Valid: true}
		})

		require.NoError(t, worker.Start())
		assert.Equal(t, entity.SyncStatusRefresh, storedSyncAccount(t, a.ID).SyncStatus)
	})
	t.Run("NewerThanInterval", func(t *testing.T) {
		a := newSyncAccount(t, func(a *entity.Service) {
			a.SyncDate = sql.NullTime{Time: time.Now().Add(-10 * time.Minute), Valid: true}
		})

		require.NoError(t, worker.Start())
		assert.Equal(t, entity.SyncStatusSynced, storedSyncAccount(t, a.ID).SyncStatus)
	})
	t.Run("ShorterInterval", func(t *testing.T) {
		a := newSyncAccount(t, func(a *entity.Service) {
			a.SyncInterval = 60
			a.SyncDate = sql.NullTime{Time: time.Now().Add(-10 * time.Minute), Valid: true}
		})

		require.NoError(t, worker.Start())
		assert.Equal(t, entity.SyncStatusRefresh, storedSyncAccount(t, a.ID).SyncStatus)
	})
	t.Run("NoSyncDate", func(t *testing.T) {
		a := newSyncAccount(t, func(a *entity.Service) {
			a.SyncDate = sql.NullTime{}
		})

		require.NoError(t, worker.Start())

		stored := storedSyncAccount(t, a.ID)
		assert.Equal(t, entity.SyncStatusSynced, stored.SyncStatus)
		assert.False(t, stored.SyncDate.Valid)
	})
	t.Run("NeverInterval", func(t *testing.T) {
		a := newSyncAccount(t, func(a *entity.Service) {
			a.SyncInterval = 0
			a.SyncDate = sql.NullTime{Time: time.Now().Add(-30 * 24 * time.Hour), Valid: true}
		})

		require.NoError(t, worker.Start())

		stored := storedSyncAccount(t, a.ID)
		assert.Equal(t, entity.SyncStatusSynced, stored.SyncStatus)

		// Saving the account still starts a new sync.
		f, err := form.NewService(stored)
		require.NoError(t, err)
		require.NoError(t, stored.SaveForm(f))
		assert.Equal(t, entity.SyncStatusRefresh, storedSyncAccount(t, a.ID).SyncStatus)
	})
	t.Run("NegativeInterval", func(t *testing.T) {
		a := newSyncAccount(t, func(a *entity.Service) {
			a.SyncInterval = -3600
			a.SyncDate = sql.NullTime{Time: time.Now().Add(-30 * 24 * time.Hour), Valid: true}
		})

		require.NoError(t, worker.Start())
		assert.Equal(t, entity.SyncStatusSynced, storedSyncAccount(t, a.ID).SyncStatus)
	})
}

func TestSyncDue(t *testing.T) {
	now := time.Date(2026, 9, 30, 12, 0, 0, 0, time.UTC)
	at := func(d time.Duration) sql.NullTime { return sql.NullTime{Time: now.Add(-d), Valid: true} }

	// The largest interval a time.Duration can hold, computed so the table also compiles with a 32-bit int.
	maxDuration := int64(math.MaxInt64) / int64(time.Second)

	for _, c := range []struct {
		name     string
		interval int
		date     sql.NullTime
		due      bool
	}{
		{"OlderThanInterval", 3600, at(2 * time.Hour), true},
		{"NewerThanInterval", 3600, at(10 * time.Minute), false},
		{"ExactlyInterval", 3600, at(time.Hour), false},
		{"Never", 0, at(365 * 24 * time.Hour), false},
		{"Negative", -3600, at(365 * 24 * time.Hour), false},
		{"NoDate", 3600, sql.NullTime{}, false},
		{"FutureDate", 3600, at(-time.Hour), false},
		{"LongerThanMaxDuration", int(maxDuration + 1), at(365 * 24 * time.Hour), false},
		{"MaxInt", math.MaxInt, at(365 * 24 * time.Hour), false},
		{"MaxClamp", 31536000, at(366 * 24 * time.Hour), true},
	} {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.due, syncDue(entity.Service{SyncInterval: c.interval, SyncDate: c.date}, now))
		})
	}
}

func TestSync_StartStoredError(t *testing.T) {
	isolateSyncAccounts(t)

	worker := NewSync(config.TestConfig())

	// The remote answers the listing with an error whose text holds a URL with credentials, a new line, a field
	// separator and more.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("listing failed https://sync-user:notreal@dav.example.com/photos\nsync \u203a admin " + strings.Repeat("x", 800)))
	}))
	t.Cleanup(server.Close)

	a := newSyncAccount(t, func(a *entity.Service) {
		a.AccURL = server.URL + "/"
		a.AccTimeout = "low"
		a.SyncStatus = entity.SyncStatusRefresh
	})

	logger, hook := test.NewNullLogger()
	prev := log
	log = logger
	t.Cleanup(func() { log = prev })

	require.NoError(t, worker.Start())

	// The failure names the account.
	var named int
	for _, e := range hook.AllEntries() {
		if e.Level == logrus.ErrorLevel && strings.Contains(e.Message, "(list remote folders of "+clean.Log(a.AccName)+")") {
			named++
		}
	}
	assert.Equal(t, 1, named)

	stored := storedSyncAccount(t, a.ID)
	assert.Equal(t, 1, stored.AccErrors)
	assert.NotEmpty(t, stored.AccError)
	assert.LessOrEqual(t, len(stored.AccError), txt.ClipError)
	assert.NotContains(t, stored.AccError, "\n")
	assert.NotContains(t, stored.AccError, "\u203a")
	assert.NotContains(t, stored.AccError, "notreal")
}

func TestSync_StartRetryLimit(t *testing.T) {
	isolateSyncAccounts(t)

	worker := NewSync(config.TestConfig())

	// withErrors returns an account update that sets the retry limit and the recorded error count.
	withErrors := func(limit, count int) func(a *entity.Service) {
		return func(a *entity.Service) {
			a.RetryLimit = limit
			a.AccErrors = count
		}
	}

	t.Run("BelowLimit", func(t *testing.T) {
		a := newSyncAccount(t, withErrors(3, 2))

		require.NoError(t, worker.Start())
		assert.True(t, storedSyncAccount(t, a.ID).AccSync)
	})
	t.Run("AtLimit", func(t *testing.T) {
		a := newSyncAccount(t, withErrors(3, 3))

		require.NoError(t, worker.Start())
		assert.True(t, storedSyncAccount(t, a.ID).AccSync)
	})
	t.Run("AboveLimit", func(t *testing.T) {
		// Both accounts are due for a refresh, which the disabled account must not get in the same
		// run, while the account sorted after it must still be processed.
		due := sql.NullTime{Time: time.Now().Add(-2 * time.Hour), Valid: true}
		a := newSyncAccount(t, func(a *entity.Service) {
			withErrors(3, 4)(a)
			a.AccName = "Sync Start A " + rnd.Base36(8)
			a.SyncDate = due
		})
		other := newSyncAccount(t, func(a *entity.Service) {
			withErrors(3, 2)(a)
			a.AccName = "Sync Start B " + rnd.Base36(8)
			a.SyncDate = due
		})

		// An account the worker skips before the check keeps sync on, although it is over its limit too.
		skipped := newSyncAccount(t, func(a *entity.Service) {
			withErrors(3, 4)(a)
			a.AccType = "test"
		})

		logger, hook := test.NewNullLogger()
		prev := log
		log = logger
		t.Cleanup(func() { log = prev })

		require.NoError(t, worker.Start())

		stored := storedSyncAccount(t, a.ID)
		assert.False(t, stored.AccSync)
		assert.Equal(t, entity.SyncStatusSynced, stored.SyncStatus)
		assert.True(t, loggedWarning(hook, "disabled sync"))
		assert.True(t, storedSyncAccount(t, skipped.ID).AccSync)

		storedOther := storedSyncAccount(t, other.ID)
		assert.True(t, storedOther.AccSync)
		assert.Equal(t, entity.SyncStatusRefresh, storedOther.SyncStatus)
	})
	t.Run("AboveLimitKeepsChanges", func(t *testing.T) {
		// The remote of the account processed first changes the second one in the database after
		// Start() has loaded it, so the switch-off must write its own column only.
		var other *entity.Service
		var once sync.Once

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			once.Do(func() {
				assert.NoError(t, entity.Db().Model(&entity.Service{ID: other.ID}).
					Updates(entity.Values{"sync_path": "/changed", "acc_share": true}).Error)
			})
			w.WriteHeader(http.StatusServiceUnavailable)
		}))
		t.Cleanup(server.Close)

		first := newSyncAccount(t, func(a *entity.Service) {
			a.AccName = "Sync Start A " + rnd.Base36(8)
			a.AccURL = server.URL + "/"
			a.AccTimeout = "low"
			a.SyncStatus = entity.SyncStatusRefresh
		})
		other = newSyncAccount(t, func(a *entity.Service) {
			withErrors(3, 4)(a)
			a.AccName = "Sync Start B " + rnd.Base36(8)
			a.SyncPath = "/"
		})

		require.NoError(t, worker.Start())

		assert.Equal(t, 1, storedSyncAccount(t, first.ID).AccErrors)

		stored := storedSyncAccount(t, other.ID)
		assert.False(t, stored.AccSync)
		assert.True(t, stored.AccShare)
		assert.Equal(t, "/changed", stored.SyncPath)
		assert.Equal(t, entity.SyncStatusSynced, stored.SyncStatus)
		assert.Equal(t, 4, stored.AccErrors)
	})
	t.Run("AboveLimitReset", func(t *testing.T) {
		// The remote of the account processed first changes the second one after Start() has loaded
		// it, as saving the account does, so sync stays on.
		// The account is skipped in that run either way, so its loaded copy never overwrites the change.
		for _, c := range []struct {
			name   string
			values entity.Values
			sync   bool
			errors int
		}{
			{"ErrorsReset", entity.Values{"acc_errors": 0}, true, 0},
			{"ErrorsAtLimit", entity.Values{"acc_errors": 3}, true, 3},
			{"LimitRemoved", entity.Values{"retry_limit": -1}, true, 4},
			{"SyncTurnedOff", entity.Values{"acc_sync": false}, false, 4},
		} {
			t.Run(c.name, func(t *testing.T) {
				var other *entity.Service
				var once sync.Once

				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					once.Do(func() {
						assert.NoError(t, entity.Db().Model(&entity.Service{ID: other.ID}).UpdateColumns(c.values).Error)
					})
					w.WriteHeader(http.StatusServiceUnavailable)
				}))
				t.Cleanup(server.Close)

				first := newSyncAccount(t, func(a *entity.Service) {
					a.AccName = "Sync Start A " + rnd.Base36(8)
					a.AccURL = server.URL + "/"
					a.AccTimeout = "low"
					a.SyncStatus = entity.SyncStatusRefresh
				})
				// Due for a refresh, so processing it would change its status.
				other = newSyncAccount(t, func(a *entity.Service) {
					withErrors(3, 4)(a)
					a.AccName = "Sync Start B " + rnd.Base36(8)
					a.SyncDate = sql.NullTime{Time: time.Now().Add(-2 * time.Hour), Valid: true}
				})

				logger, hook := test.NewNullLogger()
				prev := log
				log = logger
				t.Cleanup(func() { log = prev })

				require.NoError(t, worker.Start())

				assert.Equal(t, 1, storedSyncAccount(t, first.ID).AccErrors)
				stored := storedSyncAccount(t, other.ID)
				assert.Equal(t, c.sync, stored.AccSync)
				assert.Equal(t, c.errors, stored.AccErrors)
				assert.Equal(t, entity.SyncStatusSynced, stored.SyncStatus)
				assert.False(t, loggedWarning(hook, "disabled sync"))
			})
		}
	})
	t.Run("HigherLimit", func(t *testing.T) {
		a := newSyncAccount(t, withErrors(5, 4))

		require.NoError(t, worker.Start())
		assert.True(t, storedSyncAccount(t, a.ID).AccSync)
	})
	t.Run("NoLimit", func(t *testing.T) {
		a := newSyncAccount(t, withErrors(-1, 100))

		require.NoError(t, worker.Start())
		assert.True(t, storedSyncAccount(t, a.ID).AccSync)
	})
}

func TestSync_StartKeepsChanges(t *testing.T) {
	isolateSyncAccounts(t)

	worker := NewSync(config.TestConfig())

	// newChangingAccount creates an account whose remote fails every request and, on the first one, stores
	// values for its own row, as saving the account during the run does.
	newChangingAccount := func(t *testing.T, status string, values entity.Values) *entity.Service {
		var a *entity.Service
		var once sync.Once

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if values != nil {
				once.Do(func() {
					assert.NoError(t, entity.Db().Model(&entity.Service{ID: a.ID}).UpdateColumns(values).Error)
				})
			}
			w.WriteHeader(http.StatusServiceUnavailable)
		}))
		t.Cleanup(server.Close)

		a = newSyncAccount(t, func(m *entity.Service) {
			m.AccURL = server.URL + "/"
			m.AccTimeout = "low"
			m.SyncStatus = status
			m.SyncDownload = true
			m.AccErrors, m.AccError = 2, "previous error"
		})

		if status == entity.SyncStatusDownload {
			require.NoError(t, entity.NewFileSync(a.ID, "/changing.jpg").Create())
			t.Cleanup(func() { _ = entity.UnscopedDb().Delete(&entity.FileSync{}, "service_id = ?", a.ID).Error })
		}

		return a
	}

	t.Run("ErrorsReset", func(t *testing.T) {
		// The failed refresh is counted on the reset value.
		a := newChangingAccount(t, entity.SyncStatusRefresh, entity.Values{"acc_errors": 0, "acc_error": ""})
		require.NoError(t, worker.Start())
		stored := storedSyncAccount(t, a.ID)
		assert.Equal(t, 1, stored.AccErrors)
		assert.NotEqual(t, "previous error", stored.AccError)
		assert.NotEmpty(t, stored.AccError)
		assert.Equal(t, entity.SyncStatusRefresh, stored.SyncStatus)
	})
	t.Run("CountChanged", func(t *testing.T) {
		// Errors counted by another writer during the run, e.g. failed folder listings, are kept.
		a := newChangingAccount(t, entity.SyncStatusRefresh, entity.Values{"acc_errors": 5})
		require.NoError(t, worker.Start())
		assert.Equal(t, 6, storedSyncAccount(t, a.ID).AccErrors)
	})
	t.Run("StatusChanged", func(t *testing.T) {
		// The download stage completes without a retry limit, so the worker would move the account to synced.
		a := newChangingAccount(t, entity.SyncStatusDownload, entity.Values{"sync_status": entity.SyncStatusUpload})
		require.NoError(t, entity.Db().Model(&entity.Service{ID: a.ID}).UpdateColumn("retry_limit", -1).Error)
		logger, hook := test.NewNullLogger()
		logger.SetLevel(logrus.DebugLevel)
		prev := log
		log = logger
		t.Cleanup(func() { log = prev })
		require.NoError(t, worker.Start())
		stored := storedSyncAccount(t, a.ID)
		assert.Equal(t, entity.SyncStatusUpload, stored.SyncStatus)
		assert.Equal(t, -1, stored.RetryLimit)
		var kept int
		for _, e := range hook.AllEntries() {
			if strings.Contains(e.Message, "kept the status saved to") {
				kept++
			}
		}
		assert.Equal(t, 1, kept)
	})
	t.Run("CountKeptWhenStageCompletes", func(t *testing.T) {
		// A download stage that completes writes its status, but leaves the errors counted during the run.
		a := newChangingAccount(t, entity.SyncStatusDownload, entity.Values{"acc_errors": 9})
		require.NoError(t, entity.Db().Model(&entity.Service{ID: a.ID}).UpdateColumn("retry_limit", -1).Error)
		require.NoError(t, worker.Start())
		stored := storedSyncAccount(t, a.ID)
		assert.Equal(t, entity.SyncStatusSynced, stored.SyncStatus)
		assert.Equal(t, 9, stored.AccErrors)
		assert.Equal(t, "previous error", stored.AccError)
	})
	t.Run("OtherSettingsChanged", func(t *testing.T) {
		// A change of columns the worker does not write leaves its own values to be written.
		a := newChangingAccount(t, entity.SyncStatusRefresh, entity.Values{"retry_limit": 7, "sync_interval": 60})
		require.NoError(t, worker.Start())
		stored := storedSyncAccount(t, a.ID)
		assert.Equal(t, 3, stored.AccErrors)
		assert.NotEqual(t, "previous error", stored.AccError)
		assert.Equal(t, 7, stored.RetryLimit)
		assert.Equal(t, 60, stored.SyncInterval)
	})
	t.Run("Unchanged", func(t *testing.T) {
		a := newChangingAccount(t, entity.SyncStatusRefresh, nil)
		require.NoError(t, worker.Start())
		stored := storedSyncAccount(t, a.ID)
		assert.Equal(t, 3, stored.AccErrors)
		assert.NotEqual(t, "previous error", stored.AccError)
		assert.Equal(t, entity.SyncStatusRefresh, stored.SyncStatus)
	})
	t.Run("ErrorsChangeWhileProgressing", func(t *testing.T) {
		// Errors written by others during a stage that completes do not keep it from advancing.
		remote := t.TempDir()
		var a *entity.Service
		handler := &webdav.Handler{FileSystem: webdav.Dir(remote), LockSystem: webdav.NewMemLS()}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.NoError(t, entity.Db().Model(&entity.Service{ID: a.ID}).
				UpdateColumn("acc_errors", gorm.Expr("acc_errors + 1")).Error)
			handler.ServeHTTP(w, r)
		}))
		t.Cleanup(server.Close)
		a = newSyncAccount(t, func(m *entity.Service) {
			m.AccURL, m.AccTimeout, m.SyncStatus, m.SyncDownload = server.URL+"/", "low", entity.SyncStatusRefresh, true
		})
		require.NoError(t, worker.Start())
		stored := storedSyncAccount(t, a.ID)
		assert.Equal(t, entity.SyncStatusDownload, stored.SyncStatus)
		assert.Equal(t, 0, stored.AccErrors)
	})
	t.Run("Synced", func(t *testing.T) {
		// A download stage without queued files completes without a request and writes the sync date.
		a := newSyncAccount(t, func(m *entity.Service) {
			m.SyncStatus, m.SyncDownload, m.SyncDate = entity.SyncStatusDownload, true, sql.NullTime{}
			m.AccErrors, m.AccError = 1, "previous error"
		})
		sub := event.Subscribe("sync.synced")
		t.Cleanup(func() { event.Unsubscribe(sub) })
		require.NoError(t, worker.Start())
		stored := storedSyncAccount(t, a.ID)
		assert.Equal(t, entity.SyncStatusSynced, stored.SyncStatus)
		assert.True(t, stored.SyncDate.Valid)
		select {
		case msg := <-sub.Receiver:
			assert.Equal(t, "sync.synced", msg.Name)
		case <-time.After(2 * time.Second):
			t.Fatal("expected one sync.synced event")
		}
		assert.WithinDuration(t, time.Now(), stored.SyncDate.Time, time.Minute)
		assert.Equal(t, 1, stored.AccErrors)
	})
	t.Run("SyncedStatusKept", func(t *testing.T) {
		// A stage that completes while the status was changed is not reported as synced.
		a := newSyncAccount(t, func(m *entity.Service) {
			m.SyncStatus, m.SyncDownload, m.SyncDate = entity.SyncStatusDownload, true, sql.NullTime{}
		})
		require.NoError(t, entity.Db().Callback().Query().After("gorm:query").Register("test:change_sync_status", func(tx *gorm.DB) {
			if tx.Statement.Table == "files_sync" {
				_ = entity.UnscopedDb().Exec("UPDATE services SET sync_status = ? WHERE id = ?", entity.SyncStatusRefresh, a.ID).Error
			}
		}))
		t.Cleanup(func() { _ = entity.Db().Callback().Query().Remove("test:change_sync_status") })
		sub := event.Subscribe("sync.synced")
		t.Cleanup(func() { event.Unsubscribe(sub) })
		require.NoError(t, worker.Start())
		assert.Equal(t, entity.SyncStatusRefresh, storedSyncAccount(t, a.ID).SyncStatus)
		select {
		case msg := <-sub.Receiver:
			t.Fatalf("unexpected event %s", msg.Name)
		case <-time.After(200 * time.Millisecond):
		}
	})
	t.Run("NothingToWrite", func(t *testing.T) {
		// A synced account that is not due keeps its values, so nothing is written.
		a := newSyncAccount(t, nil)
		var updates int
		require.NoError(t, entity.Db().Callback().Update().Before("gorm:update").Register("test:count_service_updates", func(tx *gorm.DB) {
			if tx.Statement.Table == (entity.Service{}).TableName() {
				updates++
			}
		}))
		t.Cleanup(func() { _ = entity.Db().Callback().Update().Remove("test:count_service_updates") })
		require.NoError(t, worker.Start())
		assert.Zero(t, updates)
		assert.Equal(t, entity.SyncStatusSynced, storedSyncAccount(t, a.ID).SyncStatus)
	})
	t.Run("NullColumns", func(t *testing.T) {
		// Rows written by earlier releases may hold NULL in the columns the worker writes.
		a := newChangingAccount(t, entity.SyncStatusRefresh, nil)
		require.NoError(t, entity.Db().Model(&entity.Service{ID: a.ID}).UpdateColumns(map[string]any{
			"acc_errors": gorm.Expr("NULL"), "acc_error": gorm.Expr("NULL"), "sync_status": gorm.Expr("NULL")}).Error)
		require.NoError(t, worker.Start())
		stored := storedSyncAccount(t, a.ID)
		assert.Equal(t, entity.SyncStatusRefresh, stored.SyncStatus)

		require.NoError(t, worker.Start())
		stored = storedSyncAccount(t, a.ID)
		assert.Equal(t, 1, stored.AccErrors)
		assert.NotEmpty(t, stored.AccError)
	})
}

func TestSync_logAccountErr(t *testing.T) {
	worker := NewSync(config.TestConfig())
	a := entity.Service{AccName: "Account\nName"}
	t.Run("Error", func(t *testing.T) {
		logger, hook := test.NewNullLogger()
		prev := log
		log = logger
		t.Cleanup(func() { log = prev })
		worker.logAccountErr(errors.New("upload failed"), a)
		require.Len(t, hook.AllEntries(), 1)
		assert.Equal(t, logrus.ErrorLevel, hook.LastEntry().Level)
		assert.Equal(t, "sync: upload failed ("+clean.Log(a.AccName)+")", hook.LastEntry().Message)
		assert.NotContains(t, hook.LastEntry().Message, "\n")
	})
	t.Run("NoError", func(t *testing.T) {
		logger, hook := test.NewNullLogger()
		prev := log
		log = logger
		t.Cleanup(func() { log = prev })
		worker.logAccountErr(nil, a)
		assert.Empty(t, hook.AllEntries())
	})
}

func TestSync_logFileErr(t *testing.T) {
	worker := NewSync(config.TestConfig())
	a := entity.Service{AccName: "Account"}
	t.Run("Error", func(t *testing.T) {
		logger, hook := test.NewNullLogger()
		prev := log
		log = logger
		t.Cleanup(func() { log = prev })
		worker.logFileErr(errors.New("save failed"), "/photos/a\nb.jpg", a)
		require.Len(t, hook.AllEntries(), 1)
		assert.Equal(t, logrus.ErrorLevel, hook.LastEntry().Level)
		assert.Equal(t, "sync: save failed ("+clean.Log("/photos/a\nb.jpg")+" from "+clean.Log("Account")+")", hook.LastEntry().Message)
		assert.NotContains(t, hook.LastEntry().Message, "\n")
	})
	t.Run("NoError", func(t *testing.T) {
		logger, hook := test.NewNullLogger()
		prev := log
		log = logger
		t.Cleanup(func() { log = prev })
		worker.logFileErr(nil, "/photos/a.jpg", a)
		assert.Empty(t, hook.AllEntries())
	})
}

func TestSync_logErr(t *testing.T) {
	worker := NewSync(config.TestConfig())
	for _, c := range []struct {
		name  string
		log   func(err error)
		level logrus.Level
	}{
		{"Error", worker.logErr, logrus.ErrorLevel},
		{"Warning", worker.logWarn, logrus.WarnLevel},
	} {
		t.Run(c.name, func(t *testing.T) {
			logger, hook := test.NewNullLogger()
			prev := log
			log = logger
			t.Cleanup(func() { log = prev })
			c.log(nil)
			assert.Empty(t, hook.AllEntries())
			c.log(errors.New("failed\nsync: forged"))
			require.Len(t, hook.AllEntries(), 1)
			assert.Equal(t, c.level, hook.LastEntry().Level)
			assert.NotContains(t, hook.LastEntry().Message, "\n")
			assert.True(t, strings.HasPrefix(hook.LastEntry().Message, "sync: failed"), hook.LastEntry().Message)
		})
	}
}
