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

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
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

	require.NoError(t, entity.Db().Model(&entity.Service{}).Where("acc_sync = 1").Pluck("id", &ids).Error)

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

	// The remote answers the listing with an error whose text holds a new line, a field separator and more.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("listing failed\nsync \u203a admin " + strings.Repeat("x", 800)))
	}))
	t.Cleanup(server.Close)

	a := newSyncAccount(t, func(a *entity.Service) {
		a.AccURL = server.URL + "/"
		a.AccTimeout = "low"
		a.SyncStatus = entity.SyncStatusRefresh
	})

	require.NoError(t, worker.Start())

	stored := storedSyncAccount(t, a.ID)
	assert.Equal(t, 1, stored.AccErrors)
	assert.NotEmpty(t, stored.AccError)
	assert.LessOrEqual(t, len(stored.AccError), txt.ClipError)
	assert.NotContains(t, stored.AccError, "\n")
	assert.NotContains(t, stored.AccError, "\u203a")
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
