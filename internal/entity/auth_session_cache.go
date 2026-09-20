package entity

import (
	"fmt"
	"time"

	gc "github.com/patrickmn/go-cache"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/log/status"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// SessionCacheDuration specifies how long sessions are cached.
var SessionCacheDuration = 15 * time.Minute
var sessionCache = gc.New(SessionCacheDuration, time.Minute)

// FindSessionByAuthToken finds a session based on the auth token string or returns nil if it does not exist.
func FindSessionByAuthToken(token string) (*Session, error) {
	return FindSession(rnd.SessionID(token))
}

// FindSession finds a session based on the id string or returns nil if it does not exist.
func FindSession(id string) (*Session, error) {
	generation := CurrentAuthCacheGeneration()
	found := &Session{}

	if !rnd.IsSessionID(id) {
		return found, fmt.Errorf("invalid session id")
	}

	// Find the session in the cache with a fallback to the database.
	if cacheData, ok := sessionCache.Get(id); ok && cacheData != nil {
		if cached := cacheData.(*Session); !cached.Expired() {
			// Set session activity timestamp, also update the last_active column in the sessions table if it is new.
			cached.UpdateLastActive(cached.LastActive <= 0)
			return cached, nil
		} else if err := cached.Delete(); err != nil {
			event.AuditErr([]string{cached.IP(), "session %s", "failed to delete after expiration", status.Error(err)}, cached.RefID)
		}
	} else if res := Db().First(&found, "id = ?", id); res.RecordNotFound() {
		return found, fmt.Errorf("invalid session")
	} else if res.Error != nil {
		return found, res.Error
	} else if !rnd.IsSessionID(found.ID) {
		return found, fmt.Errorf("invalid session id %s", clean.LogQuote(found.ID))
	} else if !found.Expired() {
		found.cacheGeneration = &generation
		// Set session activity timestamp and update the last_active column in the sessions table.
		found.UpdateLastActive(true)
		CacheSession(found, SessionCacheDuration)
		return found, nil
	} else if err := found.Delete(); err != nil {
		event.AuditErr([]string{found.IP(), "session %s", "failed to delete after expiration", status.Error(err)}, found.RefID)
	}

	return found, fmt.Errorf("session expired")
}

// FlushSessionCache resets session and WebDAV authentication caches.
func FlushSessionCache() {
	authCacheState.Lock()
	defer authCacheState.Unlock()

	authCacheState.version++
	authCacheState.flushed = authCacheState.version
	clear(authCacheState.users)
	sessionCache.Flush()
	webDAVUserCache.Flush()
}

// FlushUserSessionCache evicts cached sessions and WebDAV credentials for one user.
// Persisted credentials and other users' cache entries are left unchanged.
func FlushUserSessionCache(userUID string) {
	if userUID == "" {
		return
	}

	authCacheState.Lock()
	defer authCacheState.Unlock()

	authCacheState.version++
	authCacheState.users[userUID] = authCacheState.version

	for key, item := range sessionCache.Items() {
		if sess, ok := item.Object.(*Session); ok && sess != nil && sess.UserUID == userUID {
			sessionCache.Delete(key)
		}
	}

	for key, item := range webDAVUserCache.Items() {
		if user, ok := item.Object.(*User); ok && user != nil && user.UserUID == userUID {
			webDAVUserCache.Delete(key)
		}
	}
}

// CacheSession adds a session to the cache if its ID is valid.
func CacheSession(s *Session, d time.Duration) {
	if s == nil {
		return
	} else if !rnd.IsSessionID(s.ID) {
		return
	}

	authCacheState.Lock()
	defer authCacheState.Unlock()

	if s.cacheGeneration == nil {
		// Unbound database records may be cached before their user is resolved.
		if s.user != nil {
			return
		}
		s.cacheGeneration = &AuthCacheGeneration{version: authCacheState.version}
	}
	if !s.cacheGeneration.valid(s.UserUID) {
		return
	}

	if d == 0 {
		d = SessionCacheDuration
	}

	if s.PreviewToken != "" {
		PreviewToken.Set(s.ID, s.PreviewToken)
	}

	sessionCache.Set(s.ID, s, d)
}
