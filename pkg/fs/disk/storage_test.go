package disk

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStorageLow(t *testing.T) {
	t.Run("AboveThreshold", func(t *testing.T) {
		reset(t)
		StorageLowPct = 0.0001

		// A 0.0001% threshold on a real filesystem is effectively never low.
		free, low, err := StorageLow("/")
		require.NoError(t, err)
		assert.False(t, low)
		assert.NotZero(t, free)
	})
	t.Run("BelowThreshold", func(t *testing.T) {
		reset(t)
		StorageLowPct = 1.0

		// Seed the cache so the verdict does not depend on the host filesystem's actual fill level.
		freeMu.Lock()
		freeCache["/below"] = freeEntry{free: 5 * MB, total: 1000 * MB, expiresAt: time.Now().Add(time.Hour)}
		freeMu.Unlock()

		free, low, err := StorageLow("/below")
		require.NoError(t, err)
		assert.True(t, low)
		assert.Equal(t, 5*MB, free)
	})
	t.Run("ExactlyAtThreshold", func(t *testing.T) {
		reset(t)
		StorageLowPct = 1.0
		StorageLowBytes = 5 * MB

		// Seed the cache with a known free / total split so the threshold math is deterministic.
		freeMu.Lock()
		freeCache["/synthetic"] = freeEntry{free: 10 * MB, total: 1000 * MB, expiresAt: time.Now().Add(time.Hour)}
		freeMu.Unlock()
		// free*100 == StorageLowPct*total is the boundary and must not count as low (strict less-than).
		_, low, err := StorageLow("/synthetic")
		require.NoError(t, err)
		assert.False(t, low, "boundary value must not be flagged as low")

		// One byte less than the boundary must flip the verdict.
		freeMu.Lock()
		freeCache["/synthetic"] = freeEntry{free: 9 * MB, total: 1000 * MB, expiresAt: time.Now().Add(time.Hour)}
		freeMu.Unlock()

		_, low, err = StorageLow("/synthetic")
		require.NoError(t, err)
		assert.True(t, low)
	})
	t.Run("BytesFloorOnly", func(t *testing.T) {
		reset(t)
		StorageLowPct = 1.0
		StorageLowBytes = 100 * MB

		// On a small volume the 1% percentage threshold (10 MB of 1000 MB) sits below the
		// absolute floor, so free space above the percentage but below the floor is still low.
		freeMu.Lock()
		freeCache["/smalldisk"] = freeEntry{free: 50 * MB, total: 1000 * MB, expiresAt: time.Now().Add(time.Hour)}
		freeMu.Unlock()

		free, low, err := StorageLow("/smalldisk")
		require.NoError(t, err)
		assert.True(t, low, "free below the byte floor must be flagged even when the percentage is fine")
		assert.Equal(t, 50*MB, free)
	})
	t.Run("ZeroTotal", func(t *testing.T) {
		reset(t)
		StorageLowPct = 50

		// A zero-total filesystem must never be reported as low so unmounted or unreadable
		// paths do not trigger a spurious abort.
		freeMu.Lock()
		freeCache["/empty"] = freeEntry{free: 0, total: 0, expiresAt: time.Now().Add(time.Hour)}
		freeMu.Unlock()

		_, low, err := StorageLow("/empty")
		require.NoError(t, err)
		assert.False(t, low)
	})
	t.Run("ThresholdOutOfRange", func(t *testing.T) {
		reset(t)
		StorageLowBytes = 0

		freeMu.Lock()
		freeCache["/out"] = freeEntry{free: 1 * MB, total: 1000 * MB, expiresAt: time.Now().Add(time.Hour)}
		freeMu.Unlock()

		for _, pct := range []float64{-1, 0, 100, 250} {
			StorageLowPct = pct
			_, low, err := StorageLow("/out")
			require.NoError(t, err)
			assert.False(t, low, "out-of-range threshold %v must never be low", pct)
		}
	})
	t.Run("InvalidPath", func(t *testing.T) {
		reset(t)
		StorageLowPct = 1.0

		_, low, err := StorageLow("")
		assert.Error(t, err)
		assert.False(t, low)
	})
}
