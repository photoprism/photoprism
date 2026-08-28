package mutex

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
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
		// Nothing interpretable, so no holder is named - but the file's own age still counts, or
		// a lock caught between being created and being written would read as free. It still
		// cannot stop an instance permanently, which is what the second half pins.
		fileName := filepath.Join(t.TempDir(), "faces.lock")
		require.NoError(t, os.WriteFile(fileName, []byte("not json"), fs.ModeFile))

		state := ReadFileLock(fileName)

		assert.Empty(t, state.Action)
		assert.Zero(t, state.PID)
		assert.False(t, state.Expired(), "a lock file written moments ago is not free")

		stale := time.Now().Add(-2 * FileLockMaxAge)
		require.NoError(t, os.Chtimes(fileName, stale, stale))

		assert.True(t, ReadFileLock(fileName).Expired(), "and one nobody renewed cannot block forever")
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

// TestAcquireFileLockIsExclusive pins that two processes reaching the lock at the same moment
// cannot both come away holding it. A read followed by a write let every racer through, and
// preventing two concurrent finalize transactions is the whole reason this lock exists.
func TestAcquireFileLockIsExclusive(t *testing.T) {
	fileName := filepath.Join(t.TempDir(), "faces.lock")

	const racers = 16

	var wg sync.WaitGroup
	var held atomic.Int32

	start := make(chan struct{})
	locks := make(chan *FileLock, racers)

	for range racers {
		wg.Add(1)

		go func() {
			defer wg.Done()
			<-start

			if lock, err := AcquireFileLock(fileName, "faces migration"); err == nil {
				held.Add(1)
				locks <- lock
			}
		}()
	}

	close(start)
	wg.Wait()
	close(locks)

	for lock := range locks {
		lock.Release()
	}

	assert.Equal(t, int32(1), held.Load(), "exactly one racer may hold the lock")
}

// TestFileLockNotYetWrittenIsHeld pins the window between creating the lock file and filling it.
// A holder writes its state after the exclusive create, so a reader arriving in between sees a
// file with no recorded expiry - and reading that as free lets a second caller take the lock the
// exclusive create had just granted.
func TestFileLockNotYetWrittenIsHeld(t *testing.T) {
	fileName := filepath.Join(t.TempDir(), "faces.lock")

	require.NoError(t, os.WriteFile(fileName, nil, 0o644))

	assert.NotEmpty(t, FileLockHeld(fileName), "an empty lock file is held, not free")
	assert.False(t, ReadFileLock(fileName).Expired())

	// Old enough to be abandoned rather than mid-write, so it can still be taken over.
	stale := time.Now().Add(-2 * FileLockMaxAge)
	require.NoError(t, os.Chtimes(fileName, stale, stale))

	assert.Empty(t, FileLockHeld(fileName), "an abandoned empty lock file does not hold it forever")

	lock, err := AcquireFileLock(fileName, "faces migration")
	require.NoError(t, err)
	t.Cleanup(lock.Release)
}

// TestFileLockRenewIsAtomic pins that a reader never sees the lock as free while it is being
// renewed. Writing in place leaves the file empty between truncating and writing, and an
// unreadable lock reads as free - a window every worker polls into, once a minute for hours.
func TestFileLockRenewIsAtomic(t *testing.T) {
	fileName := filepath.Join(t.TempDir(), "faces.lock")

	lock, err := AcquireFileLock(fileName, "faces migration")
	require.NoError(t, err)
	t.Cleanup(lock.Release)

	done := make(chan struct{})
	var free atomic.Bool

	go func() {
		defer close(done)

		for range 2000 {
			if FileLockHeld(fileName) == "" {
				free.Store(true)
				return
			}
		}
	}()

	for range 200 {
		require.NoError(t, lock.write())
	}

	<-done

	assert.False(t, free.Load(), "the lock must never read as free while it is held")
}

// TestFileLock_ReleaseKeepsAForeignLock pins that a holder whose own lock lapsed and was taken
// over cannot free the one that replaced it.
func TestFileLock_ReleaseKeepsAForeignLock(t *testing.T) {
	fileName := filepath.Join(t.TempDir(), "faces.lock")

	lock, err := AcquireFileLock(fileName, "faces migration")
	require.NoError(t, err)

	// Another process took over after this holder's renewals stopped.
	writeFileLock(t, fileName, FileLockState{
		Action:    "faces migration",
		PID:       os.Getpid() + 1,
		Host:      "elsewhere",
		UpdatedAt: time.Now(),
		ExpiresAt: time.Now().Add(FileLockMaxAge),
	})

	lock.Release()

	assert.FileExists(t, fileName)
	assert.Contains(t, FileLockHeld(fileName), "elsewhere")
}

