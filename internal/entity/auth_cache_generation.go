package entity

import (
	"sync"
	"time"
)

// AuthCacheGeneration identifies the account-cache state before an authentication lookup.
type AuthCacheGeneration struct {
	version uint64
	at      time.Time
}

// sessionTombstoneTTL is how long a deleted session ID is remembered, much longer than a session lookup takes.
const sessionTombstoneTTL = time.Minute

// sessionTombstone records the cache version at which a session was deleted.
type sessionTombstone struct {
	version uint64
	at      time.Time
}

// authCacheState coordinates invalidation with conditional cache publication.
var authCacheState = struct {
	sync.Mutex
	version uint64
	flushed uint64
	users   map[string]uint64
	deleted map[string]sessionTombstone
	pruned  time.Time
}{users: make(map[string]uint64), deleted: make(map[string]sessionTombstone)}

// CurrentAuthCacheGeneration captures the generation before loading authentication data.
func CurrentAuthCacheGeneration() AuthCacheGeneration {
	authCacheState.Lock()
	defer authCacheState.Unlock()

	return AuthCacheGeneration{version: authCacheState.version, at: time.Now()}
}

// valid reports whether the account has remained unchanged while authCacheState is locked.
func (g AuthCacheGeneration) valid(userUID string) bool {
	return g.version >= authCacheState.flushed && g.version >= authCacheState.users[userUID]
}

// sessionDeleted reports whether the session was deleted after the generation was captured, while
// authCacheState is locked. A generation older than sessionTombstoneTTL counts as deleted, since the
// tombstone it would be checked against may already be pruned.
func (g AuthCacheGeneration) sessionDeleted(sessionID string) bool {
	if time.Since(g.at) > sessionTombstoneTTL {
		return true
	}

	t, ok := authCacheState.deleted[sessionID]
	return ok && g.version < t.version
}
