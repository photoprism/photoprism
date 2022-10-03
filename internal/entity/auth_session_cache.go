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
func FindSession(id string) (s Session, err error) {
	s = Session{}

	// Valid id?
	if !rnd.IsSessionID(id) {
		return s, fmt.Errorf("id %s is invalid", clean.LogQuote(id))
	}

	// Find cached session.
	if cacheData, ok := sessionCache.Get(id); ok {
		s = cacheData.(Session)
		s.LastActive = UnixTime()
		return s, nil
	}

	// Search database and return session if found.
	if r := Db().First(&s, "id = ?", id); r.RecordNotFound() {
		return s, fmt.Errorf("not found")
	} else if r.Error != nil {
		return s, r.Error
	} else if !rnd.IsSessionID(s.ID) {
		return s, fmt.Errorf("has invalid id %s", clean.LogQuote(s.ID))
	} else if s.Expired() {
		if err = s.Delete(); err != nil {
			event.AuditErr([]string{s.IP(), "session %s", "failed to delete after expiration", "%s"}, s.RefID, err)
		}
		return s, fmt.Errorf("expired")
	} else {
		s.UpdateLastActive()
		sessionCache.SetDefault(s.ID, s)
	}

	return s, err
}