// TestFileLockStateExpiredClampsAFutureExpiry pins that a holder whose clock ran fast cannot wedge
// every worker for as long as that clock was wrong.
//
// Both timestamps inside the file come from that clock, so neither bounds the other; the file's
// modification time does, because the kernel that stored it wrote that.
func TestFileLockStateExpiredClampsAFutureExpiry(t *testing.T) {
	now := time.Now()

	t.Run("FutureExpiryOnAFreshFile", func(t *testing.T) {
		// Written now by a clock 48 hours ahead: held, but for one interval rather than two days.
		fileName := filepath.Join(t.TempDir(), "faces.lock")
		writeFileLock(t, fileName, FileLockState{
			Action:    "faces migration",
			PID:       7,
			UpdatedAt: now.Add(48 * time.Hour),
			ExpiresAt: now.Add(48*time.Hour + FileLockMaxAge),
		})

		assert.False(t, ReadFileLock(fileName).Expired())

		// The same file one interval later is stale, whatever it claims.
		stale := now.Add(-2 * FileLockMaxAge)
		require.NoError(t, os.Chtimes(fileName, stale, stale))

		assert.True(t, ReadFileLock(fileName).Expired(),
			"an expiry beyond one interval past the file's own modification time is stale")
	})
	t.Run("NoModTime", func(t *testing.T) {
		// A state that did not come from a file has only its own word to go on.
		assert.True(t, FileLockState{ExpiresAt: now.Add(-time.Second)}.Expired())
		assert.False(t, FileLockState{ExpiresAt: now.Add(FileLockMaxAge)}.Expired())
	})
	t.Run("Live", func(t *testing.T) {
		fileName := filepath.Join(t.TempDir(), "faces.lock")
		writeFileLock(t, fileName, FileLockState{PID: 7, UpdatedAt: now, ExpiresAt: now.Add(FileLockMaxAge)})

		assert.False(t, ReadFileLock(fileName).Expired())
	})
}

// TestFileLockRenewInterval pins the relation the renew loop depends on. Inverting it would make
// every run longer than one interval release its own lock mid-way, silently.
func TestFileLockRenewInterval(t *testing.T) {
	assert.Positive(t, fileLockRenewInterval)
	assert.Less(t, fileLockRenewInterval, FileLockMaxAge, "a lock must be renewed before it expires")
}

// TestFileLock_renew checks that the loop keeps the lock alive and stops when it is released.
func TestFileLock_renew(t *testing.T) {
	fileName := filepath.Join(t.TempDir(), "faces.lock")

	lock, err := AcquireFileLock(fileName, "faces migration")
	require.NoError(t, err)

	first := ReadFileLock(fileName).ExpiresAt
	time.Sleep(2 * time.Millisecond)
	require.NoError(t, lock.write())

	assert.True(t, ReadFileLock(fileName).ExpiresAt.After(first))

	lock.Release()

	// The goroutine selects on the same channel Release closes, so it is gone with the file.
	assert.NoFileExists(t, fileName)
	assert.NotPanics(t, lock.Release)
}

// TestAcquireFileLockTakesOverAnExpiredLockOnce pins that only one of two processes finding the
// same expired lock comes away holding it. Both may rename over it; the one whose write did not
// survive has to find that out by reading it back.
func TestAcquireFileLockTakesOverAnExpiredLockOnce(t *testing.T) {
	fileName := filepath.Join(t.TempDir(), "faces.lock")
	writeFileLock(t, fileName, FileLockState{Action: "faces migration", PID: 9, ExpiresAt: time.Now().Add(-time.Hour)})

	lock, err := AcquireFileLock(fileName, "faces migration")
	require.NoError(t, err)
	t.Cleanup(lock.Release)

	assert.Equal(t, os.Getpid(), ReadFileLock(fileName).PID)

	// A second process on the same host would see the live lock this one now holds.
	_, err = AcquireFileLock(fileName, "faces migration")
	assert.Error(t, err)
}

func TestFileLockState_HeldBy(t *testing.T) {
	state := FileLockState{PID: 42, Host: "neon"}

	assert.True(t, state.HeldBy(42, "neon"))
	assert.False(t, state.HeldBy(43, "neon"))
	assert.False(t, state.HeldBy(42, "other"))
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
