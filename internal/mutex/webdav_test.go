package mutex

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/webdav"
)

func TestWebDAV(t *testing.T) {
	lockSystem := WebDAV("test")
	assert.NotNil(t, lockSystem)
	// Locks are clamped to a finite lifetime by default.
	assert.IsType(t, &webdavLockSystem{}, lockSystem)
}

func TestClampLockDuration(t *testing.T) {
	original := WebDAVMaxLockLifetime
	defer func() { WebDAVMaxLockLifetime = original }()
	t.Run("InfiniteRequestClampedToCap", func(t *testing.T) {
		WebDAVMaxLockLifetime = time.Hour
		assert.Equal(t, time.Hour, ClampLockDuration(-1))
	})
	t.Run("OverLongRequestClampedToCap", func(t *testing.T) {
		WebDAVMaxLockLifetime = time.Hour
		assert.Equal(t, time.Hour, ClampLockDuration(24*time.Hour))
	})
	t.Run("ShortRequestKept", func(t *testing.T) {
		WebDAVMaxLockLifetime = time.Hour
		assert.Equal(t, 10*time.Minute, ClampLockDuration(10*time.Minute))
	})
	t.Run("ExactCapKept", func(t *testing.T) {
		WebDAVMaxLockLifetime = time.Hour
		assert.Equal(t, time.Hour, ClampLockDuration(time.Hour))
	})
	t.Run("NegativeCapDisablesClamp", func(t *testing.T) {
		WebDAVMaxLockLifetime = -1
		assert.Equal(t, time.Duration(-1), ClampLockDuration(-1))
		assert.Equal(t, 24*time.Hour, ClampLockDuration(24*time.Hour))
	})
}

func TestWebDAVLockSystem_Create(t *testing.T) {
	original := WebDAVMaxLockLifetime
	defer func() { WebDAVMaxLockLifetime = original }()
	WebDAVMaxLockLifetime = time.Hour
	now := time.Unix(1700000000, 0)
	t.Run("InfiniteLockExpires", func(t *testing.T) {
		ls := &webdavLockSystem{LockSystem: webdav.NewMemLS()}
		// Request an infinite lock; it must be clamped so it eventually expires.
		// A conflicting Create on the same root is the observable: it is refused
		// while the lock is held and succeeds once the clamped lifetime elapses.
		token, err := ls.Create(now, webdav.LockDetails{Root: "/infinite", Duration: -1, ZeroDepth: true})
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
		// Still held just before the cap.
		_, err = ls.Create(now.Add(59*time.Minute), webdav.LockDetails{Root: "/infinite", Duration: time.Minute, ZeroDepth: true})
		assert.ErrorIs(t, err, webdav.ErrLocked)
		// Expired after the cap, so the root can be locked again.
		_, err = ls.Create(now.Add(61*time.Minute), webdav.LockDetails{Root: "/infinite", Duration: time.Minute, ZeroDepth: true})
		assert.NoError(t, err)
	})
	t.Run("ShortLockKept", func(t *testing.T) {
		ls := &webdavLockSystem{LockSystem: webdav.NewMemLS()}
		token, err := ls.Create(now, webdav.LockDetails{Root: "/short", Duration: 5 * time.Minute, ZeroDepth: true})
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
		// Still held within the requested five minutes (under the cap, so kept).
		_, err = ls.Create(now.Add(1*time.Minute), webdav.LockDetails{Root: "/short", Duration: time.Minute, ZeroDepth: true})
		assert.ErrorIs(t, err, webdav.ErrLocked)
		// Expired after the requested five minutes.
		_, err = ls.Create(now.Add(6*time.Minute), webdav.LockDetails{Root: "/short", Duration: time.Minute, ZeroDepth: true})
		assert.NoError(t, err)
	})
}

func TestWebDAVLockSystem_Refresh(t *testing.T) {
	original := WebDAVMaxLockLifetime
	defer func() { WebDAVMaxLockLifetime = original }()
	WebDAVMaxLockLifetime = time.Hour
	now := time.Unix(1700000000, 0)
	ls := &webdavLockSystem{LockSystem: webdav.NewMemLS()}
	token, err := ls.Create(now, webdav.LockDetails{Root: "/refresh", Duration: 5 * time.Minute, ZeroDepth: true})
	assert.NoError(t, err)
	t.Run("InfiniteRefreshClampedToCap", func(t *testing.T) {
		details, refreshErr := ls.Refresh(now, token, -1)
		assert.NoError(t, refreshErr)
		assert.Equal(t, time.Hour, details.Duration)
	})
	t.Run("NoSuchLock", func(t *testing.T) {
		_, refreshErr := ls.Refresh(now, "unknown-token", time.Minute)
		assert.Error(t, refreshErr)
	})
}
