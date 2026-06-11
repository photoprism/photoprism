package photoprism

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestLockFileHash verifies that lockFileHash serializes concurrent holders of
// the same hash and removes map entries once the last holder releases.
func TestLockFileHash(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		unlock := lockFileHash("abc123")
		assert.Len(t, fileHashLocks, 1)
		unlock()
		assert.Len(t, fileHashLocks, 0)
	})
	// Callers skip the lock for empty hashes; this is a defensive contract check.
	t.Run("EmptyHash", func(t *testing.T) {
		unlock := lockFileHash("")
		assert.Len(t, fileHashLocks, 1)
		unlock()
		assert.Len(t, fileHashLocks, 0)
	})
	t.Run("DistinctKeysDoNotBlock", func(t *testing.T) {
		unlockFirst := lockFileHash("hash-one")
		unlockSecond := lockFileHash("hash-two")
		assert.Len(t, fileHashLocks, 2)
		unlockSecond()
		unlockFirst()
		assert.Len(t, fileHashLocks, 0)
	})
	t.Run("Contention", func(t *testing.T) {
		const workers = 8
		var wg sync.WaitGroup
		var inCriticalSection int32
		var entered int32

		for i := 0; i < workers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				unlock := lockFileHash("contended")
				defer unlock()

				if !atomic.CompareAndSwapInt32(&inCriticalSection, 0, 1) {
					t.Error("two goroutines entered the critical section for the same hash")
				}

				atomic.AddInt32(&entered, 1)
				atomic.StoreInt32(&inCriticalSection, 0)
			}()
		}

		wg.Wait()

		assert.Equal(t, int32(workers), entered)
		assert.Len(t, fileHashLocks, 0)
	})
}
