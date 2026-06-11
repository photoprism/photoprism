package photoprism

import "sync"

// fileHashLock serializes indexing of files that share the same content hash.
type fileHashLock struct {
	sync.Mutex
	refs int
}

// fileHashLocksMutex guards fileHashLocks.
var fileHashLocksMutex sync.Mutex

// fileHashLocks tracks the per-hash locks currently held or awaited by indexing workers.
var fileHashLocks = make(map[string]*fileHashLock)

// lockFileHash acquires the indexing lock for the given file hash and returns the function that releases it.
// Entries are reference counted and removed on release, so the map only holds in-flight hashes.
func lockFileHash(fileHash string) (unlock func()) {
	fileHashLocksMutex.Lock()
	l, found := fileHashLocks[fileHash]
	if !found {
		l = &fileHashLock{}
		fileHashLocks[fileHash] = l
	}
	l.refs++
	fileHashLocksMutex.Unlock()

	l.Lock()

	return func() {
		l.Unlock()
		fileHashLocksMutex.Lock()
		l.refs--
		if l.refs == 0 {
			delete(fileHashLocks, fileHash)
		}
		fileHashLocksMutex.Unlock()
	}
}
