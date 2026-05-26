package disk

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// reset clears state mutated by individual tests so they cannot leak into one another.
func reset(t *testing.T) {
	t.Helper()
	FlushFree()
	t.Cleanup(func() {
		CacheTTL = time.Minute
		FlushFree()
	})
}

func TestFree(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		reset(t)

		free, total, err := Free("/")
		require.NoError(t, err)
		assert.NotZero(t, total)
		assert.LessOrEqual(t, free, total)
	})

	t.Run("Cached", func(t *testing.T) {
		reset(t)

		free1, total1, err := Free("/")
		require.NoError(t, err)

		free2, total2, err := Free("/")
		require.NoError(t, err)

		// Repeated calls within the TTL must return the cached values
		// even if the underlying filesystem has changed in the meantime.
		assert.Equal(t, free1, free2)
		assert.Equal(t, total1, total2)

		freeMu.RLock()
		_, hit := freeCache["/"]
		freeMu.RUnlock()
		assert.True(t, hit, "expected cache entry for /")
	})

	t.Run("Expires", func(t *testing.T) {
		reset(t)
		CacheTTL = time.Nanosecond

		_, _, err := Free("/")
		require.NoError(t, err)

		// Sleep long enough for the TTL to have elapsed.
		time.Sleep(10 * time.Millisecond)

		freeMu.RLock()
		stale := freeCache["/"]
		freeMu.RUnlock()
		assert.True(t, time.Now().After(stale.expiresAt), "cached entry should be expired")

		// Restore a meaningful TTL so the refresh produces a future expiry we can assert on.
		CacheTTL = time.Hour

		_, _, err = Free("/")
		require.NoError(t, err)

		freeMu.RLock()
		refreshed := freeCache["/"]
		freeMu.RUnlock()
		assert.True(t, refreshed.expiresAt.After(stale.expiresAt), "cached entry should have been refreshed")
		assert.True(t, time.Now().Before(refreshed.expiresAt), "refreshed entry must expire in the future")
	})

	t.Run("InvalidPath", func(t *testing.T) {
		reset(t)

		_, _, err := Free("")
		assert.Error(t, err)

		freeMu.RLock()
		_, hit := freeCache[""]
		freeMu.RUnlock()
		assert.False(t, hit, "errored probe must not populate the cache")
	})
}

func TestFlushFree(t *testing.T) {
	reset(t)

	_, _, err := Free("/")
	require.NoError(t, err)

	freeMu.RLock()
	assert.NotEmpty(t, freeCache)
	freeMu.RUnlock()

	FlushFree()

	freeMu.RLock()
	assert.Empty(t, freeCache)
	freeMu.RUnlock()
}

func TestStorageLow(t *testing.T) {
	t.Run("AboveThreshold", func(t *testing.T) {
		reset(t)

		// A 0.0001% threshold on a real filesystem is effectively never low.
		free, low, err := StorageLow("/", 0.0001)
		require.NoError(t, err)
		assert.False(t, low)
		assert.NotZero(t, free)
	})

	t.Run("BelowThreshold", func(t *testing.T) {
		reset(t)

		// Seed the cache so the verdict does not depend on the host filesystem's actual fill level.
		freeMu.Lock()
		freeCache["/below"] = freeEntry{free: 5, total: 1000, expiresAt: time.Now().Add(time.Hour)}
		freeMu.Unlock()

		free, low, err := StorageLow("/below", 1.0)
		require.NoError(t, err)
		assert.True(t, low)
		assert.Equal(t, uint64(5), free)
	})

	t.Run("ExactlyAtThreshold", func(t *testing.T) {
		reset(t)

		// Seed the cache with a known free / total split so the threshold math is deterministic.
		freeMu.Lock()
		freeCache["/synthetic"] = freeEntry{free: 10, total: 1000, expiresAt: time.Now().Add(time.Hour)}
		freeMu.Unlock()

		// free*100 == minPct*total is the boundary and must not count as low (strict less-than).
		_, low, err := StorageLow("/synthetic", 1.0)
		require.NoError(t, err)
		assert.False(t, low, "boundary value must not be flagged as low")

		// One byte less than the boundary must flip the verdict.
		freeMu.Lock()
		freeCache["/synthetic"] = freeEntry{free: 9, total: 1000, expiresAt: time.Now().Add(time.Hour)}
		freeMu.Unlock()

		_, low, err = StorageLow("/synthetic", 1.0)
		require.NoError(t, err)
		assert.True(t, low)
	})

	t.Run("ZeroTotal", func(t *testing.T) {
		reset(t)

		// A zero-total filesystem must never be reported as low so unmounted or unreadable
		// paths do not trigger a spurious abort.
		freeMu.Lock()
		freeCache["/empty"] = freeEntry{free: 0, total: 0, expiresAt: time.Now().Add(time.Hour)}
		freeMu.Unlock()

		_, low, err := StorageLow("/empty", 50)
		require.NoError(t, err)
		assert.False(t, low)
	})

	t.Run("ThresholdOutOfRange", func(t *testing.T) {
		reset(t)

		freeMu.Lock()
		freeCache["/out"] = freeEntry{free: 1, total: 1000, expiresAt: time.Now().Add(time.Hour)}
		freeMu.Unlock()

		for _, pct := range []float64{-1, 0, 100, 250} {
			_, low, err := StorageLow("/out", pct)
			require.NoError(t, err)
			assert.False(t, low, "out-of-range threshold %v must never be low", pct)
		}
	})

	t.Run("InvalidPath", func(t *testing.T) {
		reset(t)

		_, low, err := StorageLow("", 1)
		assert.Error(t, err)
		assert.False(t, low)
	})
}

func TestSetFree(t *testing.T) {
	reset(t)

	SetFree("/seeded", 42, 100)

	free, total, err := Free("/seeded")
	require.NoError(t, err)
	assert.Equal(t, uint64(42), free)
	assert.Equal(t, uint64(100), total)

	// Subsequent SetFree calls must overwrite the prior entry.
	SetFree("/seeded", 7, 100)

	free, _, err = Free("/seeded")
	require.NoError(t, err)
	assert.Equal(t, uint64(7), free)
}

func TestFree_Concurrent(t *testing.T) {
	reset(t)

	// Hammer the cache from many goroutines to confirm the RWMutex layout is correct.
	var wg sync.WaitGroup
	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _, err := Free("/")
			assert.NoError(t, err)
		}()
		wg.Add(1)
		go func() {
			defer wg.Done()
			FlushFree()
		}()
	}
	wg.Wait()
}
