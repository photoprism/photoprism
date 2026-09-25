package limiter

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"
)

func TestNewLimit(t *testing.T) {
	clientIp := "192.0.2.1"

	t.Run("BelowLimit", func(t *testing.T) {
		// 10 per minute.
		l := NewLimit(0.166, 10)
		for range 9 {
			assert.True(t, l.IP(clientIp).Allow())
		}
	})
	t.Run("AboveLimit", func(t *testing.T) {
		// 10 per minute.
		l := NewLimit(0.166, 10)
		for range 10 {
			assert.True(t, l.IP(clientIp).Allow())
		}
		assert.False(t, l.IP(clientIp).Allow())
	})
	t.Run("MultipleIPs", func(t *testing.T) {
		// 10 per minute.
		l := NewLimit(0.166, 10)
		for i := range 100 {
			assert.True(t, l.IP(fmt.Sprintf("192.0.2.%d", i)).Allow())
		}
	})
	t.Run("Reject", func(t *testing.T) {
		// 10 per minute.
		l := NewLimit(0.166, 10)

		// Request counter not increased.
		for range 20 {
			assert.False(t, l.Reject(clientIp))
		}

		// Request counter checked and increased.
		for range 10 {
			assert.True(t, l.Allow(clientIp))
		}

		// Limit exceeded.
		for range 10 {
			assert.True(t, l.Reject(clientIp))
			assert.False(t, l.Allow(clientIp))
		}
	})
	t.Run("Reserve", func(t *testing.T) {
		// 10 per minute.
		l := NewLimit(0.166, 10)

		// Request counter not increased.
		for range 20 {
			assert.False(t, l.Reject(clientIp))
		}

		// Request counter checked and increased.
		for range 10 {
			assert.False(t, l.Reject(clientIp))
			l.Reserve(clientIp)
		}

		// Limit exceeded.
		for range 10 {
			l.Reserve(clientIp)
			assert.True(t, l.Reject(clientIp))
		}
	})
	t.Run("Request", func(t *testing.T) {
		// 10 per minute.
		l := NewLimit(0.166, 10)

		// Request counter not increased.
		for range 20 {
			assert.False(t, l.Reject(clientIp))
		}

		// Request not exceeded and tokens returned by calling Success().
		for i := 1; i <= 20; i++ {
			reject := l.Reject(clientIp)
			r := l.Request(clientIp)
			allow := r.Allow()
			r.Success()
			t.Logf("(1.%d) Reject: %t, Allow: %t, Tokens: %d", i, reject, allow, r.Tokens)
			assert.False(t, reject)
			assert.True(t, allow)
			assert.False(t, r.Reject())
		}

		// Limit not exceeded, but tokens not returned.
		for i := 1; i <= 10; i++ {
			reject := l.Reject(clientIp)
			r := l.Request(clientIp)
			allow := r.Allow()
			t.Logf("(2.%d) Reject: %t, Allow: %t, Tokens: %d", i, reject, allow, r.Tokens)
			assert.False(t, reject)
			assert.True(t, allow)
			assert.False(t, r.Reject())
		}

		// Limit exceeded and tokens not returned.
		for i := 1; i <= 20; i++ {
			reject := l.Reject(clientIp)
			r := l.Request(clientIp)
			allow := r.Allow()
			t.Logf("(3.%d) Reject: %t, Allow: %t, Tokens: %d", i, reject, allow, r.Tokens)
			assert.True(t, reject)
			assert.False(t, allow)
			assert.True(t, r.Reject())
		}
	})
}

func TestLimitSweep(t *testing.T) {
	clientIp := "192.0.2.1"

	// sweepNow makes a sweep due and runs it at the given time, which stands in for elapsed time
	// the bucket would otherwise have to wait out.
	sweepNow := func(l *Limit, now time.Time) {
		l.mu.Lock()
		l.swept = now.Add(-2 * SweepInterval)
		l.sweep(now)
		l.mu.Unlock()
	}

	t.Run("RemovesAFullBucket", func(t *testing.T) {
		l := NewLimit(0.166, 10)
		require.NotNil(t, l.IP(clientIp))
		require.Len(t, l.limiters, 1)
		sweepNow(l, time.Now())
		assert.Empty(t, l.limiters)
	})
	t.Run("KeepsASpentBucket", func(t *testing.T) {
		l := NewLimit(0.166, 10)
		for range 10 {
			require.True(t, l.Allow(clientIp))
		}
		require.False(t, l.Allow(clientIp))
		sweepNow(l, time.Now())
		assert.Contains(t, l.limiters, clientIp)
		assert.False(t, l.Allow(clientIp))
	})
	t.Run("RemovesABucketOnceItHasRefilled", func(t *testing.T) {
		l := NewLimit(0.166, 10)
		for range 10 {
			require.True(t, l.Allow(clientIp))
		}
		sweepNow(l, time.Now().Add(2*time.Minute))
		assert.Empty(t, l.limiters)
	})
	t.Run("KeepsABucketInDebt", func(t *testing.T) {
		// A reservation takes a bucket below zero, and the token test sees that and keeps it.
		l := NewLimit(0.166, 10)
		for range 50 {
			l.Reserve(clientIp)
		}
		sweepNow(l, time.Now().Add(2*time.Minute))
		assert.Contains(t, l.limiters, clientIp)
	})
	t.Run("KeepsABucketThatNeverRefills", func(t *testing.T) {
		l := NewLimit(0, 10)
		require.True(t, l.Allow(clientIp))
		sweepNow(l, time.Now().Add(24*time.Hour))
		assert.Contains(t, l.limiters, clientIp)
	})
	t.Run("RemovesEverythingWithNoLimit", func(t *testing.T) {
		l := NewLimit(rate.Inf, 0)
		require.NotNil(t, l.IP(clientIp))
		sweepNow(l, time.Now())
		assert.Empty(t, l.limiters)
	})
	t.Run("RunsNoOftenerThanTheInterval", func(t *testing.T) {
		l := NewLimit(0.166, 10)
		require.NotNil(t, l.IP(clientIp))
		now := time.Now()
		l.mu.Lock()
		l.swept = now
		l.sweep(now)
		l.mu.Unlock()
		assert.Contains(t, l.limiters, clientIp)
	})
}

func TestLimitAddKeepsAnExistingBucket(t *testing.T) {
	// The read lock is released before add runs, so add may find the address already present; it
	// returns the bucket it finds rather than a new one.
	clientIp := "192.0.2.1"
	l := NewLimit(0.166, 10)
	first := l.IP(clientIp)

	for range 10 {
		require.True(t, first.Allow())
	}

	second := l.add(clientIp, time.Now())

	assert.Same(t, first, second)
	assert.False(t, second.Allow())
	assert.Len(t, l.limiters, 1)
}

func TestLimitConcurrentFirstRequests(t *testing.T) {
	// Contention over the real IP path, with a sweep due so one fires under it.
	// TestLimitAddKeepsAnExistingBucket owns the one-bucket-per-address invariant.
	const burst = 10

	l := NewLimit(0.166, burst)
	l.swept = time.Now().Add(-2 * SweepInterval)
	clientIp := "192.0.2.2"

	var allowed atomic.Int64
	var wg sync.WaitGroup

	for range 200 {
		wg.Add(1)

		go func() {
			defer wg.Done()

			if l.Allow(clientIp) {
				allowed.Add(1)
			}
		}()
	}

	wg.Wait()

	assert.Equal(t, int64(burst), allowed.Load())
	assert.Len(t, l.limiters, 1)
}
