package entity

import (
	"errors"
	"fmt"
	"time"

	gc "github.com/patrickmn/go-cache"
	"gorm.io/gorm"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/clean"
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
			event.AuditErr([]string{cached.IP(), "session %s", "failed to delete after expiration", "%s"}, cached.RefID, err)
		}
	} else if res := Db().First(&found, "id = ?", id); errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return found, fmt.Errorf("invalid session")
	} else if res.Error != nil {
		return found, res.Error
	} else if !rnd.IsSessionID(found.ID) {
		return found, fmt.Errorf("invalid session id %s", clean.LogQuote(found.ID))
	} else if !found.Expired() {
		// Set session activity timestamp and update the last_active column in the sessions table.
		found.UpdateLastActive(true)
		CacheSession(found, SessionCacheDuration)
		return found, nil
	} else if err := found.Delete(); err != nil {
		event.AuditErr([]string{found.IP(), "session %s", "failed to delete after expiration", "%s"}, found.RefID, err)
	}

	return found, fmt.Errorf("session expired")
}

// FlushSessionCache resets the session cache.
func FlushSessionCache() {
	sessionCache.Flush()
}

// CacheSession adds a session to the cache if its ID is valid.
func CacheSession(s *Session, d time.Duration) {
	if s == nil {
		return
	} else if !rnd.IsSessionID(s.ID) {
		return
	}

	if d == 0 {
		d = SessionCacheDuration
	}

	if s.PreviewToken != "" {
		PreviewToken.Set(s.PreviewToken, s.ID)
	}

	if s.DownloadToken != "" {
		DownloadToken.Set(s.DownloadToken, s.ID)
	}

	sessionCache.Set(s.ID, s, d)
}
