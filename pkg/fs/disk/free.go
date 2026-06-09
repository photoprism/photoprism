package disk

import (
	"sync"
	"time"

	"github.com/photoprism/photoprism/pkg/fs/duf"
)

// DefaultCacheTTL specifies the cache TTL default.
const DefaultCacheTTL = 5 * time.Minute

// CacheTTL controls how long Free results are cached per path.
// Read on every lookup so callers may retune it at runtime.
var CacheTTL = DefaultCacheTTL

// freeEntry caches a single Free probe with its expiry time.
type freeEntry struct {
	free      uint64
	total     uint64
	expiresAt time.Time
}

var (
	freeMu    sync.RWMutex
	freeCache = map[string]freeEntry{}
)

// Free returns the cached free and total bytes for the filesystem that owns path.
func Free(path string) (free, total uint64, err error) {
	now := time.Now()

	freeMu.RLock()
	entry, ok := freeCache[path]
	freeMu.RUnlock()

	if ok && now.Before(entry.expiresAt) {
		return entry.free, entry.total, nil
	}

	mount, err := duf.PathInfo(path)
	if err != nil {
		return 0, 0, err
	}

	freeMu.Lock()
	freeCache[path] = freeEntry{
		free:      mount.Free,
		total:     mount.Total,
		expiresAt: now.Add(CacheTTL),
	}
	freeMu.Unlock()

	return mount.Free, mount.Total, nil
}

// FlushFree clears the cache so a freshly freed disk is detected immediately
// without waiting for CacheTTL to expire.
func FlushFree() {
	freeMu.Lock()
	freeCache = map[string]freeEntry{}
	freeMu.Unlock()
}

// SetFree stores a free/total entry for path in the cache, expiring after CacheTTL.
// Allows tests and external probes to inject known values without hitting the filesystem.
func SetFree(path string, free, total uint64) {
	freeMu.Lock()
	freeCache[path] = freeEntry{
		free:      free,
		total:     total,
		expiresAt: time.Now().Add(CacheTTL),
	}
	freeMu.Unlock()
}
