package server

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/pkg/http/header"
)

func TestWebDAVClampLockTimeout(t *testing.T) {
	original := mutex.WebDAVMaxLockLifetime
	defer func() { mutex.WebDAVMaxLockLifetime = original }()
	mutex.WebDAVMaxLockLifetime = time.Hour

	lockRequest := func(timeout string) *http.Request {
		r, _ := http.NewRequest(header.MethodLock, "http://localhost/originals/x.txt", nil)
		if timeout != "" {
			r.Header.Set(header.Timeout, timeout)
		}
		return r
	}

	t.Run("InfiniteClampedToCap", func(t *testing.T) {
		r := lockRequest("Infinite")
		WebDAVClampLockTimeout(r)
		assert.Equal(t, "Second-3600", r.Header.Get(header.Timeout))
	})
	t.Run("AbsentClampedToCap", func(t *testing.T) {
		r := lockRequest("")
		WebDAVClampLockTimeout(r)
		assert.Equal(t, "Second-3600", r.Header.Get(header.Timeout))
	})
	t.Run("OverCapClamped", func(t *testing.T) {
		r := lockRequest("Second-99999")
		WebDAVClampLockTimeout(r)
		assert.Equal(t, "Second-3600", r.Header.Get(header.Timeout))
	})
	t.Run("UnderCapKept", func(t *testing.T) {
		r := lockRequest("Second-600")
		WebDAVClampLockTimeout(r)
		assert.Equal(t, "Second-600", r.Header.Get(header.Timeout))
	})
	t.Run("FirstValueClamped", func(t *testing.T) {
		r := lockRequest("Infinite, Second-300")
		WebDAVClampLockTimeout(r)
		assert.Equal(t, "Second-3600", r.Header.Get(header.Timeout))
	})
	t.Run("MalformedLeftUntouched", func(t *testing.T) {
		r := lockRequest("bogus")
		WebDAVClampLockTimeout(r)
		assert.Equal(t, "bogus", r.Header.Get(header.Timeout))
	})
	t.Run("NonLockMethodIgnored", func(t *testing.T) {
		r, _ := http.NewRequest(header.MethodPut, "http://localhost/originals/x.txt", nil)
		r.Header.Set(header.Timeout, "Infinite")
		WebDAVClampLockTimeout(r)
		assert.Equal(t, "Infinite", r.Header.Get(header.Timeout))
	})
	t.Run("NegativeCapDisablesClamp", func(t *testing.T) {
		mutex.WebDAVMaxLockLifetime = -1
		r := lockRequest("Infinite")
		WebDAVClampLockTimeout(r)
		assert.Equal(t, "Infinite", r.Header.Get(header.Timeout))
	})
}
