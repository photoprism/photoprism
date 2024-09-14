package limiter

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLimit(t *testing.T) {
	clientIp := "192.0.2.1"

	t.Run("BelowLimit", func(t *testing.T) {
		// 10 per minute.
		l := NewLimit(0.166, 10)
		for i := 0; i < 9; i++ {
			assert.True(t, l.IP(clientIp).Allow())
		}
	})
	t.Run("AboveLimit", func(t *testing.T) {
		// 10 per minute.
		l := NewLimit(0.166, 10)
		for i := 0; i < 10; i++ {
			assert.True(t, l.IP(clientIp).Allow())
		}
		assert.False(t, l.IP(clientIp).Allow())
	})
	t.Run("MultipleIPs", func(t *testing.T) {
		// 10 per minute.
		l := NewLimit(0.166, 10)
		for i := 0; i < 100; i++ {
			assert.True(t, l.IP(fmt.Sprintf("192.0.2.%d", i)).Allow())
		}
	})
	t.Run("Reject", func(t *testing.T) {
		// 10 per minute.
		l := NewLimit(0.166, 10)

		// Request counter not increased.
		for i := 0; i < 20; i++ {
			assert.False(t, l.Reject(clientIp))
		}

		// Request counter checked and increased.
		for i := 0; i < 10; i++ {
			assert.True(t, l.Allow(clientIp))
		}

		// Limit exceeded.
		for i := 0; i < 10; i++ {
			assert.True(t, l.Reject(clientIp))
			assert.False(t, l.Allow(clientIp))
		}
	})
	t.Run("Reserve", func(t *testing.T) {
		// 10 per minute.
		l := NewLimit(0.166, 10)

		// Request counter not increased.
		for i := 0; i < 20; i++ {
			assert.False(t, l.Reject(clientIp))
		}

		// Request counter checked and increased.
		for i := 0; i < 10; i++ {
			assert.False(t, l.Reject(clientIp))
			l.Reserve(clientIp)
		}

		// Limit exceeded.
		for i := 0; i < 10; i++ {
			l.Reserve(clientIp)
			assert.True(t, l.Reject(clientIp))
		}
	})
	t.Run("Request", func(t *testing.T) {
		// 10 per minute.
		l := NewLimit(0.166, 10)

		// Request counter not increased.
		for i := 0; i < 20; i++ {
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
