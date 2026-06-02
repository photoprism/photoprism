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
		CacheTTL = DefaultCacheTTL
		StorageLowPct = DefaultStorageLowPct
		StorageLowBytes = DefaultStorageLowBytes
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

func TestSetFree(t *testing.T) {
	reset(t)

	SetFree("/seeded", 42*MB, 100*MB)

	free, total, err := Free("/seeded")
	require.NoError(t, err)
	assert.Equal(t, 42*MB, free)
	assert.Equal(t, 100*MB, total)

	// Subsequent SetFree calls must overwrite the prior entry.
	SetFree("/seeded", 7*MB, 100*MB)

	free, _, err = Free("/seeded")
	require.NoError(t, err)
	assert.Equal(t, 7*MB, free)
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
