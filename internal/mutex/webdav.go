package mutex

import (
	"sync"

	"golang.org/x/net/webdav"
)

var webdavMutex = sync.Mutex{}
var webdavLocks = make(map[string]webdav.LockSystem)

// WebDAV returns a webdav.LockSystem for the specified base path.
func WebDAV(dir string) webdav.LockSystem {
	webdavMutex.Lock()
	defer webdavMutex.Unlock()

	// Return existing LockSystem, or create a new one otherwise.
	if lock, ok := webdavLocks[dir]; ok {
		return lock
	} else if lock = webdav.NewMemLS(); lock != nil {
		webdavLocks[dir] = lock
		return lock
	}

	// Should never happen.
	return nil
}
