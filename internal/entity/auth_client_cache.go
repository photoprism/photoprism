package entity

import (
	"errors"
	"slices"
	"sync"
	"time"

	gc "github.com/patrickmn/go-cache"
	"gorm.io/gorm"

	"github.com/photoprism/photoprism/internal/config/ttl"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// ClientCacheDuration specifies how long client lookups are cached.
var ClientCacheDuration = ttl.CacheClient.Duration()
var clientCache = gc.New(ClientCacheDuration, time.Minute)

// clientCacheState coordinates client cache flushes with conditional cache publication.
var clientCacheState = struct {
	sync.Mutex
	version uint64
	flushed uint64
	clients map[string]uint64
}{clients: make(map[string]uint64)}

// CachedClientByUID returns a copy of the client with the specified UID, including a deleted one, from
// cache or database. It returns ErrClientNotFound if there is no such client; misses and errors are not cached.
func CachedClientByUID(uid string) (Client, error) {
	if uid == "" || rnd.InvalidUID(uid, ClientUID) {
		return Client{}, ErrClientNotFound
	}

	if cached, hit := clientCache.Get(uid); hit {
		return cached.(Client).cacheCopy(), nil
	}

	version := clientCacheVersion()
	found := Client{}

	if err := UnscopedDb().First(&found, "client_uid = ?", uid).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return Client{}, ErrClientNotFound
	} else if err != nil {
		return Client{}, err
	}

	cacheClient(found.cacheCopy(), version)

	return found, nil
}

// FlushClientCache removes all cached clients.
func FlushClientCache() {
	clientCacheState.Lock()
	defer clientCacheState.Unlock()

	clientCacheState.version++
	clientCacheState.flushed = clientCacheState.version
	clear(clientCacheState.clients)
	clientCache.Flush()
}

// flushClientCache removes the cached client with the specified UID; callers flush after the write.
func flushClientCache(uid string) {
	if uid == "" {
		return
	}

	clientCacheState.Lock()
	defer clientCacheState.Unlock()

	clientCacheState.version++
	clientCacheState.clients[uid] = clientCacheState.version
	clientCache.Delete(uid)
}

// clientCacheVersion returns the cache version to capture before a client is read from the database.
func clientCacheVersion() uint64 {
	clientCacheState.Lock()
	defer clientCacheState.Unlock()

	return clientCacheState.version
}

// cacheClient stores the client unless it was flushed after the specified version was captured.
func cacheClient(m Client, version uint64) {
	clientCacheState.Lock()
	defer clientCacheState.Unlock()

	if version < clientCacheState.flushed || version < clientCacheState.clients[m.ClientUID] {
		return
	}

	clientCache.Set(m.ClientUID, m, ClientCacheDuration)
}

// cacheCopy returns a copy that shares no references with the receiver and holds no account, parsed data or
// matched secret hash.
func (m Client) cacheCopy() Client {
	m.user = nil
	m.data = nil
	m.secretHash = ""
	m.DataJSON = slices.Clone(m.DataJSON)

	return m
}
