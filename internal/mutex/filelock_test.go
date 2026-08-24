package mutex

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/fs"
)

// writeFileLock records an arbitrary state so a test can stand in for another process.
func writeFileLock(t *testing.T, fileName string, state FileLockState) {
	t.Helper()

	b, err := json.Marshal(state)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(fileName, b, fs.ModeFile))
}

func TestFileLockState_Expired(t *testing.T) {
	t.Run("Live", func(t *testing.T) {
		assert.False(t, FileLockState{ExpiresAt: time.Now().Add(time.Minute)}.Expired())
	})
	t.Run("Expired", func(t *testing.T) {
		assert.True(t, FileLockState{ExpiresAt: time.Now().Add(-time.Second)}.Expired())
	})
	t.Run("ZeroValue", func(t *testing.T) {
		// A file that parsed but carries no expiry must not read as a live lock, or a corrupt
		// write would block the instance until someone deleted it by hand.
		assert.True(t, FileLockState{}.Expired())
	})
}

func TestFileLockState_String(t *testing.T) {
	state := FileLockState{
		Action:    "faces migration",
		PID:       4242,
		Host:      "neon",
		UpdatedAt: time.Date(2026, 8, 24, 9, 44, 0, 0, time.UTC),
	}

	assert.Equal(t, "faces migration, running as process 4242 on neon since 2026-08-24T09:44:00Z", state.String())
}

func TestReadFileLock(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		fileName := filepath.Join(t.TempDir(), "faces.lock")
		writeFileLock(t, fileName, FileLockState{Action: "faces migration", PID: 1})

		assert.Equal(t, "faces migration", ReadFileLock(fileName).Action)
	})
	t.Run("Missing", func(t *testing.T) {
		assert.Empty(t, ReadFileLock(filepath.Join(t.TempDir(), "absent.lock")).Action)
	})
	t.Run("Malformed", func(t *testing.T) {
		// Treated as absent rather than as held: an unreadable lock nobody can interpret must
		// not be able to stop an instance permanently.
		fileName := filepath.Join(t.TempDir(), "faces.lock")
		require.NoError(t, os.WriteFile(fileName, []byte("not json"), fs.ModeFile))

		assert.Equal(t, FileLockState{}, ReadFileLock(fileName))
	})
}

func TestFileLockHeld(t *testing.T) {
	t.Run("Live", func(t *testing.T) {
		fileName := filepath.Join(t.TempDir(), "faces.lock")
		writeFileLock(t, fileName, FileLockState{Action: "faces migration", PID: 7, ExpiresAt: time.Now().Add(time.Minute)})

		assert.Contains(t, FileLockHeld(fileName), "faces migration")
	})
	t.Run("Expired", func(t *testing.T) {
		// The expiry is the whole point: a run that was killed leaves its file behind, and the
		// instance has to recover from that without anyone noticing the file exists.
		fileName := filepath.Join(t.TempDir(), "faces.lock")
		writeFileLock(t, fileName, FileLockState{Action: "faces migration", PID: 7, ExpiresAt: time.Now().Add(-time.Second)})

		assert.Empty(t, FileLockHeld(fileName))
	})
	t.Run("Missing", func(t *testing.T) {
		assert.Empty(t, FileLockHeld(filepath.Join(t.TempDir(), "absent.lock")))
	})
	t.Run("NoFilename", func(t *testing.T) {
		assert.Empty(t, FileLockHeld(""))
	})
}

func TestAcquireFileLock(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		fileName := filepath.Join(t.TempDir(), "sub", "faces.lock")

		lock, err := AcquireFileLock(fileName, "faces migration")
		require.NoError(t, err)
		require.NotNil(t, lock)
		t.Cleanup(lock.Release)

		state := ReadFileLock(fileName)
		assert.Equal(t, "faces migration", state.Action)
		assert.Equal(t, os.Getpid(), state.PID)
		assert.False(t, state.Expired())
		assert.Contains(t, FileLockHeld(fileName), "faces migration")
	})
	t.Run("AlreadyHeld", func(t *testing.T) {
		fileName := filepath.Join(t.TempDir(), "faces.lock")
		writeFileLock(t, fileName, FileLockState{Action: "faces migration", PID: 9, Host: "neon", ExpiresAt: time.Now().Add(time.Minute)})

		lock, err := AcquireFileLock(fileName, "faces migration")

		require.Error(t, err)
		assert.Nil(t, lock)
		assert.Contains(t, err.Error(), "process 9")
	})
	t.Run("TakesOverAnExpiredLock", func(t *testing.T) {
		fileName := filepath.Join(t.TempDir(), "faces.lock")
		writeFileLock(t, fileName, FileLockState{Action: "faces migration", PID: 9, ExpiresAt: time.Now().Add(-time.Minute)})

		lock, err := AcquireFileLock(fileName, "faces migration")
		require.NoError(t, err)
		t.Cleanup(lock.Release)

		assert.Equal(t, os.Getpid(), ReadFileLock(fileName).PID)
	})
	t.Run("NoFilename", func(t *testing.T) {
		lock, err := AcquireFileLock("", "faces migration")

		require.Error(t, err)
		assert.Nil(t, lock)
	})
}

func TestFileLock_Release(t *testing.T) {
	t.Run("RemovesTheFile", func(t *testing.T) {
		fileName := filepath.Join(t.TempDir(), "faces.lock")

		lock, err := AcquireFileLock(fileName, "faces migration")
		require.NoError(t, err)
		require.FileExists(t, fileName)

		lock.Release()

		assert.NoFileExists(t, fileName)
		assert.Empty(t, FileLockHeld(fileName))
	})
	t.Run("Repeated", func(t *testing.T) {
		// Callers defer it and may also release early, so the second call must not panic on a
		// channel that is already closed.
		fileName := filepath.Join(t.TempDir(), "faces.lock")

		lock, err := AcquireFileLock(fileName, "faces migration")
		require.NoError(t, err)

		lock.Release()
		assert.NotPanics(t, lock.Release)
	})
	t.Run("NilLock", func(t *testing.T) {
		assert.NotPanics(t, (*FileLock)(nil).Release)
	})
}

// TestFileLock_write checks that renewing moves the expiry forward, which is what keeps a run
// longer than FileLockMaxAge from releasing its own lock mid-way.
func TestFileLock_write(t *testing.T) {
	fileName := filepath.Join(t.TempDir(), "faces.lock")

	lock, err := AcquireFileLock(fileName, "faces migration")
	require.NoError(t, err)
	t.Cleanup(lock.Release)

	first := ReadFileLock(fileName).ExpiresAt
	time.Sleep(2 * time.Millisecond)
	require.NoError(t, lock.write())

	assert.True(t, ReadFileLock(fileName).ExpiresAt.After(first))
}
