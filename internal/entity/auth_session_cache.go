package entity

import (
	"fmt"
	"time"

	gc "github.com/patrickmn/go-cache"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/rnd"
)

var sessionCacheExpiration = 15 * time.Minute
var sessionCache = gc.New(sessionCacheExpiration, 5*time.Minute)

// FlushSessionCache resets the session cache.
func FlushSessionCache() {
	sessionCache.Flush()
}

// DeleteFromSessionCache deletes a cached session.
func DeleteFromSessionCache(id string) {
	if id == "" {
		return
	}

	sessionCache.Delete(id)
}

// FindSession returns an existing session or nil if not found.
func FindSession(id string) (*Session, error) {
	found := &Session{}

	// Valid id?
	if !rnd.IsSessionID(id) {
		return found, fmt.Errorf("id %s is invalid", clean.LogQuote(id))
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
		return found, fmt.Errorf("not found")
	} else if res.Error != nil {
		return found, res.Error
	} else if !rnd.IsSessionID(found.ID) {
		return found, fmt.Errorf("has invalid id %s", clean.LogQuote(found.ID))
	} else if !found.Expired() {
		found.UpdateLastActive()
		sessionCache.SetDefault(found.ID, found)
		return found, nil
	} else if err := found.Delete(); err != nil {
		event.AuditErr([]string{found.IP(), "session %s", "failed to delete after expiration", "%s"}, found.RefID, err)
	}

	return found, fmt.Errorf("expired")
}
