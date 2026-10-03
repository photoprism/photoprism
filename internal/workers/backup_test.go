package workers

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/pkg/log/status"
)

func TestBackup_Start(t *testing.T) {
	conf := config.TestConfig()

	t.Logf("database-dsn: %s", conf.DatabaseDSN())

	worker := NewBackup(conf)

	assert.IsType(t, &Backup{}, worker)

	if err := mutex.BackupWorker.Start(); err != nil {
		t.Fatal(err)
	}

	// Mutex should prevent worker from starting.
	if err := worker.Start(true, true, true, 2); err == nil {
		t.Fatal("error expected")
	}

	mutex.BackupWorker.Stop()

	day := time.Now().UTC().Format("2006-01-02")

	// Start worker.
	if err := worker.Start(true, true, true, 2); err != nil {
		t.Fatal(err)
	}

	// Rerun worker without force, which cannot replace the database backup of the same day.
	err := worker.Start(true, true, false, 2)

	if time.Now().UTC().Format("2006-01-02") != day {
		t.Skip("the date changed between the runs")
	}

	require.Error(t, err)
	assert.Equal(t, "database backup failed", err.Error())
}

// stubCall records the arguments a backup step was called with.
type stubCall struct {
	step     string
	path     string
	fileName string
	toStdOut bool
	force    bool
	retain   int
}

// newStubBackup returns a Backup worker whose steps return the given errors and record their calls.
func newStubBackup(t *testing.T, dbErr, albumsErr error) (w *Backup, calls *[]stubCall, hook *test.Hook) {
	t.Helper()

	logger, hook := test.NewNullLogger()
	prev := log
	log = logger
	t.Cleanup(func() { log = prev })

	calls = &[]stubCall{}
	w = NewBackup(config.TestConfig())
	w.database = func(backupPath, fileName string, toStdOut, force bool, retain int) error {
		*calls = append(*calls, stubCall{step: "database", path: backupPath, fileName: fileName, toStdOut: toStdOut, force: force, retain: retain})
		return dbErr
	}
	w.albums = func(backupPath string, force bool) (int, error) {
		*calls = append(*calls, stubCall{step: "album", path: backupPath, force: force})
		return 1, albumsErr
	}

	return w, calls, hook
}

// steps returns the names of the recorded steps.
func steps(calls *[]stubCall) []string {
	names := make([]string, 0, len(*calls))

	for _, c := range *calls {
		names = append(names, c.step)
	}

	return names
}

// logged reports whether an entry at the given level contains s.
func logged(hook *test.Hook, level logrus.Level, s string) bool {
	for _, e := range hook.AllEntries() {
		if e.Level == level && strings.Contains(e.Message, s) {
			return true
		}
	}

	return false
}

// errorMessages returns the messages of all entries logged at error level.
func errorMessages(hook *test.Hook) []string {
	var messages []string

	for _, e := range hook.AllEntries() {
		if e.Level == logrus.ErrorLevel {
			messages = append(messages, e.Message)
		}
	}

	return messages
}

func TestNewStubBackup(t *testing.T) {
	w, calls, hook := newStubBackup(t, nil, nil)

	require.NoError(t, w.database("db", "", false, true, 3))
	_, err := w.albums("albums", false)
	require.NoError(t, err)
	assert.Equal(t, []stubCall{{step: "database", path: "db", force: true, retain: 3}, {step: "album", path: "albums"}}, *calls)
	assert.Equal(t, []string{"database", "album"}, steps(calls))
	log.Info("stub log")
	assert.True(t, logged(hook, logrus.InfoLevel, "stub log"))
	assert.False(t, logged(hook, logrus.ErrorLevel, "stub log"))
	log.Error("stub error")
	assert.Equal(t, []string{"stub error"}, errorMessages(hook))
}

