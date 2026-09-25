package jwt

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManager_nowUTC(t *testing.T) {
	m, err := NewManager(newTestConfig(t))
	require.NoError(t, err)

	fixed := time.Date(2026, 9, 11, 12, 0, 0, 0, time.UTC)
	m.SetNow(func() time.Time { return fixed })

	assert.Equal(t, fixed, m.nowUTC())
	t.Run("ConcurrentWithSetNow", func(t *testing.T) {
		// An active key is required, or NeedsRotation returns before it reads the clock.
		_, err := m.EnsureActiveKey()
		require.NoError(t, err)

		var wg sync.WaitGroup

		for range 8 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for range 50 {
					_ = m.nowUTC()
					_ = m.NeedsRotation(time.Hour)
					_, _ = m.RetireSuperseded()
				}
			}()
		}

		for range 4 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for range 50 {
					m.SetNow(func() time.Time { return fixed })
				}
			}()
		}

		wg.Wait()
	})
}
