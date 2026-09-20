package entity

import "sync"

// AuthCacheGeneration identifies the account-cache state before an authentication lookup.
type AuthCacheGeneration struct {
	version uint64
}

// authCacheState coordinates invalidation with conditional cache publication.
var authCacheState = struct {
	sync.Mutex
	version uint64
	flushed uint64
	users   map[string]uint64
}{users: make(map[string]uint64)}

// CurrentAuthCacheGeneration captures the generation before loading authentication data.
func CurrentAuthCacheGeneration() AuthCacheGeneration {
	authCacheState.Lock()
	defer authCacheState.Unlock()

	return AuthCacheGeneration{version: authCacheState.version}
}

// valid reports whether the account has remained unchanged while authCacheState is locked.
func (g AuthCacheGeneration) valid(userUID string) bool {
	return g.version >= authCacheState.flushed && g.version >= authCacheState.users[userUID]
}
