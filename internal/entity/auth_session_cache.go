package entity

import (
	"fmt"
	"time"

	gc "github.com/patrickmn/go-cache"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// Create a new session cache with an expiration time of 15 minutes.
var sessionCacheExpiration = 15 * time.Minute
var sessionCache = gc.New(sessionCacheExpiration, 5*time.Minute)

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
			cached.LastActive = UnixTime()
			return cached, nil
		} else if err := cached.Delete(); err != nil {
			event.AuditErr([]string{cached.IP(), "session %s", "failed to delete after expiration", "%s"}, cached.RefID, err)
		}
	} else if res := Db().First(&found, "id = ?", id); res.RecordNotFound() {
		return found, fmt.Errorf("invalid session")
	} else if res.Error != nil {
		return found, res.Error
	} else if !rnd.IsSessionID(found.ID) {
		return found, fmt.Errorf("invalid session id %s", clean.LogQuote(found.ID))
	} else if !found.Expired() {
		found.UpdateLastActive()
		CacheSession(found, sessionCacheExpiration)
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
		d = sessionCacheExpiration
	}

	if s.PreviewToken != "" {
		PreviewToken.Set(s.PreviewToken, s.ID)
	}

	if s.DownloadToken != "" {
		DownloadToken.Set(s.DownloadToken, s.ID)
	}

	sessionCache.Set(s.ID, s, d)
}

// DeleteSession permanently deletes a session.
func DeleteSession(s *Session) error {
	if s == nil {
		return nil
	} else if !rnd.IsSessionID(s.ID) {
		return fmt.Errorf("invalid session id")
	}

	DeleteFromSessionCache(s.ID)

	if s.PreviewToken != "" {
		PreviewToken.Set(s.PreviewToken, s.ID)
	}

	if s.DownloadToken != "" {
		DownloadToken.Set(s.DownloadToken, s.ID)
	}

	return UnscopedDb().Delete(s).Error
}

// DeleteFromSessionCache deletes a session from the cache.
func DeleteFromSessionCache(id string) {
	if id == "" {
		return
	}

	sessionCache.Delete(id)
}