func TestBackup_StartSteps(t *testing.T) {
	dbErr, albumsErr := errors.New("database step error"), errors.New("album step error")

	t.Run("DatabaseFails", func(t *testing.T) {
		w, calls, hook := newStubBackup(t, dbErr, nil)

		err := w.Start(true, true, true, 2)

		require.Error(t, err)
		assert.Equal(t, "database backup failed", err.Error())
		assert.ErrorIs(t, err, dbErr)
		assert.Equal(t, []string{"database", "album"}, steps(calls))
		assert.Equal(t, []string{"backup: database step error (database)", "backup: database backup failed"}, errorMessages(hook))
		assert.False(t, logged(hook, logrus.InfoLevel, "completed"))
	})
	t.Run("AlbumsFail", func(t *testing.T) {
		w, calls, hook := newStubBackup(t, nil, albumsErr)

		err := w.Start(true, true, true, 2)

		require.Error(t, err)
		assert.Equal(t, "album backup failed", err.Error())
		assert.ErrorIs(t, err, albumsErr)
		assert.Equal(t, []string{"database", "album"}, steps(calls))
		assert.Equal(t, []string{"backup: album step error (albums)", "backup: album backup failed"}, errorMessages(hook))
		assert.False(t, logged(hook, logrus.InfoLevel, "completed"))
	})
	t.Run("BothFail", func(t *testing.T) {
		w, calls, hook := newStubBackup(t, dbErr, albumsErr)

		err := w.Start(true, true, true, 2)

		require.Error(t, err)
		assert.Equal(t, "database and album backup failed", err.Error())
		assert.ErrorIs(t, err, dbErr)
		assert.ErrorIs(t, err, albumsErr)
		assert.Equal(t, []string{"database", "album"}, steps(calls))
		assert.Equal(t, []string{
			"backup: database step error (database)",
			"backup: album step error (albums)",
			"backup: database and album backup failed",
		}, errorMessages(hook))
	})
	t.Run("DatabaseOnly", func(t *testing.T) {
		w, calls, _ := newStubBackup(t, dbErr, albumsErr)

		err := w.Start(true, false, true, 2)

		require.Error(t, err)
		assert.Equal(t, "database backup failed", err.Error())
		assert.Equal(t, []string{"database"}, steps(calls))
	})
	t.Run("Arguments", func(t *testing.T) {
		w, calls, _ := newStubBackup(t, nil, nil)
		conf := config.TestConfig()

		require.NoError(t, w.Start(true, true, true, 5))
		assert.Equal(t, []stubCall{
			{step: "database", path: conf.BackupDatabasePath(), force: true, retain: 5},
			{step: "album", path: conf.BackupAlbumsPath()},
		}, *calls)
	})
	t.Run("Canceled", func(t *testing.T) {
		w, calls, _ := newStubBackup(t, nil, nil)
		w.database = func(backupPath, fileName string, toStdOut, force bool, retain int) error {
			*calls = append(*calls, stubCall{step: "database"})
			mutex.BackupWorker.Cancel()
			return dbErr
		}

		err := w.Start(true, true, true, 2)

		assert.ErrorIs(t, err, status.ErrCanceled)
		assert.Equal(t, []string{"database"}, steps(calls))
		assert.False(t, mutex.BackupWorker.Running())
	})
	t.Run("Panic", func(t *testing.T) {
		w, _, _ := newStubBackup(t, nil, nil)
		w.database = func(backupPath, fileName string, toStdOut, force bool, retain int) error {
			panic("database step panic")
		}

		err := w.Start(true, true, true, 2)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "database step panic")
		assert.False(t, mutex.BackupWorker.Running())
	})
	t.Run("Success", func(t *testing.T) {
		w, calls, hook := newStubBackup(t, nil, nil)

		require.NoError(t, w.Start(true, true, true, 2))
		assert.Equal(t, []string{"database", "album"}, steps(calls))
		assert.True(t, logged(hook, logrus.InfoLevel, "backup: saved 1 album backup"))
		assert.True(t, logged(hook, logrus.InfoLevel, "backup: completed in"))
		assert.Empty(t, errorMessages(hook))
	})
	t.Run("Nothing", func(t *testing.T) {
		w, calls, _ := newStubBackup(t, dbErr, albumsErr)

		require.NoError(t, w.Start(false, false, true, 2))
		assert.Empty(t, *calls)
	})
}

func TestBackup_StartScheduled(t *testing.T) {
	opt := config.TestConfig().Options()
	prev := opt.BackupDatabase
	t.Cleanup(func() { opt.BackupDatabase = prev })

	t.Run("StepFailsLoggedOnce", func(t *testing.T) {
		w, _, hook := newStubBackup(t, errors.New("database step error"), nil)
		opt.BackupDatabase = true

		w.StartScheduled()

		assert.Equal(t, []string{"backup: database step error (database)", "backup: database backup failed"}, errorMessages(hook))
	})
	t.Run("PanicLogged", func(t *testing.T) {
		w, _, hook := newStubBackup(t, nil, nil)
		opt.BackupDatabase = true
		w.database = func(backupPath, fileName string, toStdOut, force bool, retain int) error {
			panic("database step panic")
		}

		w.StartScheduled()

		assert.True(t, logged(hook, logrus.ErrorLevel, "scheduler: backup: database step panic (worker panic)"))
	})
	t.Run("Busy", func(t *testing.T) {
		w, calls, hook := newStubBackup(t, nil, nil)
		opt.BackupDatabase = true
		require.NoError(t, mutex.BackupWorker.Start())
		t.Cleanup(mutex.BackupWorker.Stop)

		w.StartScheduled()

		assert.Empty(t, *calls)
		assert.Equal(t, []string{"scheduler: already running (backup)"}, errorMessages(hook))
	})
}

func TestBackupError(t *testing.T) {
	t.Run("OneStep", func(t *testing.T) {
		stepErr := errors.New("step error")
		e := &backupError{}
		e.add("database", stepErr)

		assert.Equal(t, "database backup failed", e.Error())
		assert.ErrorIs(t, e, stepErr)
		assert.NotContains(t, e.Error(), "step error")
	})
	t.Run("TwoSteps", func(t *testing.T) {
		e := &backupError{}
		e.add("database", errors.New("a"))
		e.add("album", errors.New("b"))

		assert.Equal(t, "database and album backup failed", e.Error())
		assert.Len(t, e.Unwrap(), 2)
	})
}
