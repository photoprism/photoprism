package mutex

import (
	"sync"
	"time"

	"golang.org/x/net/webdav"
)

// WebDAVMaxLockLifetime caps how long a WebDAV lock may be held, so a client
// cannot mint infinite or over-long locks that leak memory and lock-null files
// until restart. A negative value disables the cap (upstream infinite default).
var WebDAVMaxLockLifetime = time.Hour

var webdavMutex = sync.Mutex{}
var webdavLocks = make(map[string]webdav.LockSystem)

// WebDAV returns a webdav.LockSystem for the specified base path.
func WebDAV(dir string) webdav.LockSystem {
	webdavMutex.Lock()
	defer webdavMutex.Unlock()

	// Return existing LockSystem, or create a new one otherwise.
	if lock, ok := webdavLocks[dir]; ok {
		return lock
	}

	lock := &webdavLockSystem{LockSystem: webdav.NewMemLS()}
	webdavLocks[dir] = lock

	return lock
}

// ClampLockDuration limits a requested lock duration to WebDAVMaxLockLifetime,
// treating a negative request or cap as infinite (webdav's convention).
func ClampLockDuration(d time.Duration) time.Duration {
	if limit := WebDAVMaxLockLifetime; limit < 0 {
		return d
	} else if d < 0 || d > limit {
		return limit
	}

	return d
}

// webdavLockSystem wraps a webdav.LockSystem to clamp lock lifetimes; it is the
// authoritative cap, while the server also clamps the LOCK timeout header so the
// granted lifetime reported to the client stays honest.
type webdavLockSystem struct {
	webdav.LockSystem
}

// Create clamps the requested lock duration before creating the lock.
func (ls *webdavLockSystem) Create(now time.Time, details webdav.LockDetails) (string, error) {
	details.Duration = ClampLockDuration(details.Duration)
	return ls.LockSystem.Create(now, details)
}

// Refresh clamps the requested lock duration before refreshing the lock.
func (ls *webdavLockSystem) Refresh(now time.Time, token string, duration time.Duration) (webdav.LockDetails, error) {
	return ls.LockSystem.Refresh(now, token, ClampLockDuration(duration))
}
